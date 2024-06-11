package httphandler

import (
	"fmt"

	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

type handler struct {
	repository *repository.Repository
}

func New(repo *repository.Repository) *handler {
	h := handler{}
	h.repository = repo
	return &h
}

func (h *handler) infoHandler(ctx *fiber.Ctx) error {
	block, err := h.repository.Queries.GetLastProcessedBlock(ctx.UserContext())
	if err != nil {
		return fmt.Errorf("Cannot get last processed block : %w", err)
	}
	err = ctx.JSON(infoResponse{
		IndexedBlockHeight: block.BlockHeight,
		IndexedBlockHash:   block.BlockHash,
	})
	if err != nil {
		return fmt.Errorf("Go fiber cannot parse JSON: %w", err)
	}
	return nil
}

func (h *handler) deployHandler(ctx *fiber.Ctx) error {
	deployId := ctx.Params("deployId")
	if deployId == "" {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Go fiber cannot send status: %w", err)
		}
		return nil
	}
	var blockHeight, txIndex int32
	count, err := fmt.Sscanf(deployId, "%d-%d", &blockHeight, &txIndex)
	if count != 2 || err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Go fiber cannot send status: %w", err)
		}
		return nil
	}
	deploys, err := h.repository.Queries.GetNodesale(ctx.UserContext(), gen.GetNodesaleParams{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
	})
	if err != nil {
		return fmt.Errorf("Cannot get nodesale from db: %w", err)
	}
	if len(deploys) < 1 {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Go fiber cannot send status: %w", err)
		}
		return nil
	}

	deploy := deploys[0]

	nodeCount, err := h.repository.Queries.GetNodeCountByTierIndex(ctx.UserContext(), gen.GetNodeCountByTierIndexParams{
		SaleBlock:   deploy.BlockHeight,
		SaleTxIndex: deploy.TxIndex,
		FromTier:    0,
		ToTier:      int32(len(deploy.Tiers) - 1),
	})
	if err != nil {
		return fmt.Errorf("Cannot get node count from db : %w", err)
	}

	tiers := make([]protobuf.Tier, len(deploy.Tiers))
	tierResponses := make([]tierResponse, len(deploy.Tiers))
	for i, tierJson := range deploy.Tiers {
		tier := &tiers[i]
		err := protojson.Unmarshal(tierJson, tier)
		if err != nil {
			return fmt.Errorf("Failed to decode tiers json : %w", err)
		}
		tierResponses[i].Limit = tiers[i].Limit
		tierResponses[i].MaxPerAddress = tiers[i].MaxPerAddress
		tierResponses[i].PriceSat = tiers[i].PriceSat
		tierResponses[i].Sold = nodeCount[i].Count
	}

	err = ctx.JSON(&deployResponse{
		Id:              deployId,
		Name:            deploy.Name,
		StartAt:         deploy.StartsAt.Time.UTC(),
		EndAt:           deploy.EndsAt.Time.UTC(),
		Tiers:           tierResponses,
		SellerPublicKey: deploy.SellerPublicKey,
		MaxPerAddress:   deploy.MaxPerAddress,
		DeployTxHash:    deploy.DeployTxHash,
	})
	if err != nil {
		return fmt.Errorf("Go fiber cannot parse JSON: %w", err)
	}
	return nil
}

func (h *handler) nodesHandler(ctx *fiber.Ctx) error {
	deployId := ctx.Query("deployId")
	if deployId == "" {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Go fiber cannot send status: %w", err)
		}
		return nil
	}
	ownerPublicKey := ctx.Query("ownerPublicKey")
	delegateePublicKey := ctx.Query("delegateePublicKey")

	var blockHeight, txIndex int32
	count, err := fmt.Sscanf(deployId, "%d-%d", &blockHeight, &txIndex)
	if count != 2 || err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Go fiber cannot send status: %w", err)
		}
		return nil
	}

	nodes, err := h.repository.Queries.GetNodesByPubkey(ctx.UserContext(), gen.GetNodesByPubkeyParams{
		SaleBlock:      blockHeight,
		SaleTxIndex:    txIndex,
		OwnerPublicKey: ownerPublicKey,
		DelegatedTo:    delegateePublicKey,
	})
	if err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Can't get nodes from db: %w", err)
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
		responses[i].PurchaseBlockHeight = node.TxIndex
	}

	err = ctx.JSON(responses)
	if err != nil {
		return fmt.Errorf("Go fiber cannot parse JSON: %w", err)
	}
	return nil
}

func (h *handler) eventsHandler(ctx *fiber.Ctx) error {
	walletAddress := ctx.Query("walletAddress")

	events, err := h.repository.Queries.GetEventsByWallet(ctx.UserContext(), walletAddress)
	if err != nil {
		err := ctx.SendStatus(404)
		if err != nil {
			return fmt.Errorf("Can't get events from db: %w", err)
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
		responses[i].BlockTimestamp = event.BlockTimestamp.Time
		responses[i].BlockHash = event.BlockHash
	}

	err = ctx.JSON(responses)
	if err != nil {
		return fmt.Errorf("Go fiber cannot parse JSON: %w", err)
	}
	return nil
}
