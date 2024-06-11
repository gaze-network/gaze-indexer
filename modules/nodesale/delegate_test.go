package nodesale

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDelegate(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
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
					Limit:         4,
					MaxPerAddress: 2,
				},
				{
					PriceSat:      400,
					Limit:         50,
					MaxPerAddress: 3,
				},
			},
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
		},
	}
	event, block := assembleTestEvent(sellerPrivateKey, "111111", "111111", 0, 0, deployMessage)
	p.processDeploy(ctx, qtx, block, event)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	payload := &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   uint64(testBlockHeigh) - 1,
			TxIndex: uint32(testTxIndex) - 1,
		},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   uint64(testBlockHeigh) + 5,
		NodeIDs:        []uint32{9, 10, 11},
		TotalAmountSat: 600,
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := sha256.Sum256(payloadBytes)
	signature := ecdsa.Sign(sellerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block = assembleTestEvent(buyerPrivateKey, "1212121212", "1212121212", 0, 0, message)

	addr, _ := btcutil.NewAddressPubKey(sellerPrivateKey.PubKey().SerializeCompressed(), p.network.ChainParams())
	pkscript, _ := txscript.PayToAddrScript(addr.AddressPubKeyHash())
	event.transaction.TxOut = []*types.TxOut{
		{
			PkScript: pkscript,
			Value:    600,
		},
	}
	p.processPurchase(ctx, qtx, block, event)

	delegateePrivateKey, _ := btcec.NewPrivateKey()
	delegateePubkeyHex := hex.EncodeToString(delegateePrivateKey.PubKey().SerializeCompressed())
	delegateMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DELEGATE,
		Delegate: &protobuf.ActionDelegate{
			DelegateePublicKey: delegateePubkeyHex,
			NodeIDs:            []uint32{9, 10},
			DeployID: &protobuf.ActionID{
				Block:   uint64(testBlockHeigh) - 2,
				TxIndex: uint32(testTxIndex) - 2,
			},
		},
	}
	event, block = assembleTestEvent(buyerPrivateKey, "131313131313", "131313131313", 0, 0, delegateMessage)
	p.processDelegate(ctx, qtx, block, event)

	nodes, _ := qtx.GetNodes(ctx, gen.GetNodesParams{
		SaleBlock:   int32(testBlockHeigh) - 3,
		SaleTxIndex: int32(testTxIndex) - 3,
		NodeIds:     []int32{9, 10, 11},
	})
	require.Len(t, nodes, 3)
	for _, node := range nodes {
		if node.NodeID == 9 || node.NodeID == 10 {
			require.NotEmpty(t, node.DelegatedTo)
		} else if node.NodeID == 11 {
			require.Empty(t, node.DelegatedTo)
		} else {
			require.Fail(t, "Unhandled")
		}
	}
}
