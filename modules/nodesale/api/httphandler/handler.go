package httphandler

import (
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
)

type handler struct {
	nodeSaleDg datagateway.NodeSaleDataGateway
}

func New(datagateway datagateway.NodeSaleDataGateway) *handler {
	h := handler{}
	h.nodeSaleDg = datagateway
	return &h
}
