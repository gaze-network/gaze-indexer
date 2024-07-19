package nodesale

import (
	"bytes"
	"context"
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"google.golang.org/protobuf/encoding/protojson"
)

func (p *Processor) processDeploy(ctx context.Context, qtx datagateway.NodesaleDataGatewayWithTx, block *types.Block, event nodesaleEvent) error {
	valid := true
	deploy := event.eventMessage.Deploy

	sellerPubKeyBytes, err := hex.DecodeString(deploy.SellerPublicKey)
	if err != nil {
		valid = false
	}

	if valid {
		sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes)
		if err != nil {
			valid = false
		}
		xOnlySellerPubKey := btcec.ToSerialized(sellerPubKey).SchnorrSerialized()
		xOnlyTxPubKey := btcec.ToSerialized(event.txPubkey).SchnorrSerialized()

		if valid && !bytes.Equal(xOnlySellerPubKey[:], xOnlyTxPubKey[:]) {
			valid = false
		}
	}

	tiers := make([][]byte, len(deploy.Tiers))
	for i, tier := range deploy.Tiers {
		tierJson, err := protojson.Marshal(tier)
		if err != nil {
			return errors.Wrap(err, "Failed to parse tiers to json")
		}
		tiers[i] = tierJson
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
		Valid:          valid,
		WalletAddress:  p.pubkeyToPkHashAddress(event.txPubkey).EncodeAddress(),
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}
	if valid {
		err = qtx.AddNodesale(ctx, datagateway.AddNodesaleParams{
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
			return errors.Wrap(err, "Failed to insert nodesale")
		}
	}

	return nil
}
