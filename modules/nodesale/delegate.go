package nodesale

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
)

func (p *Processor) processDelegate(ctx context.Context, qtx datagateway.NodesaleDataGatewayWithTx, block *types.Block, event nodesaleEvent) error {
	valid := true
	delegate := event.eventMessage.Delegate
	nodeIds := make([]int32, len(delegate.NodeIDs))
	for i, id := range delegate.NodeIDs {
		nodeIds[i] = int32(id)
	}
	nodes, err := qtx.GetNodes(ctx, datagateway.GetNodesParams{
		SaleBlock:   int64(delegate.DeployID.Block),
		SaleTxIndex: int32(delegate.DeployID.TxIndex),
		NodeIds:     nodeIds,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to get nodes")
	}

	if len(nodeIds) != len(nodes) {
		valid = false
	}

	if valid {
		for _, node := range nodes {
			OwnerPublicKeyBytes, err := hex.DecodeString(node.OwnerPublicKey)
			if err != nil {
				valid = false
				break
			}
			OwnerPublicKey, err := btcec.ParsePubKey(OwnerPublicKeyBytes)
			if err != nil {
				valid = false
				break
			}
			xOnlyOwnerPublicKey := btcec.ToSerialized(OwnerPublicKey).SchnorrSerialized()
			xOnlyTxPubKey := btcec.ToSerialized(event.txPubkey).SchnorrSerialized()
			if !bytes.Equal(xOnlyOwnerPublicKey[:], xOnlyTxPubKey[:]) {
				valid = false
				break
			}
		}
	}

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
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}

	if valid {
		_, err = qtx.SetDelegates(ctx, datagateway.SetDelegatesParams{
			SaleBlock:   int64(delegate.DeployID.Block),
			SaleTxIndex: int32(delegate.DeployID.TxIndex),
			Delegatee:   delegate.DelegateePublicKey,
			NodeIds:     nodeIds,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to set delegate")
		}
	}

	return nil
}
