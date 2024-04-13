package runes

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"golang.org/x/sync/errgroup"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	runesDg           datagateway.RunesDataGateway
	bitcoinClient     btcclient.Contract
	bitcoinDataSource indexers.BitcoinDatasource
	network           common.Network

	newRuneEntryStates map[runes.RuneId]*runes.RuneEntry
}

type NewProcessorParams struct {
	RunesDg           datagateway.RunesDataGateway
	BitcoinClient     btcclient.Contract
	BitcoinDataSource indexers.BitcoinDatasource
	Network           common.Network
}

func NewProcessor(params NewProcessorParams) *Processor {
	return &Processor{
		runesDg:            params.RunesDg,
		bitcoinClient:      params.BitcoinClient,
		bitcoinDataSource:  params.BitcoinDataSource,
		network:            params.Network,
		newRuneEntryStates: make(map[runes.RuneId]*runes.RuneEntry),
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
	err := p.runesDg.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err := p.runesDg.Rollback(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to rollback transaction", err)
		}
	}()

	sinceHeight := uint64(from + 1)

	eg, ectx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := p.runesDg.DeleteIndexedBlockSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete indexed blocks")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.DeleteRuneEntriesSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete rune entries")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.DeleteRuneEntryStatesSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete rune entry states")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.DeleteRuneTransactionsSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete rune transactions")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.DeleteRunestonesSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete runestones")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.DeleteOutPointBalancesSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete outpoint balances")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.UnspendOutPointBalancesSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to unspend outpoint balances")
		}
		return nil
	})
	eg.Go(func() error {
		if err := p.runesDg.DeleteRuneBalancesSinceHeight(ectx, sinceHeight); err != nil {
			return errors.Wrap(err, "failed to delete rune balances")
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to revert data")
	}

	if err := p.runesDg.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}
