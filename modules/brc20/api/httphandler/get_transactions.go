package httphandler

import (
	"bytes"
	"cmp"
	"encoding/hex"
	"slices"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/decimals"
	"github.com/gofiber/fiber/v2"
	"github.com/holiman/uint256"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

var ops = []string{"inscribe-deploy", "inscribe-mint", "inscribe-transfer", "transfer-transfer"}

type getTransactionsRequest struct {
	Wallet      string `query:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
	Op          string `query:"op"`
}

func (r getTransactionsRequest) Validate() error {
	var errList []error
	if r.Op != "" {
		if !lo.Contains(ops, r.Op) {
			errList = append(errList, errors.Errorf("invalid 'op' value: %s, supported values: %s", r.Op, strings.Join(ops, ", ")))
		}
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type txOpDeployArg struct {
	Op       string          `json:"op"`
	Tick     string          `json:"tick"`
	Max      decimal.Decimal `json:"max"`
	Lim      decimal.Decimal `json:"lim"`
	Dec      uint16          `json:"dec"`
	SelfMint bool            `json:"self_mint"`
}

type txOpGeneralArg struct {
	Op     string          `json:"op"`
	Tick   string          `json:"tick"`
	Amount decimal.Decimal `json:"amt"`
}

type txOperation[T any] struct {
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber int64  `json:"inscriptionNumber"`
	Op                string `json:"op"`
	Args              T      `json:"args"`
}

type txOperationsDeploy struct {
	txOperation[txOpDeployArg]
	Address string `json:"address"`
}

type txOperationsMint struct {
	txOperation[txOpGeneralArg]
	Address string `json:"address"`
}

type txOperationsInscribeTransfer struct {
	txOperation[txOpGeneralArg]
	Address     string `json:"address"`
	OutputIndex uint32 `json:"outputIndex"`
	Sats        uint64 `json:"sats"`
}

type txOperationsTransferTransfer struct {
	txOperation[txOpGeneralArg]
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
}

type transactionExtend struct {
	Operations []any `json:"operations"`
}

type amountWithDecimal struct {
	Amount   *uint256.Int `json:"amount"`
	Decimals uint16       `json:"decimals"`
}

type txInputOutput struct {
	PkScript string       `json:"pkScript"`
	Address  string       `json:"address"`
	Id       string       `json:"id"`
	Amount   *uint256.Int `json:"amount"`
	Decimals uint16       `json:"decimals"`
	Index    uint32       `json:"index"`
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
	Extend      transactionExtend            `json:"extend"`
}

type getTransactionsResult struct {
	List []transaction `json:"list"`
}

type getTransactionsResponse = common.HttpResponse[getTransactionsResult]

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
		pkScript, err = btcutils.ToPkScript(h.network, req.Wallet)
		if err != nil {
			return errs.NewPublicError("unable to resolve pkscript from \"wallet\"")
		}
	}

	blockHeight := req.BlockHeight
	// set blockHeight to the latest block height blockHeight, pkScript, and runeId are not provided
	if blockHeight == 0 && pkScript == nil && req.Id == "" {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	var (
		deployEvents           []*entity.EventDeploy
		mintEvents             []*entity.EventMint
		transferTransferEvents []*entity.EventTransferTransfer
		inscribeTransferEvents []*entity.EventInscribeTransfer
	)

	group, groupctx := errgroup.WithContext(ctx.UserContext())

	if req.Op == "" || req.Op == "inscribe-deploy" {
		group.Go(func() error {
			events, err := h.usecase.GetDeployEvents(groupctx, pkScript, req.Id, blockHeight)
			deployEvents = events
			return errors.Wrap(err, "error during get inscribe-deploy events")
		})
	}
	if req.Op == "" || req.Op == "inscribe-mint" {
		group.Go(func() error {
			events, err := h.usecase.GetMintEvents(groupctx, pkScript, req.Id, blockHeight)
			mintEvents = events
			return errors.Wrap(err, "error during get inscribe-mint events")
		})
	}
	if req.Op == "" || req.Op == "transfer-transfer" {
		group.Go(func() error {
			events, err := h.usecase.GetTransferTransferEvents(groupctx, pkScript, req.Id, blockHeight)
			transferTransferEvents = events
			return errors.Wrap(err, "error during get transfer-transfer events")
		})
	}
	if req.Op == "" || req.Op == "inscribe-transfer" {
		group.Go(func() error {
			events, err := h.usecase.GetInscribeTransferEvents(groupctx, pkScript, req.Id, blockHeight)
			inscribeTransferEvents = events
			return errors.Wrap(err, "error during get inscribe-transfer events")
		})
	}
	if err := group.Wait(); err != nil {
		return errors.WithStack(err)
	}

	allTicks := make([]string, 0, len(deployEvents)+len(mintEvents)+len(transferTransferEvents)+len(inscribeTransferEvents))
	allTicks = append(allTicks, lo.Map(deployEvents, func(event *entity.EventDeploy, _ int) string { return event.Tick })...)
	allTicks = append(allTicks, lo.Map(mintEvents, func(event *entity.EventMint, _ int) string { return event.Tick })...)
	allTicks = append(allTicks, lo.Map(transferTransferEvents, func(event *entity.EventTransferTransfer, _ int) string { return event.Tick })...)
	allTicks = append(allTicks, lo.Map(inscribeTransferEvents, func(event *entity.EventInscribeTransfer, _ int) string { return event.Tick })...)
	entries, err := h.usecase.GetTickEntryByTickBatch(ctx.UserContext(), lo.Uniq(allTicks))
	if err != nil {
		return errors.Wrap(err, "error during GetTickEntryByTickBatch")
	}

	rawTxList := make([]transaction, 0, len(deployEvents)+len(mintEvents)+len(transferTransferEvents)+len(inscribeTransferEvents))

	// Deploy events
	for _, event := range deployEvents {
		address, err := btcutils.PkScriptToAddress(event.PkScript, h.network)
		if err != nil {
			return errors.Wrapf(err, `error during PkScriptToAddress for deploy event %s, pkscript: %x, network: %v`, event.TxHash, event.PkScript, h.network)
		}
		respTx := transaction{
			TxHash:      event.TxHash,
			BlockHeight: event.BlockHeight,
			Index:       event.TxIndex,
			Timestamp:   event.Timestamp.Unix(),
			Mints:       map[string]amountWithDecimal{},
			Burns:       map[string]amountWithDecimal{},
			Extend: transactionExtend{
				Operations: []any{
					txOperationsDeploy{
						txOperation: txOperation[txOpDeployArg]{
							InscriptionId:     event.InscriptionId.String(),
							InscriptionNumber: event.InscriptionNumber,
							Op:                "deploy",
							Args: txOpDeployArg{
								Op:       "deploy",
								Tick:     event.Tick,
								Max:      event.TotalSupply,
								Lim:      event.LimitPerMint,
								Dec:      event.Decimals,
								SelfMint: event.IsSelfMint,
							},
						},
						Address: address,
					},
				},
			},
		}
		rawTxList = append(rawTxList, respTx)
	}

	// Mint events
	for _, event := range mintEvents {
		entry := entries[event.Tick]
		address, err := btcutils.PkScriptToAddress(event.PkScript, h.network)
		if err != nil {
			return errors.Wrapf(err, `error during PkScriptToAddress for deploy event %s, pkscript: %x, network: %v`, event.TxHash, event.PkScript, h.network)
		}
		amtWei := decimals.ToUint256(event.Amount, entry.Decimals)
		respTx := transaction{
			TxHash:      event.TxHash,
			BlockHeight: event.BlockHeight,
			Index:       event.TxIndex,
			Timestamp:   event.Timestamp.Unix(),
			Outputs: []txInputOutput{
				{
					PkScript: hex.EncodeToString(event.PkScript),
					Address:  address,
					Id:       event.Tick,
					Amount:   amtWei,
					Decimals: entry.Decimals,
					Index:    event.TxIndex,
				},
			},
			Mints: map[string]amountWithDecimal{
				event.Tick: {
					Amount:   amtWei,
					Decimals: entry.Decimals,
				},
			},
			Extend: transactionExtend{
				Operations: []any{
					txOperationsMint{
						txOperation: txOperation[txOpGeneralArg]{
							InscriptionId:     event.InscriptionId.String(),
							InscriptionNumber: event.InscriptionNumber,
							Op:                "inscribe-mint",
							Args: txOpGeneralArg{
								Op:     "inscribe-mint",
								Tick:   event.Tick,
								Amount: event.Amount,
							},
						},
						Address: address,
					},
				},
			},
		}
		rawTxList = append(rawTxList, respTx)
	}

	// Inscribe Transfer events
	for _, event := range inscribeTransferEvents {
		address, err := btcutils.PkScriptToAddress(event.PkScript, h.network)
		if err != nil {
			return errors.Wrapf(err, `error during PkScriptToAddress for deploy event %s, pkscript: %x, network: %v`, event.TxHash, event.PkScript, h.network)
		}
		respTx := transaction{
			TxHash:      event.TxHash,
			BlockHeight: event.BlockHeight,
			Index:       event.TxIndex,
			Timestamp:   event.Timestamp.Unix(),
			Mints:       map[string]amountWithDecimal{},
			Burns:       map[string]amountWithDecimal{},
			Extend: transactionExtend{
				Operations: []any{
					txOperationsInscribeTransfer{
						txOperation: txOperation[txOpGeneralArg]{
							InscriptionId:     event.InscriptionId.String(),
							InscriptionNumber: event.InscriptionNumber,
							Op:                "inscribe-transfer",
							Args: txOpGeneralArg{
								Op:     "inscribe-transfer",
								Tick:   event.Tick,
								Amount: event.Amount,
							},
						},
						Address:     address,
						OutputIndex: event.SatPoint.OutPoint.Index,
						Sats:        event.SatsAmount,
					},
				},
			},
		}
		rawTxList = append(rawTxList, respTx)
	}

	// Transfer Transfer events
	for _, event := range transferTransferEvents {
		entry := entries[event.Tick]
		amntWei := decimals.ToUint256(event.Amount, entry.Decimals)
		fromAddress, err := btcutils.PkScriptToAddress(event.FromPkScript, h.network)
		if err != nil {
			return errors.Wrapf(err, `error during PkScriptToAddress for deploy event %s, pkscript: %x, network: %v`, event.TxHash, event.FromPkScript, h.network)
		}
		toAddress := ""
		if len(event.ToPkScript) > 0 && !bytes.Equal(event.ToPkScript, []byte{0x6a}) {
			toAddress, err = btcutils.PkScriptToAddress(event.ToPkScript, h.network)
			if err != nil {
				return errors.Wrapf(err, `error during PkScriptToAddress for deploy event %s, pkscript: %x, network: %v`, event.TxHash, event.FromPkScript, h.network)
			}
		}

		// if toAddress is empty, it's a burn.
		burns := map[string]amountWithDecimal{}
		if len(toAddress) == 0 {
			burns[event.Tick] = amountWithDecimal{
				Amount:   amntWei,
				Decimals: entry.Decimals,
			}
		}

		respTx := transaction{
			TxHash:      event.TxHash,
			BlockHeight: event.BlockHeight,
			Index:       event.TxIndex,
			Timestamp:   event.Timestamp.Unix(),
			Inputs: []txInputOutput{
				{
					PkScript: hex.EncodeToString(event.FromPkScript),
					Address:  fromAddress,
					Id:       event.Tick,
					Amount:   amntWei,
					Decimals: entry.Decimals,
					Index:    event.ToOutputIndex,
				},
			},
			Outputs: []txInputOutput{
				{
					PkScript: hex.EncodeToString(event.ToPkScript),
					Address:  fromAddress,
					Id:       event.Tick,
					Amount:   amntWei,
					Decimals: entry.Decimals,
					Index:    event.ToOutputIndex,
				},
			},
			Mints: map[string]amountWithDecimal{},
			Burns: burns,
			Extend: transactionExtend{
				Operations: []any{
					txOperationsTransferTransfer{
						txOperation: txOperation[txOpGeneralArg]{
							InscriptionId:     event.InscriptionId.String(),
							InscriptionNumber: event.InscriptionNumber,
							Op:                "transfer-transfer",
							Args: txOpGeneralArg{
								Op:     "transfer-transfer",
								Tick:   event.Tick,
								Amount: event.Amount,
							},
						},
						FromAddress: fromAddress,
						ToAddress:   toAddress,
					},
				},
			},
		}
		rawTxList = append(rawTxList, respTx)
	}

	// merge brc-20 tx events that have the same tx hash
	txList := make([]transaction, 0, len(rawTxList))
	groupedTxs := lo.GroupBy(rawTxList, func(tx transaction) chainhash.Hash { return tx.TxHash })
	for _, txs := range groupedTxs {
		tx := txs[0]
		if tx.Mints == nil {
			tx.Mints = map[string]amountWithDecimal{}
		}
		if tx.Burns == nil {
			tx.Burns = map[string]amountWithDecimal{}
		}
		for _, tx2 := range txs[1:] {
			tx.Inputs = append(tx.Inputs, tx2.Inputs...)
			tx.Outputs = append(tx.Outputs, tx2.Outputs...)
			for tick, tx2Ammt := range tx2.Mints {
				// merge the amount if same tick
				// TODO: or it shouldn't happen?
				if txAmmt, ok := tx.Mints[tick]; ok {
					tx.Mints[tick] = amountWithDecimal{
						Amount:   new(uint256.Int).Add(txAmmt.Amount, tx2Ammt.Amount),
						Decimals: txAmmt.Decimals,
					}
				} else {
					tx.Mints[tick] = tx2Ammt
				}
			}
			for tick, tx2Ammt := range tx2.Burns {
				// merge the amount if same tick
				// TODO: or it shouldn't happen?
				if txAmmt, ok := tx.Burns[tick]; ok {
					tx.Burns[tick] = amountWithDecimal{
						Amount:   new(uint256.Int).Add(txAmmt.Amount, tx2Ammt.Amount),
						Decimals: txAmmt.Decimals,
					}
				} else {
					tx.Burns[tick] = tx2Ammt
				}
			}
			tx.Extend.Operations = append(tx.Extend.Operations, tx2.Extend.Operations...)
		}
		slices.SortFunc(tx.Inputs, func(i, j txInputOutput) int {
			return cmp.Compare(i.Index, j.Index)
		})
		slices.SortFunc(tx.Outputs, func(i, j txInputOutput) int {
			return cmp.Compare(i.Index, j.Index)
		})
		txList = append(txList, tx)
	}

	// sort by block height ASC, then index ASC
	slices.SortFunc(txList, func(t1, t2 transaction) int {
		if t1.BlockHeight != t2.BlockHeight {
			return int(t1.BlockHeight - t2.BlockHeight)
		}
		return int(t1.Index - t2.Index)
	})

	resp := getTransactionsResponse{
		Result: &getTransactionsResult{
			List: txList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
