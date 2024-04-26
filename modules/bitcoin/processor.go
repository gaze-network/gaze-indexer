package bitcoin

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/modules/bitcoin/datagateway"
)

// Make sure to implement the BitcoinProcessor interface
var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	config        config.Config
	bitcoinDg     datagateway.BitcoinDataGateway
	indexerInfoDg datagateway.IndexerInformationDataGateway
}

func NewProcessor(config config.Config, bitcoinDg datagateway.BitcoinDataGateway, indexerInfoDg datagateway.IndexerInformationDataGateway) *Processor {
	return &Processor{
		config:        config,
		bitcoinDg:     bitcoinDg,
		indexerInfoDg: indexerInfoDg,
	}
}

func (p Processor) Name() string {
	return "bitcoin"
}

func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	if len(inputs) == 0 {
		return nil
	}

	// Process the given blocks before inserting to the database
	inputs, err := p.process(ctx, inputs)
	if err != nil {
		return errors.WithStack(err)
	}

	// Insert blocks
	if err := p.bitcoinDg.InsertBlocks(ctx, inputs); err != nil {
		return errors.Wrapf(err, "error during insert blocks, from: %d, to: %d", inputs[0].Header.Height, inputs[len(inputs)-1].Header.Height)
	}

	return nil
}

func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	b, err := p.bitcoinDg.GetLatestBlockHeader(ctx)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return defaultCurrentBlock, nil
		}
		return types.BlockHeader{}, errors.WithStack(err)
	}
	return b, nil
}

func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	header, err := p.bitcoinDg.GetBlockHeaderByHeight(ctx, height)
	if err != nil {
		return types.BlockHeader{}, errors.WithStack(err)
	}
	return header, nil
}

func (p *Processor) RevertData(ctx context.Context, from int64) error {
	// to prevent remove txin/txout of duplicated coinbase transaction in the blocks 91842 and 91880
	// if you really want to revert the data before the block `227835`, you should reset the database and reindex the data instead.
	if from <= lastV1Block.Height {
		return errors.Wrapf(errs.InvalidArgument, "can't revert data before block version 2, height: %d", lastV1Block.Height)
	}

	if err := p.bitcoinDg.RevertBlocks(ctx, from); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *Processor) VerifyStates(ctx context.Context) error {
	// Check current db version with the required db version
	{
		dbVersion, err := p.indexerInfoDg.GetCurrentDBVersion(ctx)
		if err != nil {
			return errors.Wrap(err, "can't get current db version")
		}

		if dbVersion != DBVersion {
			return errors.Wrapf(errs.ConflictSetting, "db version mismatch, please upgrade to version %d", DBVersion)
		}
	}

	// Check if the latest indexed network is mismatched with configured network
	{
		_, network, err := p.indexerInfoDg.GetLatestIndexerStats(ctx)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				goto end
			}
			return errors.Wrap(err, "can't get latest indexer stats")
		}

		if network != p.config.Network {
			return errors.Wrapf(errs.ConflictSetting, "network mismatch, latest indexed network: %q, configured network: %q. If you want to change the network, please reset the database", network, p.config.Network)
		}
	}

	// TODO: Verify the states of the indexed data to ensure the last shutdown was graceful and no missing data.

end:
	if err := p.indexerInfoDg.UpdateIndexerStats(ctx, Version, p.config.Network); err != nil {
		return errors.Wrap(err, "can't update indexer stats")
	}

	return nil
}
