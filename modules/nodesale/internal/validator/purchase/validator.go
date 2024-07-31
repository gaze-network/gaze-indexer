package purchase

import (
	"context"
	"encoding/hex"
	"slices"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
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
	v := validator.New()
	return &PurchaseValidator{
		Validator: *v,
	}
}

func (v *PurchaseValidator) NodeSaleExists(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, payload *protobuf.PurchasePayload) (bool, *entity.NodeSale, error) {
	if !v.Valid {
		return false, nil, nil
	}
	// check node existed
	deploys, err := qtx.GetNodeSale(ctx, datagateway.GetNodeSaleParams{
		BlockHeight: payload.DeployID.Block,
		TxIndex:     payload.DeployID.TxIndex,
	})
	if err != nil {
		v.Valid = false
		return v.Valid, nil, errors.Wrap(err, "Failed to Get NodeSale")
	}
	if len(deploys) < 1 {
		v.Valid = false
		v.Reason = DEPLOYID_NOT_FOUND
		return v.Valid, nil, nil
	}
	v.Valid = true
	return v.Valid, &deploys[0], nil
}

func (v *PurchaseValidator) ValidTimestamp(deploy *entity.NodeSale, timestamp time.Time) bool {
	if !v.Valid {
		return false
	}
	if timestamp.Before(deploy.StartsAt) ||
		timestamp.After(deploy.EndsAt) {
		v.Valid = false
		v.Reason = PURCHASE_TIMEOUT
		return v.Valid
	}
	v.Valid = true
	return v.Valid
}

func (v *PurchaseValidator) WithinTimeoutBlock(timeOutBlock uint64, blockHeight uint64) bool {
	if !v.Valid {
		return false
	}
	if timeOutBlock == 0 {
		// No timeout
		v.Valid = true
		return v.Valid
	}
	if timeOutBlock < blockHeight {
		v.Valid = false
		v.Reason = BLOCK_HEIGHT_TIMEOUT
		return v.Valid
	}
	v.Valid = true
	return v.Valid
}

func (v *PurchaseValidator) VerifySignature(purchase *protobuf.ActionPurchase, deploy *entity.NodeSale) bool {
	if !v.Valid {
		return false
	}
	payload := purchase.Payload
	payloadBytes, _ := proto.Marshal(payload)
	signatureBytes, _ := hex.DecodeString(purchase.SellerSignature)
	signature, err := ecdsa.ParseSignature(signatureBytes)
	if err != nil {
		v.Valid = false
		v.Reason = INVALID_SIGNATURE_FORMAT
		return v.Valid
	}
	hash := chainhash.DoubleHashB(payloadBytes)
	pubkeyBytes, _ := hex.DecodeString(deploy.SellerPublicKey)
	pubKey, _ := btcec.ParsePubKey(pubkeyBytes)
	verified := signature.Verify(hash[:], pubKey)
	if !verified {
		v.Valid = false
		v.Reason = INVALID_SIGNATURE
		return v.Valid
	}
	v.Valid = true
	return v.Valid
}

type TierMap struct {
	Tiers            []protobuf.Tier
	BuyingTiersCount []uint32
	NodeIdToTier     map[uint32]int32
}

func (v *PurchaseValidator) ValidTiers(
	payload *protobuf.PurchasePayload,
	deploy *entity.NodeSale,
) (bool, TierMap) {
	if !v.Valid {
		return false, TierMap{}
	}
	tiers := make([]protobuf.Tier, len(deploy.Tiers))
	buyingTiersCount := make([]uint32, len(tiers))
	nodeIdToTier := make(map[uint32]int32)

	for i, tierJson := range deploy.Tiers {
		tier := &tiers[i]
		err := protojson.Unmarshal(tierJson, tier)
		if err != nil {
			v.Valid = false
			v.Reason = INVALID_TIER_JSON
			return v.Valid, TierMap{}
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
			v.Reason = INVALID_NODE_ID
			return false, TierMap{}
		}
	}
	v.Valid = true
	return v.Valid, TierMap{
		Tiers:            tiers,
		BuyingTiersCount: buyingTiersCount,
		NodeIdToTier:     nodeIdToTier,
	}
}

func (v *PurchaseValidator) ValidUnpurchasedNodes(
	ctx context.Context,
	qtx datagateway.NodeSaleDataGatewayWithTx,
	payload *protobuf.PurchasePayload,
) (bool, error) {
	if !v.Valid {
		return false, nil
	}

	// valid unpurchased node ID
	nodes, err := qtx.GetNodesByIds(ctx, datagateway.GetNodesByIdsParams{
		SaleBlock:   payload.DeployID.Block,
		SaleTxIndex: payload.DeployID.TxIndex,
		NodeIds:     payload.NodeIDs,
	})
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "Failed to Get nodes")
	}
	if len(nodes) > 0 {
		v.Valid = false
		v.Reason = NODE_ALREADY_PURCHASED
		return false, nil
	}
	v.Valid = true
	return true, nil
}

func (v *PurchaseValidator) ValidPaidAmount(
	payload *protobuf.PurchasePayload,
	deploy *entity.NodeSale,
	txPaid uint64,
	tiers []protobuf.Tier,
	buyingTiersCount []uint32,
	network *chaincfg.Params,
) (bool, *entity.MetadataEventPurchase) {
	if !v.Valid {
		return false, nil
	}

	meta := entity.MetadataEventPurchase{}

	meta.PaidTotalAmount = txPaid
	meta.ReportedTotalAmount = uint64(payload.TotalAmountSat)
	// total amount paid is greater than report paid
	if txPaid < uint64(payload.TotalAmountSat) {
		v.Valid = false
		v.Reason = INVALID_PAYMENT
		return v.Valid, nil
	}
	// calculate total price
	var totalPrice uint64 = 0
	for i := 0; i < len(tiers); i++ {
		totalPrice += uint64(buyingTiersCount[i] * tiers[i].PriceSat)
	}
	// report paid is greater than max discounted total price
	maxDiscounted := totalPrice * (100 - uint64(deploy.MaxDiscountPercentage))
	decimal := maxDiscounted % 100
	maxDiscounted /= 100
	if decimal%100 >= 50 {
		maxDiscounted++
	}
	meta.ExpectedTotalAmountDiscounted = maxDiscounted
	if uint64(payload.TotalAmountSat) < maxDiscounted {
		v.Valid = false
		v.Reason = INSUFFICIENT_FUND
		return v.Valid, nil
	}
	v.Valid = true
	return v.Valid, &meta
}

func (v *PurchaseValidator) WithinLimit(
	ctx context.Context,
	qtx datagateway.NodeSaleDataGatewayWithTx,
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
		v.Reason = "Purchase over limit per address."
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
			v.Reason = "Purchase over limit per tier."
			return v.Valid, nil
		}
	}
	v.Valid = true
	return v.Valid, nil
}
