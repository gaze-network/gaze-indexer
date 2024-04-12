package runes

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	runesDg           datagateway.RunesDataGateway
	bitcoinClient     btcclient.Contract
	bitcoinDataSource indexers.BitcoinDatasource
	network           common.Network
}

type NewProcessorParams struct {
	RunesDg           datagateway.RunesDataGateway
	BitcoinClient     btcclient.Contract
	BitcoinDataSource indexers.BitcoinDatasource
	Network           common.Network
}

func NewProcessor(params NewProcessorParams) *Processor {
	return &Processor{
		runesDg:           params.RunesDg,
		bitcoinClient:     params.BitcoinClient,
		bitcoinDataSource: params.BitcoinDataSource,
		network:           params.Network,
	}
}

func (p *Processor) Name() string {
	return "Runes"
}

func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	blockHeader, err := p.runesDg.GetLatestBlock(ctx)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get latest block")
	}
	return blockHeader, nil
}

// warning: GetIndexedBlock currently returns a types.BlockHeader with only Height, Hash fields populated.
// This is because it is known that all usage of this function only requires these fields. In the future, we may want to populate all fields for type safety.
func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	block, err := p.runesDg.GetIndexedBlockByHeight(ctx, height)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get indexed block")
	}
	return types.BlockHeader{
		Height: block.Height,
		Hash:   block.Hash,
	}, nil
}

func (p *Processor) RevertData(ctx context.Context, from int64) error {
	panic("implement me")
}
