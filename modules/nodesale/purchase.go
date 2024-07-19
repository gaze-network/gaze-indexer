package nodesale

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	purchasevalidator "github.com/gaze-network/indexer-network/modules/nodesale/internal/validator/purchase"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

func (p *Processor) processPurchase(ctx context.Context, qtx datagateway.NodesaleDataGatewayWithTx, block *types.Block, event nodesaleEvent) error {
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
	validator.WithinTimeoutBlock(payload, uint64(event.transaction.BlockHeight))

	_, err = validator.VerifySignature(purchase, deploy)
	if err != nil {
		logger.DebugContext(ctx, "incorrect Signature format", slogx.Error(err))
	}

	_, tiers, buyingTiersCount, nodeIdToTier, err := validator.ValidTiers(payload, deploy)
	if err != nil {
		logger.DebugContext(ctx, "invalid nodesale tiers data", slogx.Error(err))
	}

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

	metaDataBytes, _ := json.Marshal(meta)
	if meta == nil {
		metaDataBytes = []byte("{}")
	}

	err = qtx.AddEvent(ctx, datagateway.AddEventParams{
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
		Metadata:       metaDataBytes,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}

	if validator.Valid {
		// add to node
		for _, nodeId := range payload.NodeIDs {
			err := qtx.AddNode(ctx, datagateway.AddNodeParams{
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
