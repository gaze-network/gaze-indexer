package postgres

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

var _ datagateway.BRC20DataGateway = (*Repository)(nil)

// warning: GetLatestBlock currently returns a types.BlockHeader with only Height and Hash fields populated.
// This is because it is known that all usage of this function only requires these fields. In the future, we may want to populate all fields for type safety.
func (r *Repository) GetLatestBlock(ctx context.Context) (types.BlockHeader, error) {
	block, err := r.queries.GetLatestIndexedBlock(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return types.BlockHeader{}, errors.WithStack(errs.NotFound)
		}
		return types.BlockHeader{}, errors.Wrap(err, "error during query")
	}
	hash, err := chainhash.NewHashFromStr(block.Hash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to parse block hash")
	}
	return types.BlockHeader{
		Height: int64(block.Height),
		Hash:   *hash,
	}, nil
}

// GetIndexedBlockByHeight implements datagateway.BRC20DataGateway.
func (r *Repository) GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error) {
	model, err := r.queries.GetIndexedBlockByHeight(ctx, int32(height))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	indexedBlock, err := mapIndexedBlockModelToType(model)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse indexed block model")
	}
	return &indexedBlock, nil
}

func (r *Repository) GetProcessorStats(ctx context.Context) (*entity.ProcessorStats, error) {
	model, err := r.queries.GetLatestProcessorStats(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.WithStack(err)
	}
	stats := mapProcessorStatsModelToType(model)
	return &stats, nil
}

func (r *Repository) GetInscriptionTransfersInOutPoints(ctx context.Context, outPoints []wire.OutPoint) (map[ordinals.SatPoint][]*entity.InscriptionTransfer, error) {
	txHashArr := lo.Map(outPoints, func(outPoint wire.OutPoint, _ int) string {
		return outPoint.Hash.String()
	})
	txOutIdxArr := lo.Map(outPoints, func(outPoint wire.OutPoint, _ int) int32 {
		return int32(outPoint.Index)
	})
	models, err := r.queries.GetInscriptionTransfersInOutPoints(ctx, gen.GetInscriptionTransfersInOutPointsParams{
		TxHashArr:   txHashArr,
		TxOutIdxArr: txOutIdxArr,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	results := make(map[ordinals.SatPoint][]*entity.InscriptionTransfer)
	for _, model := range models {
		inscriptionTransfer, err := mapInscriptionTransferModelToType(model)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		results[inscriptionTransfer.NewSatPoint] = append(results[inscriptionTransfer.NewSatPoint], &inscriptionTransfer)
	}
	return results, nil
}

func (r *Repository) GetInscriptionEntriesByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]*ordinals.InscriptionEntry, error) {
	idStrs := lo.Map(ids, func(id ordinals.InscriptionId, _ int) string { return id.String() })
	models, err := r.queries.GetInscriptionEntriesByIds(ctx, idStrs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry)
	for _, model := range models {
		inscriptionEntry, err := mapInscriptionEntryModelToType(model)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse inscription entry model")
		}
		result[inscriptionEntry.Id] = &inscriptionEntry
	}
	return result, nil
}

func (r *Repository) GetInscriptionNumbersByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]int64, error) {
	idStrs := lo.Map(ids, func(id ordinals.InscriptionId, _ int) string { return id.String() })
	models, err := r.queries.GetInscriptionNumbersByIds(ctx, idStrs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[ordinals.InscriptionId]int64)
	for _, model := range models {
		inscriptionId, err := ordinals.NewInscriptionIdFromString(model.Id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse inscription id")
		}
		result[inscriptionId] = model.Number
	}
	return result, nil
}

func (r *Repository) GetInscriptionParentsByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]ordinals.InscriptionId, error) {
	idStrs := lo.Map(ids, func(id ordinals.InscriptionId, _ int) string { return id.String() })
	models, err := r.queries.GetInscriptionParentsByIds(ctx, idStrs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[ordinals.InscriptionId]ordinals.InscriptionId)
	for _, model := range models {
		if len(model.Parents) == 0 {
			// no parent
			continue
		}
		if len(model.Parents) > 1 {
			// sanity check, should not happen since 0.14 ord supports only 1 parent
			continue
		}
		inscriptionId, err := ordinals.NewInscriptionIdFromString(model.Id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse inscription id")
		}
		parentId, err := ordinals.NewInscriptionIdFromString(model.Parents[0])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse parent id")
		}
		result[inscriptionId] = parentId
	}
	return result, nil
}

