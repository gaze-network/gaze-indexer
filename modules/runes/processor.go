package runes

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/constants"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/reportingclient"
	"github.com/gaze-network/uint128"
)

// Make sure to implement the Bitcoin Processor interface
var _ indexer.Processor[*types.Block] = (*Processor)(nil)

type Processor struct {
	runesDg         datagateway.RunesDataGateway
	indexerInfoDg   datagateway.IndexerInfoDataGateway
	bitcoinClient   btcclient.Contract
	network         common.Network
	reportingClient *reportingclient.ReportingClient
	cleanupFuncs    []func(context.Context) error

	newRuneEntries      map[runes.RuneId]*runes.RuneEntry
	newRuneEntryStates  map[runes.RuneId]*runes.RuneEntry
	newOutPointBalances map[wire.OutPoint][]*entity.OutPointBalance
	newSpendOutPoints   []wire.OutPoint
	newBalances         map[string]map[runes.RuneId]uint128.Uint128 // pkScript(hex) -> runeId -> amount
	newRuneTxs          []*entity.RuneTransaction
}

func NewProcessor(runesDg datagateway.RunesDataGateway, indexerInfoDg datagateway.IndexerInfoDataGateway, bitcoinClient btcclient.Contract, network common.Network, reportingClient *reportingclient.ReportingClient, cleanupFuncs []func(context.Context) error) *Processor {
	return &Processor{
		runesDg:             runesDg,
		indexerInfoDg:       indexerInfoDg,
		bitcoinClient:       bitcoinClient,
		network:             network,
		reportingClient:     reportingClient,
		cleanupFuncs:        cleanupFuncs,
		newRuneEntries:      make(map[runes.RuneId]*runes.RuneEntry),
		newRuneEntryStates:  make(map[runes.RuneId]*runes.RuneEntry),
		newOutPointBalances: make(map[wire.OutPoint][]*entity.OutPointBalance),
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
	if constants.NetworkHasGenesisRune(p.network) {
		if err := p.ensureGenesisRune(ctx, p.network); err != nil {
			return errors.Wrap(err, "error during ensureGenesisRune")
		}
	}
	if p.reportingClient != nil {
		if err := p.reportingClient.SubmitNodeReport(ctx, "runes", p.network); err != nil {
			return errors.Wrap(err, "failed to submit node report")
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
			DBVersion:        constants.DBVersion,
			EventHashVersion: constants.EventHashVersion,
		}); err != nil {
			return errors.Wrap(err, "failed to set indexer state")
		}
	} else {
		if indexerState.DBVersion != constants.DBVersion {
			return errors.Wrapf(errs.ConflictSetting, "db version mismatch: current version is %d. Please upgrade to version %d", indexerState.DBVersion, constants.DBVersion)
		}
		if indexerState.EventHashVersion != constants.EventHashVersion {
			return errors.Wrapf(errs.ConflictSetting, "event version mismatch: current version is %d. Please reset rune's db first.", indexerState.EventHashVersion, constants.EventHashVersion)
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

func (p *Processor) ensureGenesisRune(ctx context.Context, network common.Network) error {
	genesisRuneConfig, ok := constants.GenesisRuneConfigMap[network]
	if !ok {
		logger.Panic("genesis rune config not found", slogx.Stringer("network", network))
	}
	_, err := p.runesDg.GetRuneEntryByRuneId(ctx, genesisRuneConfig.RuneId)
	if err != nil && !errors.Is(err, errs.NotFound) {
		return errors.Wrap(err, "failed to get genesis rune entry")
	}
	if errors.Is(err, errs.NotFound) {
		runeEntry := &runes.RuneEntry{
			RuneId:            genesisRuneConfig.RuneId,
			Number:            genesisRuneConfig.Number,
			Divisibility:      genesisRuneConfig.Divisibility,
			Premine:           genesisRuneConfig.Premine,
			SpacedRune:        genesisRuneConfig.SpacedRune,
			Symbol:            genesisRuneConfig.Symbol,
			Terms:             genesisRuneConfig.Terms,
			Turbo:             genesisRuneConfig.Turbo,
			Mints:             uint128.Zero,
			BurnedAmount:      uint128.Zero,
			CompletedAt:       time.Time{},
			CompletedAtHeight: nil,
			EtchingBlock:      genesisRuneConfig.RuneId.BlockHeight,
			EtchingTxHash:     genesisRuneConfig.EtchingTxHash,
			EtchedAt:          genesisRuneConfig.EtchedAt,
		}
		if err := p.runesDg.CreateRuneEntries(ctx, []*runes.RuneEntry{runeEntry}); err != nil {
			return errors.Wrap(err, "failed to create genesis rune entry")
		}
		if err := p.runesDg.CreateRuneEntryStates(ctx, []*runes.RuneEntry{runeEntry}, genesisRuneConfig.RuneId.BlockHeight); err != nil {
			return errors.Wrap(err, "failed to create genesis rune entry")
		}
	}
	return nil
}

func (p *Processor) Name() string {
	return "runes"
}

func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	blockHeader, err := p.runesDg.GetLatestBlock(ctx)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return constants.StartingBlockHeader[p.network], nil
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
	runesDgTx, err := p.runesDg.BeginRunesTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err := runesDgTx.Rollback(ctx); err != nil {
			logger.WarnContext(ctx, "failed to rollback transaction",
				slogx.Error(err),
				slogx.String("event", "rollback_runes_revert"),
			)
		}
	}()

	if err := runesDgTx.DeleteIndexedBlockSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete indexed blocks")
	}
	if err := runesDgTx.DeleteRuneEntriesSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete rune entries")
	}
	if err := runesDgTx.DeleteRuneEntryStatesSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete rune entry states")
	}
	if err := runesDgTx.DeleteRuneTransactionsSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete rune transactions")
	}
	if err := runesDgTx.DeleteRunestonesSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete runestones")
	}
	if err := runesDgTx.DeleteOutPointBalancesSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete outpoint balances")
	}
	if err := runesDgTx.UnspendOutPointBalancesSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to unspend outpoint balances")
	}
	if err := runesDgTx.DeleteRuneBalancesSinceHeight(ctx, uint64(from)); err != nil {
		return errors.Wrap(err, "failed to delete rune balances")
	}

	if err := runesDgTx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

func (p *Processor) Shutdown(ctx context.Context) error {
	var errs []error
	for _, cleanup := range p.cleanupFuncs {
		if err := cleanup(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.WithStack(errors.Join(errs...))
}
