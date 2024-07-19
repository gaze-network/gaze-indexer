package purchasevalidator

import (
	"bytes"
	"context"
	"encoding/hex"
	"slices"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/validator"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type PurchaseValidator struct {
	validator.Validator
}

func New() *PurchaseValidator {
	return &PurchaseValidator{
		Validator: validator.Validator{Valid: true},
	}
}

func (v *PurchaseValidator) NodeSaleExists(ctx context.Context, qtx datagateway.NodesaleDataGatewayWithTx, payload *protobuf.PurchasePayload) (bool, *entity.NodeSale, error) {
	if !v.Valid {
		return false, nil, nil
	}
	// check node existed
	deploys, err := qtx.GetNodesale(ctx, datagateway.GetNodesaleParams{
		BlockHeight: int64(payload.DeployID.Block),
		TxIndex:     int32(payload.DeployID.TxIndex),
	})
	if err != nil {
		v.Valid = false
		return v.Valid, nil, errors.Wrap(err, "Failed to Get nodesale")
	}
	if len(deploys) < 1 {
		v.Valid = false
		return v.Valid, nil, nil
	} else {
		v.Valid = true
		return v.Valid, &deploys[0], nil
	}
}

func (v *PurchaseValidator) ValidTimestamp(deploy *entity.NodeSale, timestamp time.Time) bool {
	if !v.Valid {
		return false
	}
	if timestamp.Before(deploy.StartsAt) ||
		timestamp.After(deploy.EndsAt) {
		v.Valid = false
		return v.Valid
	}
	v.Valid = true
	return v.Valid
}

func (v *PurchaseValidator) WithinTimeoutBlock(payload *protobuf.PurchasePayload, blockHeight uint64) bool {
	if !v.Valid {
		return false
	}
	if payload.TimeOutBlock < blockHeight {
		v.Valid = false
		return v.Valid
	}
	v.Valid = true
	return v.Valid
}

func (v *PurchaseValidator) VerifySignature(purchase *protobuf.ActionPurchase, deploy *entity.NodeSale) (bool, error) {
	if !v.Valid {
		return false, nil
	}
	payload := purchase.Payload
	payloadBytes, _ := proto.Marshal(payload)
	signatureBytes, _ := hex.DecodeString(purchase.SellerSignature)
	signature, err := ecdsa.ParseSignature(signatureBytes)
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "cannot parse signature")
	}
	hash := chainhash.DoubleHashB(payloadBytes)
	pubkeyBytes, _ := hex.DecodeString(deploy.SellerPublicKey)
	pubKey, _ := btcec.ParsePubKey(pubkeyBytes)
	verified := signature.Verify(hash[:], pubKey)
	if !verified {
		v.Valid = false
		return v.Valid, nil
	}
	v.Valid = true
	return v.Valid, nil
}

func (v *PurchaseValidator) ValidTiers(
	payload *protobuf.PurchasePayload,
	deploy *entity.NodeSale,
) (bool, []protobuf.Tier, []uint32, map[uint32]int32, error) {
	if !v.Valid {
		return false, nil, nil, nil, nil
	}
	tiers := make([]protobuf.Tier, len(deploy.Tiers))
	buyingTiersCount := make([]uint32, len(tiers))
	nodeIdToTier := make(map[uint32]int32)

	for i, tierJson := range deploy.Tiers {
		tier := &tiers[i]
		err := protojson.Unmarshal(tierJson, tier)
		if err != nil {
			v.Valid = false
			return v.Valid, nil, nil, nil, errors.Wrap(err, "Failed to decode tiers json")
		}
	}

	slices.Sort(payload.NodeIDs)

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
			v.Valid = false
			return false, nil, nil, nil, nil
		}
	}
	v.Valid = true
	return v.Valid, tiers, buyingTiersCount, nodeIdToTier, nil
}

func (v *PurchaseValidator) ValidUnpurchasedNodes(
	ctx context.Context,
	qtx datagateway.NodesaleDataGatewayWithTx,
	payload *protobuf.PurchasePayload,
) (bool, error) {
	if !v.Valid {
		return false, nil
	}

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
		v.Valid = false
		return v.Valid, errors.Wrap(err, "Failed to Get nodes")
	}
	if len(nodes) > 0 {
		v.Valid = false
		return false, nil
	}
	v.Valid = true
	return true, nil
}

func (v *PurchaseValidator) ValidPaidAmount(
	payload *protobuf.PurchasePayload,
	deploy *entity.NodeSale,
	txOuts []*types.TxOut,
	tiers []protobuf.Tier,
	buyingTiersCount []uint32,
	network *chaincfg.Params,
) (bool, *entity.MetaData, error) {
	if !v.Valid {
		return false, nil, nil
	}
	sellerAddr, err := btcutil.DecodeAddress(deploy.SellerWallet, network) // default to mainnet
	if err != nil {
		v.Valid = false
		return v.Valid, nil, errors.Wrap(err, "Cannot decode Sellerwallet")
	}

	var txPaid int64 = 0
	meta := entity.MetaData{}

	// get total amount paid to seller
	for _, txOut := range txOuts {
		_, txOutAddrs, _, _ := txscript.ExtractPkScriptAddrs(txOut.PkScript, network)

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
		v.Valid = false
		return v.Valid, nil, nil
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
		v.Valid = false
		return v.Valid, nil, nil
	}
	v.Valid = true
	return v.Valid, &meta, nil
}

func (v *PurchaseValidator) WithinLimit(
	ctx context.Context,
	qtx datagateway.NodesaleDataGatewayWithTx,
	payload *protobuf.PurchasePayload,
	deploy *entity.NodeSale,
	tiers []protobuf.Tier,
	buyingTiersCount []uint32,
) (bool, error) {
	if !v.Valid {
		return false, nil
	}

	// check node limit
	// get all selled by seller and owned by buyer
	buyerOwnedNodes, err := qtx.GetNodesByOwner(ctx, datagateway.GetNodesByOwnerParams{
		SaleBlock:      deploy.BlockHeight,
		SaleTxIndex:    deploy.TxIndex,
		OwnerPublicKey: payload.BuyerPublicKey,
	})
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "Failed to GetNodesByOwner")
	}
	if len(buyerOwnedNodes)+len(payload.NodeIDs) > int(deploy.MaxPerAddress) {
		v.Valid = false
		return v.Valid, nil
	}

	// check limit
	// count each tiers
	// check limited for each tier
	ownedTiersCount := make([]uint32, len(tiers))
	for _, node := range buyerOwnedNodes {
		ownedTiersCount[node.TierIndex]++
	}
	for i := 0; i < len(tiers); i++ {
		if ownedTiersCount[i]+buyingTiersCount[i] > tiers[i].MaxPerAddress {
			v.Valid = false
			return v.Valid, nil
		}
	}
	v.Valid = true
	return v.Valid, nil
}
