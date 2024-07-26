package delegate

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/validator"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
)

type DelegateValidator struct {
	validator.Validator
}

func New() *DelegateValidator {
	v := validator.New()
	return &DelegateValidator{
		Validator: *v,
	}
}

func (v *DelegateValidator) NodesExist(
	ctx context.Context,
	qtx datagateway.NodeSaleDataGatewayWithTx,
	deployId *protobuf.ActionID,
	nodeIds []uint32,
) (bool, []entity.Node, error) {
	if !v.Valid {
		return false, nil, nil
	}

	nodes, err := qtx.GetNodesByIds(ctx, datagateway.GetNodesByIdsParams{
		SaleBlock:   deployId.Block,
		SaleTxIndex: deployId.TxIndex,
		NodeIds:     nodeIds,
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
