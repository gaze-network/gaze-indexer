package validator

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
)

type Validator struct {
	Valid bool
}

func New() *Validator {
	return &Validator{
		Valid: true,
	}
}

func EqualXonlyPublicKey(valid bool, target string, expected *btcec.PublicKey) (bool, error) {
	if !valid {
		return false, nil
	}
	targetBytes, err := hex.DecodeString(target)
	if err != nil {
		return false, errors.Wrap(err, "cannot decode hexstring")
	}

	targetPubKey, err := btcec.ParsePubKey(targetBytes)
	if err != nil {
		return false, errors.Wrap(err, "cannot parse public key")
	}
	xOnlyTargetPubKey := btcec.ToSerialized(targetPubKey).SchnorrSerialized()
	xOnlyExpectedPubKey := btcec.ToSerialized(expected).SchnorrSerialized()

	return bytes.Equal(xOnlyTargetPubKey[:], xOnlyExpectedPubKey[:]), nil
}

func (v *Validator) EqualXonlyPublicKey(target string, expected *btcec.PublicKey) (bool, error) {
	if !v.Valid {
		return false, nil
	}
	targetBytes, err := hex.DecodeString(target)
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "cannot decode hexstring")
	}

	targetPubKey, err := btcec.ParsePubKey(targetBytes)
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "cannot parse public key")
	}
	xOnlyTargetPubKey := btcec.ToSerialized(targetPubKey).SchnorrSerialized()
	xOnlyExpectedPubKey := btcec.ToSerialized(expected).SchnorrSerialized()

	v.Valid = bytes.Equal(xOnlyTargetPubKey[:], xOnlyExpectedPubKey[:])
	return v.Valid, nil
}

func (v *Validator) NodesExist(
	ctx context.Context,
	qtx datagateway.NodesaleDataGatewayWithTx,
	deployId *protobuf.ActionID,
	nodeIds []uint32,
) (bool, error) {
	if !v.Valid {
		return false, nil
	}

	nodeIdsInt32 := make([]int32, len(nodeIds))
	for i, id := range nodeIds {
		nodeIdsInt32[i] = int32(id)
	}
	nodes, err := qtx.GetNodes(ctx, datagateway.GetNodesParams{
		SaleBlock:   int64(deployId.Block),
		SaleTxIndex: int32(deployId.TxIndex),
		NodeIds:     nodeIdsInt32,
	})
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "Failed to get nodes")
	}

	if len(nodeIds) != len(nodes) {
		v.Valid = false
		return v.Valid, nil
	}

	v.Valid = true
	return v.Valid, nil
}
