package httphandler

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

type handler struct {
	datagateway datagateway.NodesaleDataGateway
}

func New(datagateway datagateway.NodesaleDataGateway) *handler {
	h := handler{}
	h.datagateway = datagateway
	return &h
}

func (h *handler) infoHandler(ctx *fiber.Ctx) error {
	block, err := h.datagateway.GetLastProcessedBlock(ctx.UserContext())
	if err != nil {
		return errors.Wrap(err, "Cannot get last processed block")
	}
	err = ctx.JSON(infoResponse{
		IndexedBlockHeight: block.BlockHeight,
		IndexedBlockHash:   block.BlockHash,
	})
	if err != nil {
		return errors.Wrap(err, "Go fiber cannot parse JSON")
	}
	return nil
}

func (h *handler) deployHandler(ctx *fiber.Ctx) error {
	deployId := ctx.Params("deployId")
	if deployId == "" {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Go fiber cannot send status")
		}
		return nil
	}
	var blockHeight int64
	var txIndex int32
	count, err := fmt.Sscanf(deployId, "%d-%d", &blockHeight, &txIndex)
	if count != 2 || err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Go fiber cannot send status")
		}
		return nil
	}
	deploys, err := h.datagateway.GetNodesale(ctx.UserContext(), datagateway.GetNodesaleParams{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
	})
	if err != nil {
		return errors.Wrap(err, "Cannot get nodesale from db")
	}
	if len(deploys) < 1 {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Go fiber cannot send status")
		}
		return nil
	}

	deploy := deploys[0]

	nodeCount, err := h.datagateway.GetNodeCountByTierIndex(ctx.UserContext(), datagateway.GetNodeCountByTierIndexParams{
		SaleBlock:   deploy.BlockHeight,
		SaleTxIndex: deploy.TxIndex,
		FromTier:    0,
		ToTier:      int32(len(deploy.Tiers) - 1),
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
		Id:              deployId,
		Name:            deploy.Name,
		StartAt:         deploy.StartsAt,
		EndAt:           deploy.EndsAt,
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

func (h *handler) nodesHandler(ctx *fiber.Ctx) error {
	deployId := ctx.Query("deployId")
	if deployId == "" {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Go fiber cannot send status")
		}
		return nil
	}
	ownerPublicKey := ctx.Query("ownerPublicKey")
	delegateePublicKey := ctx.Query("delegateePublicKey")

	var blockHeight int64
	var txIndex int32
	count, err := fmt.Sscanf(deployId, "%d-%d", &blockHeight, &txIndex)
	if count != 2 || err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Go fiber cannot send status")
		}
		return nil
	}

	nodes, err := h.datagateway.GetNodesByPubkey(ctx.UserContext(), datagateway.GetNodesByPubkeyParams{
		SaleBlock:      blockHeight,
		SaleTxIndex:    txIndex,
		OwnerPublicKey: ownerPublicKey,
		DelegatedTo:    delegateePublicKey,
	})
	if err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Can't get nodes from db")
		}
		return nil
	}
	responses := make([]nodeResponse, len(nodes))
	for i, node := range nodes {
		responses[i].DeployId = deployId
		responses[i].NodeId = node.NodeID
		responses[i].TierIndex = node.TierIndex
		responses[i].DelegatedTo = node.DelegatedTo
		responses[i].OwnerPublicKey = node.OwnerPublicKey
		responses[i].PurchaseTxHash = node.PurchaseTxHash
		responses[i].DelegateTxHash = node.DelegateTxHash
		responses[i].PurchaseBlockHeight = txIndex
	}

	err = ctx.JSON(responses)
	if err != nil {
		return errors.Wrap(err, "Go fiber cannot parse JSON")
	}
	return nil
}

func (h *handler) eventsHandler(ctx *fiber.Ctx) error {
	walletAddress := ctx.Query("walletAddress")

	events, err := h.datagateway.GetEventsByWallet(ctx.UserContext(), walletAddress)
	if err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return errors.Wrap(err, "Can't get events from db")
		}
		return nil
	}

	responses := make([]eventResposne, len(events))
	for i, event := range events {
		responses[i].TxHash = event.TxHash
		responses[i].BlockHeight = event.BlockHeight
		responses[i].TxIndex = event.TxIndex
		responses[i].WalletAddress = event.WalletAddress
		responses[i].Action = protobuf.Action_name[event.Action]
		responses[i].ParsedMessage = event.ParsedMessage
		responses[i].BlockTimestamp = event.BlockTimestamp
		responses[i].BlockHash = event.BlockHash
	}

	err = ctx.JSON(responses)
	if err != nil {
		return errors.Wrap(err, "Go fiber cannot parse JSON")
	}
	return nil
}
