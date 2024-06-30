package usecase

import (
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
)

type Usecase struct {
	dg            datagateway.BRC20DataGateway
	bitcoinClient btcclient.Contract
}

func New(dg datagateway.BRC20DataGateway, bitcoinClient btcclient.Contract) *Usecase {
	return &Usecase{
		dg:            dg,
		bitcoinClient: bitcoinClient,
	}
}
