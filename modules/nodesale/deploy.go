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

func (p *Processor) ProcessDeploy(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, block *types.Block, event NodeSaleEvent) error {
	deploy := event.EventMessage.Deploy

	validator := validator.New()

	validator.EqualXonlyPublicKey(deploy.SellerPublicKey, event.TxPubkey)

	err := qtx.CreateEvent(ctx, entity.NodeSaleEvent{
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
			BlockHeight:           uint64(event.Transaction.BlockHeight),
			TxIndex:               event.Transaction.Index,
			Name:                  deploy.Name,
			StartsAt:              time.Unix(int64(deploy.StartsAt), 0),
			EndsAt:                time.Unix(int64(deploy.EndsAt), 0),
			Tiers:                 tiers,
			SellerPublicKey:       deploy.SellerPublicKey,
			MaxPerAddress:         deploy.MaxPerAddress,
			DeployTxHash:          event.Transaction.TxHash.String(),
			MaxDiscountPercentage: int32(deploy.MaxDiscountPercentage),
			SellerWallet:          deploy.SellerWallet,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to insert NodeSale")
		}
	}

	return nil
}
