package nodesale

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	purchasevalidator "github.com/gaze-network/indexer-network/modules/nodesale/internal/validator/purchase"
)

func (p *Processor) ProcessPurchase(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, block *types.Block, event NodeSaleEvent) error {
	purchase := event.EventMessage.Purchase
	payload := purchase.Payload

	validator := purchasevalidator.New()

	validator.EqualXonlyPublicKey(payload.BuyerPublicKey, event.TxPubkey)

	_, deploy, err := validator.NodeSaleExists(ctx, qtx, payload)
	if err != nil {
		return errors.Wrap(err, "cannot query. Something wrong.")
	}

	validator.ValidTimestamp(deploy, block.Header.Timestamp)
	validator.WithinTimeoutBlock(payload.TimeOutBlock, uint64(event.Transaction.BlockHeight))

	validator.VerifySignature(purchase, deploy)

	_, tierMap := validator.ValidTiers(payload, deploy)

	tiers := tierMap.Tiers
	buyingTiersCount := tierMap.BuyingTiersCount
	nodeIdToTier := tierMap.NodeIdToTier

	_, err = validator.ValidUnpurchasedNodes(ctx, qtx, payload)
	if err != nil {
		return errors.Wrap(err, "cannot query. Something wrong.")
	}

	_, meta := validator.ValidPaidAmount(payload, deploy, event.Transaction.TxOut, tiers, buyingTiersCount, p.Network.ChainParams())

	_, err = validator.WithinLimit(ctx, qtx, payload, deploy, tiers, buyingTiersCount)
	if err != nil {
		return errors.Wrap(err, "cannot query. Something wrong.")
	}

	err = qtx.CreateEvent(ctx, entity.NodeSaleEvent{
		TxHash:         event.Transaction.TxHash.String(),
		TxIndex:        int32(event.Transaction.Index),
		Action:         int32(event.EventMessage.Action),
		RawMessage:     event.RawData,
		ParsedMessage:  event.EventJson,
		BlockTimestamp: block.Header.Timestamp,
		BlockHash:      event.Transaction.BlockHash.String(),
		BlockHeight:    event.Transaction.BlockHeight,
		Valid:          validator.Valid,
		WalletAddress:  p.PubkeyToPkHashAddress(event.TxPubkey).EncodeAddress(),
		Metadata:       meta,
		Reason:         validator.Reason,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}

	if validator.Valid {
		// add to node
		for _, nodeId := range payload.NodeIDs {
			err := qtx.CreateNode(ctx, entity.Node{
				SaleBlock:      deploy.BlockHeight,
				SaleTxIndex:    deploy.TxIndex,
				NodeID:         int32(nodeId),
				TierIndex:      nodeIdToTier[nodeId],
				DelegatedTo:    "",
				OwnerPublicKey: payload.BuyerPublicKey,
				PurchaseTxHash: event.Transaction.TxHash.String(),
				DelegateTxHash: "",
			})
			if err != nil {
				return errors.Wrap(err, "Failed to insert node")
			}
		}
	}

	return nil
}
