package httphandler

import (
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/decimals"
	"github.com/gofiber/fiber/v2"
	"github.com/holiman/uint256"
	"github.com/samber/lo"
)

type getUTXOsByAddressRequest struct {
	Wallet      string `params:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r getUTXOsByAddressRequest) Validate() error {
	var errList []error
	if r.Wallet == "" {
		errList = append(errList, errors.New("'wallet' is required"))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type transferableInscription struct {
	Ticker   string       `json:"ticker"`
	Amount   *uint256.Int `json:"amount"`
	Decimals uint16       `json:"decimals"`
}

type utxoExtend struct {
	TransferableInscriptions []transferableInscription `json:"transferableInscriptions"`
}

type utxo struct {
	TxHash      chainhash.Hash `json:"txHash"`
	OutputIndex uint32         `json:"outputIndex"`
	Extend      utxoExtend     `json:"extend"`
}

type getUTXOsByAddressResult struct {
	List        []utxo `json:"list"`
	BlockHeight uint64 `json:"blockHeight"`
}

type getUTXOsByAddressResponse = common.HttpResponse[getUTXOsByAddressResult]

func (h *HttpHandler) GetUTXOsByAddress(ctx *fiber.Ctx) (err error) {
	var req getUTXOsByAddressRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	pkScript, err := btcutils.ToPkScript(h.network, req.Wallet)
	if err != nil {
		return errs.NewPublicError("unable to resolve pkscript from \"wallet\"")
	}

	blockHeight := req.BlockHeight
	if blockHeight == 0 {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	transferables, err := h.usecase.GetTransferableTransfersByPkScript(ctx.UserContext(), pkScript, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetTransferableTransfersByPkScript")
	}

	transferableTicks := lo.Map(transferables, func(src *entity.EventInscribeTransfer, _ int) string { return src.Tick })
	entries, err := h.usecase.GetTickEntryByTickBatch(ctx.UserContext(), transferableTicks)
	if err != nil {
		return errors.Wrap(err, "error during GetTickEntryByTickBatch")
	}

	groupedtransferableTi := lo.GroupBy(transferables, func(src *entity.EventInscribeTransfer) wire.OutPoint { return src.SatPoint.OutPoint })
	utxoList := make([]utxo, 0, len(groupedtransferableTi))
	for outPoint, transferables := range groupedtransferableTi {
		transferableInscriptions := make([]transferableInscription, 0, len(transferables))
		for _, transferable := range transferables {
			entry := entries[transferable.Tick]
			transferableInscriptions = append(transferableInscriptions, transferableInscription{
				Ticker:   transferable.Tick,
				Amount:   decimals.ToUint256(transferable.Amount, entry.Decimals),
				Decimals: entry.Decimals,
			})
		}

		utxoList = append(utxoList, utxo{
			TxHash:      outPoint.Hash,
			OutputIndex: outPoint.Index,
			Extend: utxoExtend{
				TransferableInscriptions: transferableInscriptions,
			},
		})
	}

	// filter by req.Id if exists
	{
		utxoList = lo.Filter(utxoList, func(u utxo, _ int) bool {
			for _, transferableInscriptions := range u.Extend.TransferableInscriptions {
				if ok := strings.EqualFold(req.Id, transferableInscriptions.Ticker); ok {
					return ok
				}
			}
			return false
		})
	}

	resp := getUTXOsByAddressResponse{
		Result: &getUTXOsByAddressResult{
			BlockHeight: blockHeight,
			List:        utxoList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