func (r *Repository) GetLatestEventId(ctx context.Context) (int64, error) {
	row, err := r.queries.GetLatestEventIds(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return max(row.EventDeployID.(int64), row.EventMintID.(int64), row.EventInscribeTransferID.(int64), row.EventTransferTransferID.(int64)), nil
}

func (r *Repository) GetBalancesBatchAtHeight(ctx context.Context, blockHeight uint64, queries []datagateway.GetBalancesBatchAtHeightQuery) (map[string]map[string]*entity.Balance, error) {
	pkScripts := make([]string, 0)
	ticks := make([]string, 0)
	for _, query := range queries {
		pkScripts = append(pkScripts, query.PkScriptHex)
		ticks = append(ticks, query.Tick)
	}
	models, err := r.queries.GetBalancesBatchAtHeight(ctx, gen.GetBalancesBatchAtHeightParams{
		PkscriptArr: pkScripts,
		TickArr:     ticks,
		BlockHeight: int32(blockHeight),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[string]map[string]*entity.Balance)
	for _, model := range models {
		balance, err := mapBalanceModelToType(model)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance model")
		}
		if _, ok := result[model.Pkscript]; !ok {
			result[model.Pkscript] = make(map[string]*entity.Balance)
		}
		result[model.Pkscript][model.Tick] = &balance
	}
	return result, nil
}

func (r *Repository) GetEventInscribeTransfersByInscriptionIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]*entity.EventInscribeTransfer, error) {
	idStrs := lo.Map(ids, func(id ordinals.InscriptionId, _ int) string { return id.String() })
	models, err := r.queries.GetEventInscribeTransfersByInscriptionIds(ctx, idStrs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[ordinals.InscriptionId]*entity.EventInscribeTransfer)
	for _, model := range models {
		event, err := mapEventInscribeTransferModelToType(model)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse event inscribe transfer model")
		}
		result[event.InscriptionId] = &event
	}
	return result, nil
}

func (r *Repository) GetTickEntriesByTicks(ctx context.Context, ticks []string) (map[string]*entity.TickEntry, error) {
	models, err := r.queries.GetTickEntriesByTicks(ctx, ticks)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[string]*entity.TickEntry)
	for _, model := range models {
		tickEntry, err := mapTickEntryModelToType(model)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse tick entry model")
		}
		result[tickEntry.Tick] = &tickEntry
	}
	return result, nil
}

