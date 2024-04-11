package runes

import (
	"context"

	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	runesDg       datagateway.RunesDataGateway
	bitcoinClient btcclient.Contract
	network       common.Network
}

type NewProcessorParams struct {
	RunesDg       datagateway.RunesDataGateway
	BitcoinClient btcclient.Contract
	Network       common.Network
}

func NewProcessor(params NewProcessorParams) *Processor {
	return &Processor{
		runesDg:       params.RunesDg,
		bitcoinClient: params.BitcoinClient,
		network:       params.Network,
	}
}

func (p *Processor) Name() string {
	return "Runes"
}

func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	panic("implement me")
}

func (p *Processor) PrepareData(ctx context.Context, from, to int64) ([]*types.Block, error) {
	panic("implement me")
}

func (p *Processor) RevertData(ctx context.Context, from int64) error {
	panic("implement me")
}
