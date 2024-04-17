package usecase

import (
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
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
