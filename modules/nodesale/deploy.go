package nodesale

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/encoding/protojson"
)

func (p *Processor) processDeploy(ctx context.Context, qtx gen.Querier, block *types.Block, event nodesaleEvent) error {
	valid := true
	deploy := event.eventMessage.Deploy
	sellerAddr, err := p.pubkeyToTaprootAddress(deploy.SellerPublicKey, event.rawScript)
	if err != nil || !bytes.Equal(
		[]byte(sellerAddr.EncodeAddress()),
		[]byte(event.txAddress.EncodeAddress()),
	) {
		valid = false
	}
	tiers := make([][]byte, len(deploy.Tiers))
	for i, tier := range deploy.Tiers {
		tierJson, err := protojson.Marshal(tier)
		if err != nil {
			return fmt.Errorf("Failed to parse tiers to json : %w", err)
		}
		tiers[i] = tierJson
	}

	err = qtx.AddEvent(ctx, gen.AddEventParams{
		TxHash:         event.transaction.TxHash.String(),
		TxIndex:        int32(event.transaction.Index),
		Action:         int32(event.eventMessage.Action),
		RawMessage:     event.rawData,
		ParsedMessage:  event.eventJson,
		BlockTimestamp: pgtype.Timestamp{Time: block.Header.Timestamp, Valid: true},
		BlockHash:      event.transaction.BlockHash.String(),
		BlockHeight:    int32(event.transaction.BlockHeight),
		Valid:          valid,
		WalletAddress:  event.txAddress.EncodeAddress(),
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return fmt.Errorf("Failed to insert event : %w", err)
	}
	if valid {
		err = qtx.AddNodesale(ctx, gen.AddNodesaleParams{
			BlockHeight: int32(event.transaction.BlockHeight),
			TxIndex:     int32(event.transaction.Index),
			Name:        deploy.Name,
			StartsAt: pgtype.Timestamp{
				Time:  time.Unix(int64(deploy.StartsAt), 0).UTC(),
				Valid: true,
			},
			EndsAt: pgtype.Timestamp{
				Time:  time.Unix(int64(deploy.EndsAt), 0).UTC(),
				Valid: true,
			},
			Tiers:                 tiers,
			SellerPublicKey:       deploy.SellerPublicKey,
			MaxPerAddress:         int32(deploy.MaxPerAddress),
			DeployTxHash:          event.transaction.TxHash.String(),
			MaxDiscountPercentage: int32(deploy.MaxDiscountPercentage),
		})
		if err != nil {
			return fmt.Errorf("Failed to insert nodesale : %w", err)
		}
	}

	return nil
}
