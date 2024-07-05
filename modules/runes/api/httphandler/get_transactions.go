package httphandler

import (
	"encoding/hex"
	"fmt"
	"slices"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getTransactionsRequest struct {
	Wallet    string `query:"wallet"`
	Id        string `query:"id"`
	FromBlock int64  `query:"fromBlock"`
	ToBlock   int64  `query:"toBlock"`
	Limit     int32  `query:"limit"`
	Offset    int32  `query:"offset"`
}

const getTransactionsMaxLimit = 3000

func (r getTransactionsRequest) Validate() error {
	var errList []error
	if r.Id != "" && !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.New("'id' is not valid rune id or rune name"))
	}
	if r.FromBlock < -1 {
		errList = append(errList, errors.Errorf("invalid fromBlock range"))
	}
	if r.ToBlock < -1 {
		errList = append(errList, errors.Errorf("invalid toBlock range"))
	}
	if r.Limit < 0 {
		errList = append(errList, errors.New("'limit' must be non-negative"))
	}
	if r.Limit > getTransactionsMaxLimit {
		errList = append(errList, errors.Errorf("'limit' cannot exceed %d", getTransactionsMaxLimit))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type txInputOutput struct {
	PkScript string          `json:"pkScript"`
	Address  string          `json:"address"`
	Id       runes.RuneId    `json:"id"`
	Amount   uint128.Uint128 `json:"amount"`
	Decimals uint8           `json:"decimals"`
	Index    uint32          `json:"index"`
}

type terms struct {
	Amount      *uint128.Uint128 `json:"amount"`
	Cap         *uint128.Uint128 `json:"cap"`
	HeightStart *uint64          `json:"heightStart"`
	HeightEnd   *uint64          `json:"heightEnd"`
	OffsetStart *uint64          `json:"offsetStart"`
	OffsetEnd   *uint64          `json:"offsetEnd"`
}

type etching struct {
	Divisibility *uint8           `json:"divisibility"`
	Premine      *uint128.Uint128 `json:"premine"`
	Rune         *runes.Rune      `json:"rune"`
	Spacers      *uint32          `json:"spacers"`
	Symbol       *string          `json:"symbol"`
	Terms        *terms           `json:"terms"`
	Turbo        bool             `json:"turbo"`
}

type edict struct {
	Id     runes.RuneId    `json:"id"`
	Amount uint128.Uint128 `json:"amount"`
	Output int             `json:"output"`
}

type runestone struct {
	Cenotaph bool          `json:"cenotaph"`
	Flaws    []string      `json:"flaws"`
	Etching  *etching      `json:"etching"`
	Edicts   []edict       `json:"edicts"`
	Mint     *runes.RuneId `json:"mint"`
	Pointer  *uint64       `json:"pointer"`
}

type runeTransactionExtend struct {
	RuneEtched bool       `json:"runeEtched"`
	Runestone  *runestone `json:"runestone"`
}

type amountWithDecimal struct {
	Amount   uint128.Uint128 `json:"amount"`
	Decimals uint8           `json:"decimals"`
}

type transaction struct {
	TxHash      chainhash.Hash               `json:"txHash"`
	BlockHeight uint64                       `json:"blockHeight"`
	Index       uint32                       `json:"index"`
	Timestamp   int64                        `json:"timestamp"`
	Inputs      []txInputOutput              `json:"inputs"`
	Outputs     []txInputOutput              `json:"outputs"`
	Mints       map[string]amountWithDecimal `json:"mints"`
	Burns       map[string]amountWithDecimal `json:"burns"`
	Extend      runeTransactionExtend        `json:"extend"`
}

type getTransactionsResult struct {
	List []transaction `json:"list"`
}

type getTransactionsResponse = HttpResponse[getTransactionsResult]

func (h *HttpHandler) GetTransactions(ctx *fiber.Ctx) (err error) {
	var req getTransactionsRequest
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	var pkScript []byte
	if req.Wallet != "" {
		var ok bool
		pkScript, ok = resolvePkScript(h.network, req.Wallet)
		if !ok {
			return errs.NewPublicError("unable to resolve pkscript from \"wallet\"")
		}
	}

	var runeId runes.RuneId
	if req.Id != "" {
		var ok bool
		runeId, ok = h.resolveRuneId(ctx.UserContext(), req.Id)
		if !ok {
			return errs.NewPublicError("unable to resolve rune id from \"id\"")
		}
	}
	if req.Limit == 0 {
		req.Limit = getBalancesByAddressMaxLimit
	}

	// default to latest block
	if req.ToBlock == 0 {
		req.ToBlock = -1
	}

	// get latest block height if block height is -1
	if req.FromBlock == -1 || req.ToBlock == -1 {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		if req.FromBlock == -1 {
			req.FromBlock = blockHeader.Height
		}
		if req.ToBlock == -1 {
			req.ToBlock = blockHeader.Height
		}
	}

	// validate block height range
	if req.FromBlock > req.ToBlock {
		return errs.NewPublicError(fmt.Sprintf("fromBlock must be less than or equal to toBlock, got fromBlock=%d, toBlock=%d", req.FromBlock, req.ToBlock))
	}

	txs, err := h.usecase.GetRuneTransactions(ctx.UserContext(), pkScript, runeId, uint64(req.FromBlock), uint64(req.ToBlock), req.Limit, req.Offset)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneTransactions")
	}

	{
		txHashes := lo.Map(txs, func(tx *entity.RuneTransaction, _ int) chainhash.Hash {
			return tx.Hash
		})
		logger.Debug("txHashes", slogx.Any("txHashes", txHashes))
	}

	var allRuneIds []runes.RuneId
	for _, tx := range txs {
		for id := range tx.Mints {
			allRuneIds = append(allRuneIds, id)
		}
		for id := range tx.Burns {
			allRuneIds = append(allRuneIds, id)
		}
		for _, input := range tx.Inputs {
			allRuneIds = append(allRuneIds, input.RuneId)
		}
		for _, output := range tx.Outputs {
			allRuneIds = append(allRuneIds, output.RuneId)
		}
	}
	allRuneIds = lo.Uniq(allRuneIds)
	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), allRuneIds)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

	txList := make([]transaction, 0, len(txs))
	for _, tx := range txs {
		respTx := transaction{
			TxHash:      tx.Hash,
			BlockHeight: tx.BlockHeight,
			Index:       tx.Index,
			Timestamp:   tx.Timestamp.Unix(),
			Inputs:      make([]txInputOutput, 0, len(tx.Inputs)),
			Outputs:     make([]txInputOutput, 0, len(tx.Outputs)),
			Mints:       make(map[string]amountWithDecimal, len(tx.Mints)),
			Burns:       make(map[string]amountWithDecimal, len(tx.Burns)),
			Extend: runeTransactionExtend{
				RuneEtched: tx.RuneEtched,
				Runestone:  nil,
			},
		}
		for _, input := range tx.Inputs {
			address := addressFromPkScript(input.PkScript, h.network)
			respTx.Inputs = append(respTx.Inputs, txInputOutput{
				PkScript: hex.EncodeToString(input.PkScript),
				Address:  address,
				Id:       input.RuneId,
				Amount:   input.Amount,
				Decimals: runeEntries[input.RuneId].Divisibility,
				Index:    input.Index,
			})
		}
		for _, output := range tx.Outputs {
			address := addressFromPkScript(output.PkScript, h.network)
			respTx.Outputs = append(respTx.Outputs, txInputOutput{
				PkScript: hex.EncodeToString(output.PkScript),
				Address:  address,
				Id:       output.RuneId,
				Amount:   output.Amount,
				Decimals: runeEntries[output.RuneId].Divisibility,
				Index:    output.Index,
			})
		}
		for id, amount := range tx.Mints {
			respTx.Mints[id.String()] = amountWithDecimal{
				Amount:   amount,
				Decimals: runeEntries[id].Divisibility,
			}
		}
		for id, amount := range tx.Burns {
			respTx.Burns[id.String()] = amountWithDecimal{
				Amount:   amount,
				Decimals: runeEntries[id].Divisibility,
			}
		}
		if tx.Runestone != nil {
			var e *etching
			if tx.Runestone.Etching != nil {
				var symbol *string
				if tx.Runestone.Etching.Symbol != nil {
					symbol = lo.ToPtr(string(*tx.Runestone.Etching.Symbol))
				}
				var t *terms
				if tx.Runestone.Etching.Terms != nil {
					t = &terms{
						Amount:      tx.Runestone.Etching.Terms.Amount,
						Cap:         tx.Runestone.Etching.Terms.Cap,
						HeightStart: tx.Runestone.Etching.Terms.HeightStart,
						HeightEnd:   tx.Runestone.Etching.Terms.HeightEnd,
						OffsetStart: tx.Runestone.Etching.Terms.OffsetStart,
						OffsetEnd:   tx.Runestone.Etching.Terms.OffsetEnd,
					}
				}
				e = &etching{
					Divisibility: tx.Runestone.Etching.Divisibility,
					Premine:      tx.Runestone.Etching.Premine,
					Rune:         tx.Runestone.Etching.Rune,
					Spacers:      tx.Runestone.Etching.Spacers,
					Symbol:       symbol,
					Terms:        t,
					Turbo:        tx.Runestone.Etching.Turbo,
				}
			}
			respTx.Extend.Runestone = &runestone{
				Cenotaph: tx.Runestone.Cenotaph,
				Flaws:    lo.Ternary(tx.Runestone.Cenotaph, tx.Runestone.Flaws.CollectAsString(), nil),
				Etching:  e,
				Edicts: lo.Map(tx.Runestone.Edicts, func(ed runes.Edict, _ int) edict {
					return edict{
						Id:     ed.Id,
						Amount: ed.Amount,
						Output: ed.Output,
					}
				}),
				Mint:    tx.Runestone.Mint,
				Pointer: tx.Runestone.Pointer,
			}
		}
		txList = append(txList, respTx)
	}
	// sort by block height DESC, then index DESC
	slices.SortFunc(txList, func(t1, t2 transaction) int {
		if t1.BlockHeight != t2.BlockHeight {
			return int(t2.BlockHeight - t1.BlockHeight)
		}
		return int(t2.Index - t1.Index)
	})

	resp := getTransactionsResponse{
		Result: &getTransactionsResult{
			List: txList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
