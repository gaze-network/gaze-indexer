package usecase

import (
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
)

type Usecase struct {
	runesDg       datagateway.RunesDataGateway
	bitcoinClient btcclient.Contract
}

type NewParams struct {
	RunesDg       datagateway.RunesDataGateway
	BitcoinClient btcclient.Contract
}

func New(params NewParams) *Usecase {
	return &Usecase{
		runesDg:       params.RunesDg,
		bitcoinClient: params.BitcoinClient,
	}
}
