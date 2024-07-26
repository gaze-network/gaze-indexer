package nodesale

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway/mocks"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/validator"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/validator/purchase"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestInvalidPurchase(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	buyerPrivateKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   111,
					TxIndex: 1,
				},
				NodeIDs:        []uint32{1, 2},
				BuyerPublicKey: buyerPubkeyHex,
				TotalAmountSat: 500,
				TimeOutBlock:   uint64(testBlockHeight) + 5,
			},
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "030303030303", "030303030303", 0, 0, message)

	mockDgTx.EXPECT().GetNodeSale(mock.Anything, mock.Anything).Return(nil, nil)

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false
	})).Return(nil)

	err = p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNode")
}

func TestInvalidBuyerKey(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	strangerPrivateKey, _ := btcec.NewPrivateKey()
	strangerPrivateKeyHex := hex.EncodeToString(strangerPrivateKey.PubKey().SerializeCompressed())

	buyerPrivateKey, _ := btcec.NewPrivateKey()

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   100,
					TxIndex: 1,
				},
				NodeIDs:        []uint32{1, 2},
				BuyerPublicKey: strangerPrivateKeyHex,
				TotalAmountSat: 200,
				TimeOutBlock:   uint64(testBlockHeight) + 5,
			},
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "0707070707", "0707070707", 0, 0, message)
	block.Header.Timestamp = time.Now().UTC()

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false && event.Reason == validator.INVALID_PUBKEY
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNode")
}

func TestInvalidTimestamp(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
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
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})
	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   100,
					TxIndex: 1,
				},
				NodeIDs:        []uint32{1, 2},
				BuyerPublicKey: buyerPubkeyHex,
				TotalAmountSat: 200,
				TimeOutBlock:   uint64(testBlockHeight) + 5,
			},
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "050505050505", "050505050505", 0, 0, message)

	block.Header.Timestamp = time.Now().UTC().Add(time.Hour * 2)

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false && event.Reason == purchase.PURCHASE_TIMEOUT
	})).Return(nil)

	err = p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNode")
}

func TestTimeOut(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
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
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})

	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   100,
					TxIndex: 1,
				},
				NodeIDs:        []uint32{1, 2},
				BuyerPublicKey: buyerPubkeyHex,
				TimeOutBlock:   uint64(testBlockHeight) - 5,
				TotalAmountSat: 200,
			},
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "090909090909", "090909090909", 0, 0, message)

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false && event.Reason == purchase.BLOCK_HEIGHT_TIMEOUT
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNode")
}

func TestSignatureInvalid(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
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
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})
	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	payload := &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   100,
			TxIndex: 1,
		},
		NodeIDs:        []uint32{1, 2},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   testBlockHeight + 5,
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := chainhash.DoubleHashB(payloadBytes)
	signature := ecdsa.Sign(buyerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "0B0B0B", "0B0B0B", 0, 0, message)

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false && event.Reason == purchase.INVALID_SIGNATURE
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNode")
}

func TestValidPurchase(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
		{
			PriceSat:      100,
			Limit:         5,
			MaxPerAddress: 100,
		},
		{
			PriceSat:      200,
			Limit:         4,
			MaxPerAddress: 2,
		},
		{
			PriceSat:      400,
			Limit:         3,
			MaxPerAddress: 100,
		},
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})

	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	mockDgTx.EXPECT().GetNodesByIds(mock.Anything, mock.Anything).Return(nil, nil)

	mockDgTx.EXPECT().GetNodesByOwner(mock.Anything, mock.Anything).Return(nil, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	payload := &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   100,
			TxIndex: 1,
		},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   uint64(testBlockHeight) + 5,
		NodeIDs:        []uint32{0, 5, 6, 9},
		TotalAmountSat: 500,
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := chainhash.DoubleHashB(payloadBytes)
	signature := ecdsa.Sign(sellerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "0D0D0D0D", "0D0D0D0D", 0, 0, message)
	event.InputValue = 500

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == true && event.Reason == ""
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 0 &&
			node.TierIndex == 0 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 5 &&
			node.TierIndex == 1 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 6 &&
			node.TierIndex == 1 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 9 &&
			node.TierIndex == 2 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)
}

/*func TestMismatchPayment(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
		{
			PriceSat:      100,
			Limit:         5,
			MaxPerAddress: 100,
		},
		{
			PriceSat:      200,
			Limit:         4,
			MaxPerAddress: 2,
		},
		{
			PriceSat:      400,
			Limit:         3,
			MaxPerAddress: 100,
		},
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})

	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	mockDgTx.EXPECT().GetNodesByIds(mock.Anything, mock.Anything).Return(nil, nil)

	mockDgTx.EXPECT().GetNodesByOwner(mock.Anything, mock.Anything).Return(nil, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	payload := &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   100,
			TxIndex: 1,
		},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   uint64(testBlockHeight) + 5,
		NodeIDs:        []uint32{0, 5, 6, 9},
		TotalAmountSat: 500,
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := chainhash.DoubleHashB(payloadBytes)
	signature := ecdsa.Sign(sellerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "0D0D0D0D", "0D0D0D0D", 0, 0, message)
	event.InputValue = 500

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == true && event.Reason == ""
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 0 &&
			node.TierIndex == 0 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 5 &&
			node.TierIndex == 1 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 6 &&
			node.TierIndex == 1 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	mockDgTx.EXPECT().CreateNode(mock.Anything, mock.MatchedBy(func(node entity.Node) bool {
		return node.NodeID == 9 &&
			node.TierIndex == 2 &&
			node.OwnerPublicKey == buyerPubkeyHex &&
			node.PurchaseTxHash == event.Transaction.TxHash.String() &&
			node.SaleBlock == 100 &&
			node.SaleTxIndex == 1
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)
}*/

