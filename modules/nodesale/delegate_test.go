package nodesale

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway/mocks"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDelegate(t *testing.T) {
	ctx := context.Background()
	mockDgTx := mocks.NewNodeSaleDataGatewayWithTx(t)
	p := NewProcessor(mockDgTx, nil, common.NetworkMainnet, nil, 0)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	delegateePrivateKey, _ := btcec.NewPrivateKey()
	delegateePubkeyHex := hex.EncodeToString(delegateePrivateKey.PubKey().SerializeCompressed())

	delegateMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DELEGATE,
		Delegate: &protobuf.ActionDelegate{
			DelegateePublicKey: delegateePubkeyHex,
			NodeIDs:            []uint32{9, 10},
			DeployID: &protobuf.ActionID{
				Block:   uint64(testBlockHeight) - 2,
				TxIndex: uint32(testTxIndex) - 2,
			},
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "131313131313", "131313131313", 0, 0, delegateMessage)

	mockDgTx.EXPECT().CreateEvent(mock.Anything, mock.MatchedBy(func(event entity.NodeSaleEvent) bool {
		return event.Valid == true
	})).Return(nil)

	mockDgTx.EXPECT().GetNodesByIds(mock.Anything, datagateway.GetNodesByIdsParams{
		SaleBlock:   delegateMessage.Delegate.DeployID.Block,
		SaleTxIndex: delegateMessage.Delegate.DeployID.TxIndex,
		NodeIds:     []uint32{9, 10},
	}).Return([]entity.Node{
		{
			SaleBlock:      delegateMessage.Delegate.DeployID.Block,
			SaleTxIndex:    delegateMessage.Delegate.DeployID.TxIndex,
			NodeID:         9,
			TierIndex:      1,
			DelegatedTo:    "",
			OwnerPublicKey: buyerPubkeyHex,
			PurchaseTxHash: mock.Anything,
			DelegateTxHash: "",
		},
		{
			SaleBlock:      delegateMessage.Delegate.DeployID.Block,
			SaleTxIndex:    delegateMessage.Delegate.DeployID.TxIndex,
			NodeID:         10,
			TierIndex:      2,
			DelegatedTo:    "",
			OwnerPublicKey: buyerPubkeyHex,
			PurchaseTxHash: mock.Anything,
			DelegateTxHash: "",
		},
	}, nil)

	mockDgTx.EXPECT().SetDelegates(mock.Anything, datagateway.SetDelegatesParams{
		SaleBlock:      delegateMessage.Delegate.DeployID.Block,
		SaleTxIndex:    int32(delegateMessage.Delegate.DeployID.TxIndex),
		Delegatee:      delegateMessage.Delegate.DelegateePublicKey,
		DelegateTxHash: event.Transaction.TxHash.String(),
		NodeIds:        delegateMessage.Delegate.NodeIDs,
	}).Return(2, nil)

	err := p.ProcessDelegate(ctx, mockDgTx, block, event)
	require.NoError(t, err)
}
