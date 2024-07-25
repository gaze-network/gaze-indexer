package nodesale

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	purchasevalidator "github.com/gaze-network/indexer-network/modules/nodesale/internal/validator/purchase"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

func (p *Processor) processPurchase(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, block *types.Block, event nodeSaleEvent) error {
	purchase := event.eventMessage.Purchase
	payload := purchase.Payload

	validator := purchasevalidator.New()

	_, err := validator.EqualXonlyPublicKey(payload.BuyerPublicKey, event.txPubkey)
	if err != nil {
		logger.DebugContext(ctx, "Invalid public key", slogx.Error(err))
	}

	_, deploy, err := validator.NodeSaleExists(ctx, qtx, payload)
	if err != nil {
		return errors.Wrap(err, "Cannot connect to datagateway")
	}

	validator.ValidTimestamp(deploy, block.Header.Timestamp)
	validator.WithinTimeoutBlock(payload.TimeOutBlock, uint64(event.transaction.BlockHeight))

	_, err = validator.VerifySignature(purchase, deploy)
	if err != nil {
		logger.DebugContext(ctx, "incorrect Signature format", slogx.Error(err))
	}

	_, tierMap, err := validator.ValidTiers(payload, deploy)
	if err != nil {
		logger.DebugContext(ctx, "invalid NodeSale tiers data", slogx.Error(err))
	}
	tiers := tierMap.Tiers
	buyingTiersCount := tierMap.BuyingTiersCount
	nodeIdToTier := tierMap.NodeIdToTier

	_, err = validator.ValidUnpurchasedNodes(ctx, qtx, payload)
	if err != nil {
		return errors.Wrap(err, "cannot connect to datagateway")
	}

	_, meta, err := validator.ValidPaidAmount(payload, deploy, event.transaction.TxOut, tiers, buyingTiersCount, p.network.ChainParams())
	if err != nil {
		logger.DebugContext(ctx, "Invalid seller address", slogx.Error(err))
	}

	_, err = validator.WithinLimit(ctx, qtx, payload, deploy, tiers, buyingTiersCount)
	if err != nil {
		return errors.Wrap(err, "Cannot connect to datagateway")
	}

	err = qtx.CreateEvent(ctx, entity.NodeSaleEvent{
		TxHash:         event.transaction.TxHash.String(),
		TxIndex:        int32(event.transaction.Index),
		Action:         int32(event.eventMessage.Action),
		RawMessage:     event.rawData,
		ParsedMessage:  event.eventJson,
		BlockTimestamp: block.Header.Timestamp,
		BlockHash:      event.transaction.BlockHash.String(),
		BlockHeight:    event.transaction.BlockHeight,
		Valid:          validator.Valid,
		WalletAddress:  p.pubkeyToPkHashAddress(event.txPubkey).EncodeAddress(),
		Metadata:       meta,
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
				PurchaseTxHash: event.transaction.TxHash.String(),
				DelegateTxHash: "",
			})
			if err != nil {
				return errors.Wrap(err, "Failed to insert node")
			}
		}
	}

	return nil
}
