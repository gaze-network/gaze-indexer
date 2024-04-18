package httphandler

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getTransactionsRequest struct {
	Wallet      string `query:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r getTransactionsRequest) Validate() error {
	var errList []error
	if r.Id != "" && !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.New("'id' is not valid rune id or rune name"))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type outPointBalance struct {
	PkScript string          `json:"pkScript"`
	Address  string          `json:"address"`
	Id       runes.RuneId    `json:"id"`
	Amount   uint128.Uint128 `json:"amount"`
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

type transaction struct {
	TxHash      chainhash.Hash             `json:"txHash"`
	BlockHeight uint64                     `json:"blockHeight"`
	Timestamp   int64                      `json:"timestamp"`
	Inputs      []outPointBalance          `json:"inputs"`
	Outputs     []outPointBalance          `json:"outputs"`
	Mints       map[string]uint128.Uint128 `json:"mints"`
	Burns       map[string]uint128.Uint128 `json:"burns"`
	Extend      runeTransactionExtend      `json:"extend"`
}

type getTransactionsResult struct {
	List []transaction `json:"list"`
}

type getTransactionsResponse = HttpResponse[getTransactionsResult]

func (h *HttpHandler) GetTransactions(ctx *fiber.Ctx) (err error) {
	var req getTransactionsRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
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

	blockHeight := req.BlockHeight
	if blockHeight == 0 {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	var runeId runes.RuneId
	if req.Id != "" {
		var ok bool
		runeId, ok = h.resolveRuneId(ctx.UserContext(), req.Id)
		if !ok {
			return errs.NewPublicError("unable to resolve rune id from \"id\"")
		}
	}

	txs, err := h.usecase.GetTransactionsByHeight(ctx.UserContext(), blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetTransactionsByHeight")
	}

	filteredTxs := make([]*entity.RuneTransaction, 0)
	isTxContainPkScript := func(tx *entity.RuneTransaction) bool {
		for _, input := range tx.Inputs {
			if bytes.Equal(input.PkScript, pkScript) {
				return true
			}
		}
		for _, output := range tx.Outputs {
			if bytes.Equal(output.PkScript, pkScript) {
				return true
			}
		}
		return false
	}
	isTxContainRuneId := func(tx *entity.RuneTransaction) bool {
		for _, input := range tx.Inputs {
			if input.RuneId == runeId {
				return true
			}
		}
		for _, output := range tx.Outputs {
			if output.RuneId == runeId {
				return true
			}
		}
		for mintedRuneId := range tx.Mints {
			if mintedRuneId == runeId {
				return true
			}
		}
		for burnedRuneId := range tx.Burns {
			if burnedRuneId == runeId {
				return true
			}
		}
		if tx.Runestone != nil {
			if tx.Runestone.Mint != nil && *tx.Runestone.Mint == runeId {
				return true
			}
			// returns true if this tx etched this runeId
			if tx.RuneEtched && tx.BlockHeight == runeId.BlockHeight && tx.Index == runeId.TxIndex {
				return true
			}
		}
		return false
	}
	for _, tx := range txs {
		if pkScript != nil && !isTxContainPkScript(tx) {
			continue
		}
		if runeId != (runes.RuneId{}) && isTxContainRuneId(tx) {
			continue
		}
		filteredTxs = append(filteredTxs, tx)
	}

	txList := make([]transaction, 0, len(filteredTxs))
	for _, tx := range filteredTxs {
		respTx := transaction{
			TxHash:      tx.Hash,
			BlockHeight: tx.BlockHeight,
			Timestamp:   tx.Timestamp.Unix(),
			Inputs:      make([]outPointBalance, 0, len(tx.Inputs)),
			Outputs:     make([]outPointBalance, 0, len(tx.Outputs)),
			Mints:       make(map[string]uint128.Uint128, len(tx.Mints)),
			Burns:       make(map[string]uint128.Uint128, len(tx.Burns)),
			Extend: runeTransactionExtend{
				RuneEtched: tx.RuneEtched,
				Runestone:  nil,
			},
		}
		for _, input := range tx.Inputs {
			address := addressFromPkScript(input.PkScript, h.network)
			respTx.Inputs = append(respTx.Inputs, outPointBalance{
				PkScript: hex.EncodeToString(input.PkScript),
				Address:  address,
				Id:       input.RuneId,
				Amount:   input.Amount,
				Index:    input.Index,
			})
		}
		for _, output := range tx.Outputs {
			address := addressFromPkScript(output.PkScript, h.network)
			respTx.Outputs = append(respTx.Outputs, outPointBalance{
				PkScript: hex.EncodeToString(output.PkScript),
				Address:  address,
				Id:       output.RuneId,
				Amount:   output.Amount,
				Index:    output.Index,
			})
		}
		for id, amount := range tx.Mints {
			respTx.Mints[id.String()] = amount
		}
		for id, amount := range tx.Burns {
			respTx.Burns[id.String()] = amount
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

	resp := getTransactionsResponse{
		Result: &getTransactionsResult{
			List: txList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
