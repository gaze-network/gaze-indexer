package runes

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	runesDg           datagateway.RunesDataGateway
	indexerInfoDg     datagateway.IndexerInfoDataGateway
	bitcoinClient     btcclient.Contract
	bitcoinDataSource indexers.BitcoinDatasource
	network           common.Network

	newRuneEntries      map[runes.RuneId]*runes.RuneEntry
	newRuneEntryStates  map[runes.RuneId]*runes.RuneEntry
	newOutPointBalances map[wire.OutPoint]map[runes.RuneId]uint128.Uint128
	newSpendOutPoints   []wire.OutPoint
	newBalances         map[string]map[runes.RuneId]uint128.Uint128 // pkScript(hex) -> runeId -> amount
	newRuneTxs          []*entity.RuneTransaction
}

func NewProcessor(runesDg datagateway.RunesDataGateway, indexerInfoDg datagateway.IndexerInfoDataGateway, bitcoinClient btcclient.Contract, bitcoinDataSource indexers.BitcoinDatasource, network common.Network) *Processor {
	return &Processor{
		runesDg:             runesDg,
		indexerInfoDg:       indexerInfoDg,
		bitcoinClient:       bitcoinClient,
		bitcoinDataSource:   bitcoinDataSource,
		network:             network,
		newRuneEntries:      make(map[runes.RuneId]*runes.RuneEntry),
		newRuneEntryStates:  make(map[runes.RuneId]*runes.RuneEntry),
		newOutPointBalances: make(map[wire.OutPoint]map[runes.RuneId]uint128.Uint128),
		newSpendOutPoints:   make([]wire.OutPoint, 0),
		newBalances:         make(map[string]map[runes.RuneId]uint128.Uint128),
		newRuneTxs:          make([]*entity.RuneTransaction, 0),
	}
}

var (
	ErrDBVersionMismatch        = errors.New("db version mismatch: please migrate db")
	ErrEventHashVersionMismatch = errors.New("event hash version mismatch: please reset db and reindex")
)

func (p *Processor) VerifyStates(ctx context.Context) error {
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
	indexerState, err := p.indexerInfoDg.GetLatestIndexerState(ctx)
	if err != nil && !errors.Is(err, errs.NotFound) {
		return errors.Wrap(err, "failed to get latest indexer state")
	}
	// if not found, set indexer state
	if errors.Is(err, errs.NotFound) {
		if err := p.indexerInfoDg.SetIndexerState(ctx, entity.IndexerState{
			DBVersion:        DBVersion,
			EventHashVersion: EventHashVersion,
		}); err != nil {
			return errors.Wrap(err, "failed to set indexer state")
		}
	} else {
		if indexerState.DBVersion != DBVersion {
			return errors.Wrapf(errs.ConflictSetting, "db version mismatch: current version is %d. Please upgrade to version %d", indexerState.DBVersion, DBVersion)
		}
		if indexerState.EventHashVersion != EventHashVersion {
			// TODO: automate reset db instead of returning error
			return errors.Wrapf(errs.ConflictSetting, "event version mismatch: current version is %d. Please reset rune's db first.", indexerState.EventHashVersion, EventHashVersion)
		}
	}

	_, network, err := p.indexerInfoDg.GetLatestIndexerStats(ctx)
	if err != nil && !errors.Is(err, errs.NotFound) {
		return errors.Wrap(err, "failed to get latest indexer stats")
	}
	// if found, verify indexer stats
	if err == nil {
		if network != p.network {
			return errors.Wrapf(errs.ConflictSetting, "network mismatch: latest indexed network is %d, configured network is %d. If you want to change the network, please reset the database", network, p.network)
		}
	}
	if err := p.indexerInfoDg.UpdateIndexerStats(ctx, p.network.String(), p.network); err != nil {
		return errors.Wrap(err, "failed to update indexer stats")
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
			Number:       0,
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
			Turbo:             true,
			Mints:             uint128.Zero,
			BurnedAmount:      uint128.Zero,
			CompletedAt:       time.Time{},
			CompletedAtHeight: nil,
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
		if errors.Is(err, errs.NotFound) {
			return startingBlockHeader[p.network], nil
		}
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
