package nodesale

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway/mocks"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestDeployInvalid(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	prvKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	strangerKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	strangerPubkeyHex := hex.EncodeToString(strangerKey.PubKey().SerializeCompressed())

	sellerWallet := p.PubkeyToPkHashAddress(prvKey.PubKey())

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
			SellerPublicKey:       strangerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}

	event, block := assembleTestEvent(prvKey, "0101010101", "0101010101", 0, 0, message)

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false
	})).Return(nil)

	err = p.ProcessDeploy(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNodeSale")
}

func TestDeployValid(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	privateKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	pubkeyHex := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())

	sellerWallet := p.PubkeyToPkHashAddress(privateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     t.Name(),
			StartsAt: uint32(startAt.UTC().Unix()),
			EndsAt:   uint32(endAt.UTC().Unix()),
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

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == true
	})).Return(nil)

	tiers := lo.Map(message.Deploy.Tiers, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})

	mockDgTx.EXPECT().CreateNodeSale(mock.Anything, entity.NodeSale{
		BlockHeight:           uint64(event.Transaction.BlockHeight),
		TxIndex:               uint32(event.Transaction.Index),
		Name:                  message.Deploy.Name,
		StartsAt:              time.Unix(int64(message.Deploy.StartsAt), 0),
		EndsAt:                time.Unix(int64(message.Deploy.EndsAt), 0),
		Tiers:                 tiers,
		SellerPublicKey:       message.Deploy.SellerPublicKey,
		MaxPerAddress:         message.Deploy.MaxPerAddress,
		DeployTxHash:          event.Transaction.TxHash.String(),
		MaxDiscountPercentage: int32(message.Deploy.MaxDiscountPercentage),
		SellerWallet:          message.Deploy.SellerWallet,
	}).Return(nil)

	p.ProcessDeploy(ctx, mockDgTx, block, event)
}
