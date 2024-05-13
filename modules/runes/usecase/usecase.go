package usecase

import (
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
)

type Usecase struct {
	runesDg       datagateway.RunesDataGateway
	bitcoinClient btcclient.Contract
}

func New(runesDg datagateway.RunesDataGateway, bitcoinClient btcclient.Contract) *Usecase {
	return &Usecase{
		runesDg:       runesDg,
		bitcoinClient: bitcoinClient,
	}
}
