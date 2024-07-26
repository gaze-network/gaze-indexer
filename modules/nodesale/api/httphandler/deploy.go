package httphandler

import (
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

type deployRequest struct {
	DeployID string `params:"deployId"`
}

type tierResponse struct {
	PriceSat      uint32 `json:"priceSat"`
	Limit         uint32 `json:"limit"`
	MaxPerAddress uint32 `json:"maxPerAddress"`
	Sold          int64  `json:"sold"`
}

type deployResponse struct {
	Id              string         `json:"id"`
	Name            string         `json:"name"`
	StartsAt        time.Time      `json:"startsAt"`
	EndsAt          time.Time      `json:"endsAt"`
	Tiers           []tierResponse `json:"tiers"`
	SellerPublicKey string         `json:"sellerPublicKey"`
	MaxPerAddress   uint32         `json:"maxPerAddress"`
	DeployTxHash    string         `json:"deployTxHash"`
}

func (h *handler) deployHandler(ctx *fiber.Ctx) error {
	var request deployRequest
	err := ctx.ParamsParser(&request)
	if err != nil {
		return errors.Wrap(err, "cannot parse param")
	}
	var blockHeight uint64
	var txIndex uint32
	count, err := fmt.Sscanf(request.DeployID, "%d-%d", &blockHeight, &txIndex)
	if count != 2 || err != nil {
		return errs.NewPublicError("Invalid deploy ID")
	}
	deploys, err := h.nodeSaleDg.GetNodeSale(ctx.UserContext(), datagateway.GetNodeSaleParams{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
	})
	if err != nil {
		return errors.Wrap(err, "Cannot get NodeSale from db")
	}
	if len(deploys) < 1 {
		return errs.NewPublicError("NodeSale not found")
	}

	deploy := deploys[0]

	nodeCount, err := h.nodeSaleDg.GetNodeCountByTierIndex(ctx.UserContext(), datagateway.GetNodeCountByTierIndexParams{
		SaleBlock:   deploy.BlockHeight,
		SaleTxIndex: deploy.TxIndex,
		FromTier:    0,
		ToTier:      uint32(len(deploy.Tiers) - 1),
	})
	if err != nil {
		return errors.Wrap(err, "Cannot get node count from db")
	}

	tiers := make([]protobuf.Tier, len(deploy.Tiers))
	tierResponses := make([]tierResponse, len(deploy.Tiers))
	for i, tierJson := range deploy.Tiers {
		tier := &tiers[i]
		err := protojson.Unmarshal(tierJson, tier)
		if err != nil {
			return errors.Wrap(err, "Failed to decode tiers json")
		}
		tierResponses[i].Limit = tiers[i].Limit
		tierResponses[i].MaxPerAddress = tiers[i].MaxPerAddress
		tierResponses[i].PriceSat = tiers[i].PriceSat
		tierResponses[i].Sold = nodeCount[i].Count
	}

	err = ctx.JSON(&deployResponse{
		Id:              request.DeployID,
		Name:            deploy.Name,
		StartsAt:        deploy.StartsAt,
		EndsAt:          deploy.EndsAt,
		Tiers:           tierResponses,
		SellerPublicKey: deploy.SellerPublicKey,
		MaxPerAddress:   deploy.MaxPerAddress,
		DeployTxHash:    deploy.DeployTxHash,
	})
	if err != nil {
		return errors.Wrap(err, "Go fiber cannot parse JSON")
	}
	return nil
}
