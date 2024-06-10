package api

import (
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/runes/api/httphandler"
	"github.com/gaze-network/indexer-network/modules/runes/usecase"
)

func NewHTTPHandler(network common.Network, usecase *usecase.Usecase) *httphandler.HttpHandler {
	return httphandler.New(network, usecase)
}
