package nodesale

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Processor) processDelegate(ctx context.Context, qtx gen.Querier, block *types.Block, event nodesaleEvent) error {
	valid := true
	delegate := event.eventMessage.Delegate
	nodeIds := make([]int32, len(delegate.NodeIDs))
	for i, id := range delegate.NodeIDs {
		nodeIds[i] = int32(id)
	}
	nodes, err := qtx.GetNodes(ctx, gen.GetNodesParams{
		SaleBlock:   int32(delegate.DeployID.Block),
		SaleTxIndex: int32(delegate.DeployID.TxIndex),
		NodeIds:     nodeIds,
	})
	if err != nil {
		return fmt.Errorf("Failed to get nodes : %w", err)
	}

	if len(nodeIds) != len(nodes) {
		fmt.Println(nodeIds)
		fmt.Println(nodes)
		valid = false
	}

	if valid {
		for _, node := range nodes {
			ownerAddress, err := p.pubkeyToTaprootAddress(node.OwnerPublicKey, event.rawScript)
			if err != nil || !bytes.Equal(
				[]byte(ownerAddress.EncodeAddress()),
				[]byte(event.txAddress.EncodeAddress()),
			) {
				valid = false
			}
		}
	}

	err = qtx.AddEvent(ctx, gen.AddEventParams{
		TxHash:         event.transaction.TxHash.String(),
		TxIndex:        int32(event.transaction.Index),
		Action:         int32(event.eventMessage.Action),
		RawMessage:     event.rawData,
		ParsedMessage:  event.eventJson,
		BlockTimestamp: pgtype.Timestamp{Time: block.Header.Timestamp, Valid: true},
		BlockHash:      block.Header.Hash.String(),
		BlockHeight:    int32(block.Header.Height),
		Valid:          valid,
		WalletAddress:  event.txAddress.EncodeAddress(),
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return fmt.Errorf("Failed to insert event : %w", err)
	}
	if valid {
		_, err = qtx.SetDelegates(ctx, gen.SetDelegatesParams{
			SaleBlock:   int32(delegate.DeployID.Block),
			SaleTxIndex: int32(delegate.DeployID.TxIndex),
			Delegatee: pgtype.Text{
				String: delegate.DelegateePublicKey,
				Valid:  true,
			},
			NodeIds: nodeIds,
		})
		if err != nil {
			return fmt.Errorf("Failed to set delegate : %w", err)
		}
	}

	return nil
}
