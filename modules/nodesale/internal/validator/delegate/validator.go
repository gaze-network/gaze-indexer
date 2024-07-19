package delegate

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/validator"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/samber/lo"
)

type DelegateValidator struct {
	validator.Validator
}

func New() *DelegateValidator {
	return &DelegateValidator{
		Validator: validator.Validator{Valid: true},
	}
}

func (v *DelegateValidator) NodesExist(
	ctx context.Context,
	qtx datagateway.NodesaleDataGatewayWithTx,
	deployId *protobuf.ActionID,
	nodeIds []uint32,
) (bool, []entity.Node, error) {
	if !v.Valid {
		return false, nil, nil
	}

	nodes, err := qtx.GetNodes(ctx, datagateway.GetNodesParams{
		SaleBlock:   int64(deployId.Block),
		SaleTxIndex: int32(deployId.TxIndex),
		NodeIds:     lo.Map(nodeIds, func(item uint32, index int) int32 { return int32(item) }),
	})
	if err != nil {
		v.Valid = false
		return v.Valid, nil, errors.Wrap(err, "Failed to get nodes")
	}

	if len(nodeIds) != len(nodes) {
		v.Valid = false
		return v.Valid, nil, nil
	}

	v.Valid = true
	return v.Valid, nodes, nil
}
