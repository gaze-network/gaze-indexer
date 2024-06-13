package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/decimals"
	"github.com/gofiber/fiber/v2"
	"github.com/holiman/uint256"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type getTokenInfoRequest struct {
	Id          string `params:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r getTokenInfoRequest) Validate() error {
	var errList []error
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type tokenInfoExtend struct {
	DeployedBy              string       `json:"deployedBy"`
	LimitPerMint            *uint256.Int `json:"limitPerMint"`
	DeployInscriptionId     string       `json:"deployInscriptionId"`
	DeployInscriptionNumber int64        `json:"deployInscriptionNumber"`
	InscriptionStartNumber  int64        `json:"inscriptionStartNumber"`
	InscriptionEndNumber    int64        `json:"inscriptionEndNumber"`
}

type getTokenInfoResult struct {
	Id                string          `json:"id"`
	Name              string          `json:"name"`
	Symbol            string          `json:"symbol"`
	TotalSupply       *uint256.Int    `json:"totalSupply"`
	CirculatingSupply *uint256.Int    `json:"circulatingSupply"`
	MintedAmount      *uint256.Int    `json:"mintedAmount"`
	BurnedAmount      *uint256.Int    `json:"burnedAmount"`
	Decimals          uint16          `json:"decimals"`
	DeployedAt        uint64          `json:"deployedAt"`
	DeployedAtHeight  uint64          `json:"deployedAtHeight"`
	CompletedAt       *uint64         `json:"completedAt"`
	CompletedAtHeight *uint64         `json:"completedAtHeight"`
	HoldersCount      int             `json:"holdersCount"`
	Extend            tokenInfoExtend `json:"extend"`
}

type getTokenInfoResponse = common.HttpResponse[getTokenInfoResult]

func (h *HttpHandler) GetTokenInfo(ctx *fiber.Ctx) (err error) {
	var req getTokenInfoRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	blockHeight := req.BlockHeight
	if blockHeight == 0 {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	group, groupctx := errgroup.WithContext(ctx.UserContext())
	var (
		entry                                         *entity.TickEntry
		firstInscriptionNumber, lastInscriptionNumber int64
		deployEvent                                   *entity.EventDeploy
		holdingBalances                               []*entity.Balance
	)
	group.Go(func() error {
		deployEvent, err = h.usecase.GetDeployEventByTick(groupctx, req.Id)
		if err != nil {
			return errors.Wrap(err, "error during GetDeployEventByTick")
		}
		return nil
	})
	group.Go(func() error {
		// TODO: at block height to parameter.
		firstInscriptionNumber, lastInscriptionNumber, err = h.usecase.GetFirstLastInscriptionNumberByTick(groupctx, req.Id)
		if err != nil {
			return errors.Wrap(err, "error during GetFirstLastInscriptionNumberByTick")
		}
		return nil
	})
	group.Go(func() error {
		entry, err = h.usecase.GetTickEntryByTickAndHeight(groupctx, req.Id, blockHeight)
		if err != nil {
			return errors.Wrap(err, "error during GetTickEntryByTickAndHeight")
		}
		return nil
	})
	group.Go(func() error {
		balances, err := h.usecase.GetBalancesByTick(groupctx, req.Id, blockHeight)
		if err != nil {
			return errors.Wrap(err, "error during GetBalancesByRuneId")
		}
		holdingBalances = lo.Filter(balances, func(b *entity.Balance, _ int) bool {
			return !b.OverallBalance.IsZero()
		})
		return nil
	})
	if err := group.Wait(); err != nil {
		return errors.WithStack(err)
	}

	address, err := btcutils.PkScriptToAddress(deployEvent.PkScript, h.network)
	if err != nil {
		return errors.Wrapf(err, `error during PkScriptToAddress for pkscript: %x, network: %v`, deployEvent.PkScript, h.network)
	}

	resp := getTokenInfoResponse{
		Result: &getTokenInfoResult{
			Id:                entry.Tick,
			Name:              entry.OriginalTick,
			Symbol:            entry.Tick,
			TotalSupply:       decimals.ToUint256(entry.TotalSupply, entry.Decimals),
			CirculatingSupply: decimals.ToUint256(entry.MintedAmount.Sub(entry.BurnedAmount), entry.Decimals),
			MintedAmount:      decimals.ToUint256(entry.MintedAmount, entry.Decimals),
			BurnedAmount:      decimals.ToUint256(entry.BurnedAmount, entry.Decimals),
			Decimals:          entry.Decimals,
			DeployedAt:        uint64(entry.DeployedAt.Unix()),
			DeployedAtHeight:  entry.DeployedAtHeight,
			CompletedAt:       lo.Ternary(entry.CompletedAt.IsZero(), nil, lo.ToPtr(uint64(entry.CompletedAt.Unix()))),
			CompletedAtHeight: lo.Ternary(entry.CompletedAtHeight == 0, nil, lo.ToPtr(entry.CompletedAtHeight)),
			HoldersCount:      len(holdingBalances),
			Extend: tokenInfoExtend{
				DeployedBy:              address,
				LimitPerMint:            decimals.ToUint256(entry.LimitPerMint, entry.Decimals),
				DeployInscriptionId:     deployEvent.InscriptionId.String(),
				DeployInscriptionNumber: deployEvent.InscriptionNumber,
				InscriptionStartNumber:  lo.Ternary(firstInscriptionNumber < 0, deployEvent.InscriptionNumber, firstInscriptionNumber),
				InscriptionEndNumber:    lo.Ternary(lastInscriptionNumber < 0, deployEvent.InscriptionNumber, lastInscriptionNumber),
			},
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
