package nodesale

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	delegatevalidator "github.com/gaze-network/indexer-network/modules/nodesale/internal/validator/delegate"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/samber/lo"
)

func (p *Processor) processDelegate(ctx context.Context, qtx datagateway.NodeSaleDataGatewayWithTx, block *types.Block, event nodeSaleEvent) error {
	validator := delegatevalidator.New()
	delegate := event.eventMessage.Delegate

	_, nodes, err := validator.NodesExist(ctx, qtx, delegate.DeployID, delegate.NodeIDs)
	if err != nil {
		return errors.Wrap(err, "Cannot query")
	}

	for _, node := range nodes {
		valid, err := validator.EqualXonlyPublicKey(node.OwnerPublicKey, event.txPubkey)
		if err != nil {
			logger.DebugContext(ctx, "Invalid public key", slogx.Error(err))
		}
		if !valid {
			break
		}
	}

	err = qtx.CreateEvent(ctx, entity.NodeSaleEvent{
		TxHash:         event.transaction.TxHash.String(),
		TxIndex:        int32(event.transaction.Index),
		Action:         int32(event.eventMessage.Action),
		RawMessage:     event.rawData,
		ParsedMessage:  event.eventJson,
		BlockTimestamp: block.Header.Timestamp,
		BlockHash:      event.transaction.BlockHash.String(),
		BlockHeight:    event.transaction.BlockHeight,
		Valid:          validator.Valid,
		WalletAddress:  p.pubkeyToPkHashAddress(event.txPubkey).EncodeAddress(),
		Metadata:       nil,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to insert event")
	}

	if validator.Valid {
		nodeIds := lo.Map(delegate.NodeIDs, func(item uint32, index int) int32 { return int32(item) })
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
