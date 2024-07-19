package nodesale

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"slices"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type metaData struct {
	ExpectedTotalAmountDiscounted int64
	ReportedTotalAmount           int64
	PaidTotalAmount               int64
}

func (p *Processor) processPurchase(ctx context.Context, qtx datagateway.NodesaleDataGatewayWithTx, block *types.Block, event nodesaleEvent) error {
	valid := true
	purchase := event.eventMessage.Purchase
	payload := purchase.Payload

	buyerPubkeyBytes, err := hex.DecodeString(payload.BuyerPublicKey)
	if err != nil {
		valid = false
	}

	if valid {
		buyerPubkey, err := btcec.ParsePubKey(buyerPubkeyBytes)
		if err != nil {
			valid = false
		}
		xOnlyBuyerPubkey := btcec.ToSerialized(buyerPubkey).SchnorrSerialized()
		xOnlyTxPubKey := btcec.ToSerialized(event.txPubkey).SchnorrSerialized()

		if valid && !bytes.Equal(xOnlyBuyerPubkey[:], xOnlyTxPubKey[:]) {
			valid = false
		}
	}

	var deploy *entity.NodeSale
	if valid {
		// check node existed
		deploys, err := qtx.GetNodesale(ctx, datagateway.GetNodesaleParams{
			BlockHeight: int64(payload.DeployID.Block),
			TxIndex:     int32(payload.DeployID.TxIndex),
		})
		if err != nil {
			return errors.Wrap(err, "Failed to Get nodesale")
		}
		if len(deploys) < 1 {
			valid = false
		} else {
			deploy = &deploys[0]
		}
	}

	if valid {
		// check timestamp
		timestamp := block.Header.Timestamp
		if timestamp.Before(deploy.StartsAt) ||
			timestamp.After(deploy.EndsAt) {
			valid = false
		}
	}

	if valid {
		if payload.TimeOutBlock < uint64(event.transaction.BlockHeight) {
			valid = false
		}
	}

	if valid {
		// verified signature
		payloadBytes, _ := proto.Marshal(payload)
		signatureBytes, _ := hex.DecodeString(purchase.SellerSignature)
		signature, err := ecdsa.ParseSignature(signatureBytes)
		if err != nil {
			valid = false
		}
		if valid {
			hash := chainhash.DoubleHashB(payloadBytes)
			pubkeyBytes, _ := hex.DecodeString(deploy.SellerPublicKey)
			pubKey, _ := btcec.ParsePubKey(pubkeyBytes)
			verified := signature.Verify(hash[:], pubKey)
			if !verified {
				valid = false
			}
		}
	}

	var tiers []protobuf.Tier
	var buyingTiersCount []uint32
	nodeIdToTier := make(map[uint32]int32, 1)
	if valid {
		// valid nodeID tier
		tiers = make([]protobuf.Tier, len(deploy.Tiers))
		for i, tierJson := range deploy.Tiers {
			tier := &tiers[i]
			err := protojson.Unmarshal(tierJson, tier)
			if err != nil {
				return errors.Wrap(err, "Failed to decode tiers json")
			}
		}
		slices.Sort(payload.NodeIDs)
		buyingTiersCount = make([]uint32, len(tiers))
		var currentTier int32 = -1
		var tierSum uint32 = 0
		for _, nodeId := range payload.NodeIDs {
			for nodeId >= tierSum && currentTier < int32(len(tiers)-1) {
				currentTier++
				tierSum += tiers[currentTier].Limit
			}
			if nodeId < tierSum {
				buyingTiersCount[currentTier]++
				nodeIdToTier[nodeId] = currentTier
			} else {
				valid = false
			}
		}
	}

	if valid {
		// valid unpurchased node ID
		nodeIds := make([]int32, len(payload.NodeIDs))
		for i, id := range payload.NodeIDs {
			nodeIds[i] = int32(id)
		}
		nodes, err := qtx.GetNodes(ctx, datagateway.GetNodesParams{
			SaleBlock:   int64(payload.DeployID.Block),
			SaleTxIndex: int32(payload.DeployID.TxIndex),
			NodeIds:     nodeIds,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to Get nodes")
		}
		if len(nodes) > 0 {
			valid = false
		}
	}

	var sellerAddr btcutil.Address
	if valid {
		sellerAddr, err = btcutil.DecodeAddress(deploy.SellerWallet, p.network.ChainParams())
		if err != nil {
			valid = false
		}
	}

	var txPaid int64 = 0
	meta := metaData{}

	if valid {
		// get total amount paid to seller
		for _, txOut := range event.transaction.TxOut {
			_, txOutAddrs, _, _ := txscript.ExtractPkScriptAddrs(txOut.PkScript, p.network.ChainParams())

			if len(txOutAddrs) == 1 && bytes.Equal(
				[]byte(sellerAddr.EncodeAddress()),
				[]byte(txOutAddrs[0].EncodeAddress()),
			) {
				txPaid += txOut.Value
			}
		}
		meta.PaidTotalAmount = txPaid
		meta.ReportedTotalAmount = payload.TotalAmountSat
		// total amount paid is greater than report paid
		if txPaid < payload.TotalAmountSat {
			valid = false
		}
		// calculate total price
		var totalPrice int64 = 0
		for i := 0; i < len(tiers); i++ {
			totalPrice += int64(buyingTiersCount[i] * tiers[i].PriceSat)
		}
		// report paid is greater than max discounted total price
		maxDiscounted := totalPrice * (100 - int64(deploy.MaxDiscountPercentage))
		decimal := maxDiscounted % 100
		maxDiscounted /= 100
		if decimal%100 >= 50 {
			maxDiscounted++
		}
		meta.ExpectedTotalAmountDiscounted = maxDiscounted
		if payload.TotalAmountSat < maxDiscounted {
			valid = false
		}
	}

	var buyerOwnedNodes []entity.Node
	if valid {
		var err error
		// check node limit
		// get all selled by seller and owned by buyer
		buyerOwnedNodes, err = qtx.GetNodesByOwner(ctx, datagateway.GetNodesByOwnerParams{
			SaleBlock:      deploy.BlockHeight,
			SaleTxIndex:    deploy.TxIndex,
			OwnerPublicKey: payload.BuyerPublicKey,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to GetNodesByOwner")
		}
		if len(buyerOwnedNodes)+len(payload.NodeIDs) > int(deploy.MaxPerAddress) {
			valid = false
		}
	}

	if valid {
		// check limit
		// count each tiers
		// check limited for each tier
		ownedTiersCount := make([]uint32, len(tiers))
		for _, node := range buyerOwnedNodes {
			ownedTiersCount[node.TierIndex]++
		}
		for i := 0; i < len(tiers); i++ {
			if ownedTiersCount[i]+buyingTiersCount[i] > tiers[i].MaxPerAddress {
				valid = false
				break
			}
		}
	}

	metaDataBytes, _ := json.Marshal(meta)

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
		Metadata:       metaDataBytes,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}

	if valid {
		// add to node
		for _, nodeId := range payload.NodeIDs {
			err := qtx.AddNode(ctx, datagateway.AddNodeParams{
				SaleBlock:      deploy.BlockHeight,
				SaleTxIndex:    deploy.TxIndex,
				NodeID:         int32(nodeId),
				TierIndex:      nodeIdToTier[nodeId],
				DelegatedTo:    "",
				OwnerPublicKey: payload.BuyerPublicKey,
				PurchaseTxHash: event.transaction.TxHash.String(),
				DelegateTxHash: "",
			})
			if err != nil {
				return errors.Wrap(err, "Failed to insert node")
			}
		}
	}

	return nil
}
