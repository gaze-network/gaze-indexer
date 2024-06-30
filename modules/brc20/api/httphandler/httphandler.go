package httphandler

import (
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/usecase"
)

type HttpHandler struct {
	usecase *usecase.Usecase
	network common.Network
}

func New(network common.Network, usecase *usecase.Usecase) *HttpHandler {
	return &HttpHandler{
		network: network,
		usecase: usecase,
	}
}
