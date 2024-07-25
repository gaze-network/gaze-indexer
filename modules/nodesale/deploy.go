package nodesale

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/validator"
	"google.golang.org/protobuf/encoding/protojson"
)

func (p *Processor) processDeploy(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, block *types.Block, event nodeSaleEvent) error {
	deploy := event.eventMessage.Deploy

	validator := validator.New()

	validator.EqualXonlyPublicKey(deploy.SellerPublicKey, event.txPubkey)

	err := qtx.CreateEvent(ctx, entity.NodeSaleEvent{
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
		Metadata:       nil,
		Reason:         validator.Reason,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}
	if validator.Valid {
		tiers := make([][]byte, len(deploy.Tiers))
		for i, tier := range deploy.Tiers {
			tierJson, err := protojson.Marshal(tier)
			if err != nil {
				return errors.Wrap(err, "Failed to parse tiers to json")
			}
			tiers[i] = tierJson
		}
		err = qtx.CreateNodeSale(ctx, entity.NodeSale{
			BlockHeight:           event.transaction.BlockHeight,
			TxIndex:               int32(event.transaction.Index),
			Name:                  deploy.Name,
			StartsAt:              time.Unix(int64(deploy.StartsAt), 0),
			EndsAt:                time.Unix(int64(deploy.EndsAt), 0),
			Tiers:                 tiers,
			SellerPublicKey:       deploy.SellerPublicKey,
			MaxPerAddress:         int32(deploy.MaxPerAddress),
			DeployTxHash:          event.transaction.TxHash.String(),
			MaxDiscountPercentage: int32(deploy.MaxDiscountPercentage),
			SellerWallet:          deploy.SellerWallet,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to insert NodeSale")
		}
	}

	return nil
}
