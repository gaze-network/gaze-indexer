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
	"google.golang.org/protobuf/proto"
)

func TestInvalidPurchase(t *testing.T) {
	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   111,
					TxIndex: 1,
				},
				BuyerPublicKey: buyerPubkeyHex,
			},
		},
	}

	event, block := assembleTestEvent(buyerPrivateKey, "030303030303", "030303030303", 0, 0, message)

	p.processPurchase(ctx, qtx, block, event)
}

func TestInvalidTimestamp(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     "deploy_for_puchase",
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
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
		},
	}
	event, block := assembleTestEvent(sellerPrivateKey, "040404040404", "040404040404", 0, 0, deployMessage)
	p.processDeploy(ctx, qtx, block, event)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   uint64(testBlockHeigh) - 1,
					TxIndex: uint32(testTxIndex) - 1,
				},
				BuyerPublicKey: buyerPubkeyHex,
			},
		},
	}

	event, block = assembleTestEvent(buyerPrivateKey, "050505050505", "050505050505", 0, 0, message)
	block.Header.Timestamp = time.Now().UTC().Add(time.Hour * 2)
	p.processPurchase(ctx, qtx, block, event)
}

func TestInvalidBuyerKey(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     "deploy_for_puchase",
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
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
		},
	}
	event, block := assembleTestEvent(sellerPrivateKey, "060606060606", "060606060606", 0, 0, deployMessage)
	p.processDeploy(ctx, qtx, block, event)

	buyerPrivateKey, _ := btcec.NewPrivateKey()

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   uint64(testBlockHeigh) - 1,
					TxIndex: uint32(testTxIndex) - 1,
				},
				BuyerPublicKey: sellerPubkeyHex,
			},
		},
	}

	event, block = assembleTestEvent(buyerPrivateKey, "0707070707", "0707070707", 0, 0, message)
	block.Header.Timestamp = time.Now().UTC().Add(time.Hour * 2)
	p.processPurchase(ctx, qtx, block, event)
}

func TestTimeOut(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     "deploy_for_puchase",
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
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
		},
	}
	event, block := assembleTestEvent(sellerPrivateKey, "0808080808", "0808080808", 0, 0, deployMessage)
	p.processDeploy(ctx, qtx, block, event)

	buyerPrivateKey, _ := btcec.NewPrivateKey()
	buyerPubkeyHex := hex.EncodeToString(buyerPrivateKey.PubKey().SerializeCompressed())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload: &protobuf.PurchasePayload{
				DeployID: &protobuf.ActionID{
					Block:   uint64(testBlockHeigh) - 1,
					TxIndex: uint32(testTxIndex) - 1,
				},
				BuyerPublicKey: buyerPubkeyHex,
				TimeOutBlock:   uint64(testBlockHeigh) - 5,
			},
		},
	}

	event, block = assembleTestEvent(buyerPrivateKey, "090909090909", "090909090909", 0, 0, message)
	p.processPurchase(ctx, qtx, block, event)
}

func TestSignatureInvalid(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     "deploy_for_puchase",
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
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
		},
	}
	event, block := assembleTestEvent(sellerPrivateKey, "0A0A0A0A", "0A0A0A0A", 0, 0, deployMessage)
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
	}

	payloadBytes, _ := proto.Marshal(payload)
	payloadHash := sha256.Sum256(payloadBytes)
	signature := ecdsa.Sign(buyerPrivateKey, payloadHash[:])
	signatureHex := hex.EncodeToString(signature.Serialize())

	message := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block = assembleTestEvent(buyerPrivateKey, "0B0B0B", "0B0B0B", 0, 0, message)
	p.processPurchase(ctx, qtx, block, event)
}

func TestValidPurchase(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     "deploy_for_puchase",
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
					Limit:         3,
					MaxPerAddress: 100,
				},
			},
			SellerPublicKey:       sellerPubkeyHex,
			MaxPerAddress:         100,
			MaxDiscountPercentage: 50,
		},
	}
	event, block := assembleTestEvent(sellerPrivateKey, "0C0C0C0C0C", "0C0C0C0C0C", 0, 0, deployMessage)
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
		NodeIDs:        []uint32{0, 5, 6, 9},
		TotalAmountSat: 500,
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

	event, block = assembleTestEvent(buyerPrivateKey, "0D0D0D0D", "0D0D0D0D", 0, 0, message)

	addr, _ := btcutil.NewAddressPubKey(sellerPrivateKey.PubKey().SerializeCompressed(), p.network.ChainParams())
	pkscript, _ := txscript.PayToAddrScript(addr.AddressPubKeyHash())
	event.transaction.TxOut = []*types.TxOut{
		{
			PkScript: pkscript,
			Value:    500,
		},
	}
	p.processPurchase(ctx, qtx, block, event)
}

func TestBuyingLimit(t *testing.T) {
	sellerPrivateKey, _ := btcec.NewPrivateKey()
	sellerPubkeyHex := hex.EncodeToString(sellerPrivateKey.PubKey().SerializeCompressed())
	startAt := time.Now().Add(time.Hour * -1)
	endAt := time.Now().Add(time.Hour * 1)
	deployMessage := &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_DEPLOY,
		Deploy: &protobuf.ActionDeploy{
			Name:     "deploy_for_puchase",
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
	event, block := assembleTestEvent(sellerPrivateKey, "0E0E0E0E", "0E0E0E0E", 0, 0, deployMessage)
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

	event, block = assembleTestEvent(buyerPrivateKey, "0F0F0F0F0F", "0F0F0F0F0F", 0, 0, message)

	addr, _ := btcutil.NewAddressPubKey(sellerPrivateKey.PubKey().SerializeCompressed(), p.network.ChainParams())
	pkscript, _ := txscript.PayToAddrScript(addr.AddressPubKeyHash())
	event.transaction.TxOut = []*types.TxOut{
		{
			PkScript: pkscript,
			Value:    600,
		},
	}
	p.processPurchase(ctx, qtx, block, event)

	tx.Commit(ctx)
	tx, _ = p.repository.Db.Begin(ctx)
	qtx = p.repository.WithTx(tx)

	payload = &protobuf.PurchasePayload{
		DeployID: &protobuf.ActionID{
			Block:   uint64(testBlockHeigh) - 2,
			TxIndex: uint32(testTxIndex) - 2,
		},
		BuyerPublicKey: buyerPubkeyHex,
		TimeOutBlock:   uint64(testBlockHeigh) + 5,
		NodeIDs:        []uint32{12, 13, 14},
		TotalAmountSat: 600,
	}

	payloadBytes, _ = proto.Marshal(payload)
	payloadHash = sha256.Sum256(payloadBytes)
	signature = ecdsa.Sign(sellerPrivateKey, payloadHash[:])
	signatureHex = hex.EncodeToString(signature.Serialize())

	message = &protobuf.NodeSaleEvent{
		Action: protobuf.Action_ACTION_PURCHASE,
		Purchase: &protobuf.ActionPurchase{
			Payload:         payload,
			SellerSignature: signatureHex,
		},
	}

	event, block = assembleTestEvent(buyerPrivateKey, "10101010", "10101010", 0, 0, message)

	addr, _ = btcutil.NewAddressPubKey(sellerPrivateKey.PubKey().SerializeCompressed(), p.network.ChainParams())
	pkscript, _ = txscript.PayToAddrScript(addr.AddressPubKeyHash())
	event.transaction.TxOut = []*types.TxOut{
		{
			PkScript: pkscript,
			Value:    600,
		},
	}
	p.processPurchase(ctx, qtx, block, event)
}
