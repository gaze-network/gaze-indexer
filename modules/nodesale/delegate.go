package nodesale

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	delegatevalidator "github.com/gaze-network/indexer-network/modules/nodesale/internal/validator/delegate"
	"github.com/samber/lo"
)

func (p *Processor) ProcessDelegate(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, block *types.Block, event NodeSaleEvent) error {
	validator := delegatevalidator.New()
	delegate := event.EventMessage.Delegate

	_, nodes, err := validator.NodesExist(ctx, qtx, delegate.DeployID, delegate.NodeIDs)
	if err != nil {
		return errors.Wrap(err, "Cannot query")
	}

	for _, node := range nodes {
		valid := validator.EqualXonlyPublicKey(node.OwnerPublicKey, event.TxPubkey)
		if !valid {
			break
		}
	}

	err = qtx.CreateEvent(ctx, entity.NodeSaleEvent{
		TxHash:         event.Transaction.TxHash.String(),
		TxIndex:        int32(event.Transaction.Index),
		Action:         int32(event.EventMessage.Action),
		RawMessage:     event.RawData,
		ParsedMessage:  event.EventJson,
		BlockTimestamp: block.Header.Timestamp,
		BlockHash:      event.Transaction.BlockHash.String(),
		BlockHeight:    event.Transaction.BlockHeight,
		Valid:          validator.Valid,
		WalletAddress:  p.PubkeyToPkHashAddress(event.TxPubkey).EncodeAddress(),
		Metadata:       nil,
		Reason:         validator.Reason,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}

	if validator.Valid {
		nodeIds := lo.Map(delegate.NodeIDs, func(item uint32, index int) int32 { return int32(item) })
		_, err = qtx.SetDelegates(ctx, datagateway.SetDelegatesParams{
			SaleBlock:      int64(delegate.DeployID.Block),
			SaleTxIndex:    int32(delegate.DeployID.TxIndex),
			Delegatee:      delegate.DelegateePublicKey,
			DelegateTxHash: event.Transaction.TxHash.String(),
			NodeIds:        nodeIds,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to set delegate")
		}
	}

	return nil
}
