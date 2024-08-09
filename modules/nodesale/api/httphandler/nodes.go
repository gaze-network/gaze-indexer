package httphandler

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gofiber/fiber/v2"
)

type nodeRequest struct {
	DeployId           string `query:"deployId"`
	OwnerPublicKey     string `query:"ownerPublicKey"`
	DelegateePublicKey string `query:"delegateePublicKey"`
}

type nodeResponse struct {
	DeployId            string `json:"deployId"`
	NodeId              uint32 `json:"nodeId"`
	TierIndex           int32  `json:"tierIndex"`
	DelegatedTo         string `json:"delegatedTo"`
	OwnerPublicKey      string `json:"ownerPublicKey"`
	PurchaseTxHash      string `json:"purchaseTxHash"`
	DelegateTxHash      string `json:"delegateTxHash"`
	PurchaseBlockHeight int32  `json:"purchaseBlockHeight"`
}

func (h *handler) nodesHandler(ctx *fiber.Ctx) error {
	var request nodeRequest
	err := ctx.QueryParser(&request)
	if err != nil {
		return errors.Wrap(err, "cannot parse query")
	}

	ownerPublicKey := request.OwnerPublicKey
	delegateePublicKey := request.DelegateePublicKey

	var blockHeight int64
	var txIndex int32
	count, err := fmt.Sscanf(request.DeployId, "%d-%d", &blockHeight, &txIndex)
	if count != 2 || err != nil {
		return errs.NewPublicError("Invalid deploy ID")
	}

	var nodes []entity.Node
	if ownerPublicKey == "" {
		nodes, err = h.nodeSaleDg.GetNodesByDeploy(ctx.UserContext(), blockHeight, txIndex)
		if err != nil {
			return errors.Wrap(err, "Can't get nodes from db")
		}
	} else {
		nodes, err = h.nodeSaleDg.GetNodesByPubkey(ctx.UserContext(), datagateway.GetNodesByPubkeyParams{
			SaleBlock:      blockHeight,
			SaleTxIndex:    txIndex,
			OwnerPublicKey: ownerPublicKey,
			DelegatedTo:    delegateePublicKey,
		})
		if err != nil {
			return errors.Wrap(err, "Can't get nodes from db")
		}
	}

	responses := make([]nodeResponse, len(nodes))
	for i, node := range nodes {
		responses[i].DeployId = request.DeployId
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
