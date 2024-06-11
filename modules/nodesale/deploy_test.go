package nodesale

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/stretchr/testify/require"
)

func TestDeployInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	prvKey, _ := btcec.NewPrivateKey()
	sellerWallet := p.pubkeyToPkHashAddress(prvKey.PubKey())
	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     t.Name(),
			StartsAt: 100,
			EndsAt:   200,
			Tiers: []*protobuf.Tier{
				{
					PriceSat:      100,
					Limit:         5,
					MaxPerAddress: 100,
				},
				{
					PriceSat:      200,
					Limit:         5,
					MaxPerAddress: 100,
				},
			},
			SellerPublicKey:       "0102030405",
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}

	event, block := assembleTestEvent(prvKey, "0101010101", "0101010101", 0, 0, message)
	p.processDeploy(ctx, qtx, block, event)

	nodesales, _ := qtx.GetNodesale(ctx, gen.GetNodesaleParams{
		BlockHeight: int32(testBlockHeigh) - 1,
		TxIndex:     int32(testTxIndex) - 1,
	})
	require.Len(t, nodesales, 0)
}

func TestDeployValid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	privateKey, _ := btcec.NewPrivateKey()
	pubkeyHex := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())
	sellerWallet := p.pubkeyToPkHashAddress(privateKey.PubKey())
	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     t.Name(),
			StartsAt: 100,
			EndsAt:   200,
			Tiers: []*protobuf.Tier{
				{
					PriceSat:      100,
					Limit:         5,
					MaxPerAddress: 100,
				},
				{
					PriceSat:      200,
					Limit:         5,
					MaxPerAddress: 100,
				},
			},
			SellerPublicKey:       pubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}

	event, block := assembleTestEvent(privateKey, "0202020202", "0202020202", 0, 0, message)
	p.processDeploy(ctx, qtx, block, event)

	nodesales, _ := qtx.GetNodesale(ctx, gen.GetNodesaleParams{
		BlockHeight: int32(testBlockHeigh) - 1,
		TxIndex:     int32(testTxIndex) - 1,
	})
	require.Len(t, nodesales, 1)
}