func TestBuyingLimit(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
		{
			PriceSat:      100,
			Limit:         5,
			MaxPerAddress: 100,
		},
		{
			PriceSat:      200,
			Limit:         4,
			MaxPerAddress: 2,
		},
		{
			PriceSat:      400,
			Limit:         50,
			MaxPerAddress: 100,
		},
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})

	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         2,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	mockDgTx.EXPECT().GetNodesByIds(mock.Anything, mock.Anything).Return(nil, nil)

	mockDgTx.EXPECT().GetNodesByOwner(mock.Anything, datagateway.GetNodesByOwnerParams{
		SaleBlock:      100,
		SaleTxIndex:    1,
		OwnerPublicKey: buyerPubkeyHex,
	}).Return([]entity.Node{
		{
			SaleBlock:      100,
			SaleTxIndex:    1,
			NodeID:         9,
			TierIndex:      2,
			OwnerPublicKey: buyerPubkeyHex,
		},
		{
			SaleBlock:      100,
			SaleTxIndex:    1,
			NodeID:         10,
			TierIndex:      2,
			OwnerPublicKey: buyerPubkeyHex,
		},
	}, nil)

	payload := &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   100,
			TxIndex: 1,
		},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   uint64(testBlockHeight) + 5,
		NodeIDs:        []uint32{11},
		TotalAmountSat: 600,
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := chainhash.DoubleHashB(payloadBytes)
	signature := ecdsa.Sign(sellerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "22222222", "22222222", 0, 0, message)
	event.InputValue = 600

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false && event.Reason == purchase.OVER_LIMIT_PER_ADDR
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)

	mockDgTx.AssertNotCalled(t, "CreateNode")
}

func TestBuyingTierLimit(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	sellerWallet := p.PubkeyToPkHashAddress(sellerPrivateKey.PubKey())

	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)

	tiers := lo.Map([]*protobuf.Tier{
		{
			PriceSat:      100,
			Limit:         5,
			MaxPerAddress: 100,
		},
		{
			PriceSat:      200,
			Limit:         4,
			MaxPerAddress: 2,
		},
		{
			PriceSat:      400,
			Limit:         50,
			MaxPerAddress: 3,
		},
	}, func(tier *protobuf.Tier, _ int) []byte {
		tierJson, err := protojson.Marshal(tier)
		require.NoError(t, err)
		return tierJson
	})

	mockDgTx.EXPECT().GetNodeSale(mock.Anything, datagateway.GetNodeSaleParams{
		BlockHeight: 100,
		TxIndex:     1,
	}).Return([]entity.NodeSale{
		{
			BlockHeight:           100,
			TxIndex:               1,
			Name:                  t.Name(),
			StartsAt:              startAt,
			EndsAt:                endAt,
			Tiers:                 tiers,
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			DeployTxHash:          "040404040404",
			MaxDiscountPercentage: 50,
			SellerWallet:          sellerWallet.EncodeAddress(),
		},
	}, nil)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	mockDgTx.EXPECT().GetNodesByIds(mock.Anything, mock.Anything).Return(nil, nil)

	mockDgTx.EXPECT().GetNodesByOwner(mock.Anything, datagateway.GetNodesByOwnerParams{
		SaleBlock:      100,
		SaleTxIndex:    1,
		OwnerPublicKey: buyerPubkeyHex,
	}).Return([]entity.Node{
		{
			SaleBlock:      100,
			SaleTxIndex:    1,
			NodeID:         9,
			TierIndex:      2,
			OwnerPublicKey: buyerPubkeyHex,
		},
		{
			SaleBlock:      100,
			SaleTxIndex:    1,
			NodeID:         10,
			TierIndex:      2,
			OwnerPublicKey: buyerPubkeyHex,
		},
		{
			SaleBlock:      100,
			SaleTxIndex:    1,
			NodeID:         11,
			TierIndex:      2,
			OwnerPublicKey: buyerPubkeyHex,
		},
	}, nil)

	payload := &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   100,
			TxIndex: 1,
		},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   uint64(testBlockHeight) + 5,
		NodeIDs:        []uint32{12, 13, 14},
		TotalAmountSat: 600,
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := chainhash.DoubleHashB(payloadBytes)
	signature := ecdsa.Sign(sellerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "10101010", "10101010", 0, 0, message)
	event.InputValue = 600

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == false && event.Reason == purchase.OVER_LIMIT_PER_TIER
	})).Return(nil)

	err := p.ProcessPurchase(ctx, mockDgTx, block, event)
	require.NoError(t, err)
}
