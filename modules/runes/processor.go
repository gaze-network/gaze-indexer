package runes

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

const (
	dbVersion        = 1
	eventHashVersion = 1
)

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

var (
	ErrDBVersionMismatch        = errors.New("db version mismatch: please migrate db")
	ErrEventHashVersionMismatch = errors.New("event hash version mismatch: please reset db and reindex")
)

func (p *Processor) Init(ctx context.Context) error {
	// TODO: ensure db is migrated
	if err := p.ensureValidState(ctx); err != nil {
		return errors.Wrap(err, "error during ensureValidState")
	}
	if p.network == common.NetworkMainnet {
		if err := p.ensureGenesisRune(ctx); err != nil {
			return errors.Wrap(err, "error during ensureGenesisRune")
		}
	}
	return nil
}

func (p *Processor) ensureValidState(ctx context.Context) error {
	indexerState, err := p.runesDg.GetLatestIndexerState(ctx)
	if err != nil && !errors.Is(err, errs.NotFound) {
		return errors.Wrap(err, "failed to get latest indexer state")
	}
	// if not found, set indexer state
	if errors.Is(err, errs.NotFound) {
		if err := p.runesDg.SetIndexerState(ctx, entity.IndexerState{
			DBVersion:        dbVersion,
			EventHashVersion: eventHashVersion,
		}); err != nil {
			return errors.Wrap(err, "failed to set indexer state")
		}
	} else {
		if indexerState.DBVersion != dbVersion {
			return errors.WithStack(ErrDBVersionMismatch)
		}
		if indexerState.EventHashVersion != eventHashVersion {
			return errors.WithStack(ErrEventHashVersionMismatch)
		}
	}
	return nil
}

var genesisRuneId = runes.RuneId{BlockHeight: 1, TxIndex: 0}

func (p *Processor) ensureGenesisRune(ctx context.Context) error {
	_, err := p.runesDg.GetRuneEntryByRuneId(ctx, genesisRuneId)
	if err != nil && !errors.Is(err, errs.NotFound) {
		return errors.Wrap(err, "failed to get genesis rune entry")
	}
	if errors.Is(err, errs.NotFound) {
		runeEntry := &runes.RuneEntry{
			RuneId:       genesisRuneId,
			Divisibility: 0,
			Premine:      uint128.Zero,
			SpacedRune:   runes.NewSpacedRune(runes.NewRune(2055900680524219742), 0b10000000),
			Symbol:       '\u29c9',
			Terms: &runes.Terms{
				Amount:      lo.ToPtr(uint128.From64(1)),
				Cap:         &uint128.Max,
				HeightStart: lo.ToPtr(uint64(common.HalvingInterval * 4)),
				HeightEnd:   lo.ToPtr(uint64(common.HalvingInterval * 5)),
				OffsetStart: nil,
				OffsetEnd:   nil,
			},
			Turbo:          true,
			Mints:          uint128.Zero,
			BurnedAmount:   uint128.Zero,
			CompletionTime: time.Time{},
		}
		if err := p.runesDg.CreateRuneEntry(ctx, runeEntry, genesisRuneId.BlockHeight); err != nil {
			return errors.Wrap(err, "failed to create genesis rune entry")
		}
	}
	return nil
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