func (r *Repository) CreateIndexedBlock(ctx context.Context, block *entity.IndexedBlock) error {
	params := mapIndexedBlockTypeToParams(*block)
	if err := r.queries.CreateIndexedBlock(ctx, params); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) CreateProcessorStats(ctx context.Context, stats *entity.ProcessorStats) error {
	params := mapProcessorStatsTypeToParams(*stats)
	if err := r.queries.CreateProcessorStats(ctx, params); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Repository) CreateTickEntries(ctx context.Context, blockHeight uint64, entries []*entity.TickEntry) error {
	entryParams := make([]gen.CreateTickEntriesParams, 0)
	for _, entry := range entries {
		params, _, err := mapTickEntryTypeToParams(*entry, blockHeight)
		if err != nil {
			return errors.Wrap(err, "cannot map tick entry to create params")
		}
		entryParams = append(entryParams, params)
	}
	results := r.queries.CreateTickEntries(ctx, entryParams)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateTickEntryStates(ctx context.Context, blockHeight uint64, entryStates []*entity.TickEntry) error {
	entryParams := make([]gen.CreateTickEntryStatesParams, 0)
	for _, entry := range entryStates {
		_, params, err := mapTickEntryTypeToParams(*entry, blockHeight)
		if err != nil {
			return errors.Wrap(err, "cannot map tick entry to create params")
		}
		entryParams = append(entryParams, params)
	}
	results := r.queries.CreateTickEntryStates(ctx, entryParams)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateInscriptionEntries(ctx context.Context, blockHeight uint64, entries []*ordinals.InscriptionEntry) error {
	inscriptionEntryParams := make([]gen.CreateInscriptionEntriesParams, 0)
	for _, entry := range entries {
		params, _, err := mapInscriptionEntryTypeToParams(*entry, blockHeight)
		if err != nil {
			return errors.Wrap(err, "cannot map inscription entry to create params")
		}
		inscriptionEntryParams = append(inscriptionEntryParams, params)
	}
	results := r.queries.CreateInscriptionEntries(ctx, inscriptionEntryParams)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateInscriptionEntryStates(ctx context.Context, blockHeight uint64, entryStates []*ordinals.InscriptionEntry) error {
	inscriptionEntryStatesParams := make([]gen.CreateInscriptionEntryStatesParams, 0)
	for _, entry := range entryStates {
		_, params, err := mapInscriptionEntryTypeToParams(*entry, blockHeight)
		if err != nil {
			return errors.Wrap(err, "cannot map inscription entry to create params")
		}
		inscriptionEntryStatesParams = append(inscriptionEntryStatesParams, params)
	}
	results := r.queries.CreateInscriptionEntryStates(ctx, inscriptionEntryStatesParams)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateInscriptionTransfers(ctx context.Context, transfers []*entity.InscriptionTransfer) error {
	params := lo.Map(transfers, func(transfer *entity.InscriptionTransfer, _ int) gen.CreateInscriptionTransfersParams {
		return mapInscriptionTransferTypeToParams(*transfer)
	})
	results := r.queries.CreateInscriptionTransfers(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateEventDeploys(ctx context.Context, events []*entity.EventDeploy) error {
	params := make([]gen.CreateEventDeploysParams, 0)
	for _, event := range events {
		param, err := mapEventDeployTypeToParams(*event)
		if err != nil {
			return errors.Wrap(err, "cannot map event deploy to create params")
		}
		params = append(params, param)
	}
	results := r.queries.CreateEventDeploys(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateEventMints(ctx context.Context, events []*entity.EventMint) error {
	params := make([]gen.CreateEventMintsParams, 0)
	for _, event := range events {
		param, err := mapEventMintTypeToParams(*event)
		if err != nil {
			return errors.Wrap(err, "cannot map event mint to create params")
		}
		params = append(params, param)
	}
	results := r.queries.CreateEventMints(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateEventInscribeTransfers(ctx context.Context, events []*entity.EventInscribeTransfer) error {
	params := make([]gen.CreateEventInscribeTransfersParams, 0)
	for _, event := range events {
		param, err := mapEventInscribeTransferTypeToParams(*event)
		if err != nil {
			return errors.Wrap(err, "cannot map event transfer to create params")
		}
		params = append(params, param)
	}
	results := r.queries.CreateEventInscribeTransfers(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateEventTransferTransfers(ctx context.Context, events []*entity.EventTransferTransfer) error {
	params := make([]gen.CreateEventTransferTransfersParams, 0)
	for _, event := range events {
		param, err := mapEventTransferTransferTypeToParams(*event)
		if err != nil {
			return errors.Wrap(err, "cannot map event transfer to create params")
		}
		params = append(params, param)
	}
	results := r.queries.CreateEventTransferTransfers(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) CreateBalances(ctx context.Context, balances []*entity.Balance) error {
	params := lo.Map(balances, func(balance *entity.Balance, _ int) gen.CreateBalancesParams {
		return mapBalanceTypeToParams(*balance)
	})
	results := r.queries.CreateBalances(ctx, params)
	var execErrors []error
	results.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) DeleteIndexedBlocksSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteIndexedBlocksSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteProcessorStatsSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteProcessorStatsSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteTickEntriesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteTickEntriesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteTickEntryStatesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteTickEntryStatesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteEventDeploysSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteEventDeploysSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteEventMintsSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteEventMintsSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteEventInscribeTransfersSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteEventInscribeTransfersSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteEventTransferTransfersSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteEventTransferTransfersSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteBalancesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteInscriptionEntriesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteInscriptionEntriesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteInscriptionEntryStatesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteInscriptionEntryStatesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteInscriptionTransfersSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteInscriptionTransfersSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}
