package postgres

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
)

var _ datagateway.RunesDataGateway = (*Repository)(nil)

// warning: GetLatestBlock currently returns a types.BlockHeader with only Height, Hash, and PrevBlock fields populated.
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
	prevHash, err := chainhash.NewHashFromStr(block.PrevHash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to parse prev block hash")
	}
	return types.BlockHeader{
		Height:    int64(block.Height),
		Hash:      *hash,
		PrevBlock: *prevHash,
	}, nil
}

func (r *Repository) GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error) {
	indexedBlockModel, err := r.queries.GetIndexedBlockByHeight(ctx, int32(height))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	indexedBlock, err := mapIndexedBlockModelToType(indexedBlockModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse indexed block model")
	}
	return indexedBlock, nil
}

const maxRuneTransactionsLimit = 10000 // temporary limit to prevent large queries from overwhelming the database

func (r *Repository) GetRuneTransactions(ctx context.Context, pkScript []byte, runeId runes.RuneId, fromBlock, toBlock uint64, limit int32, offset int32) ([]*entity.RuneTransaction, error) {
	if limit == -1 {
		limit = maxRuneTransactionsLimit
	}
	if limit < 0 {
		return nil, errors.Wrap(errs.InvalidArgument, "limit must be -1 or non-negative")
	}
	if limit > maxRuneTransactionsLimit {
		return nil, errors.Wrapf(errs.InvalidArgument, "limit cannot exceed %d", maxRuneTransactionsLimit)
	}
	pkScriptParam := []byte(fmt.Sprintf(`[{"pkScript":"%s"}]`, hex.EncodeToString(pkScript)))
	runeIdParam := []byte(fmt.Sprintf(`[{"runeId":"%s"}]`, runeId.String()))
	rows, err := r.queries.GetRuneTransactions(ctx, gen.GetRuneTransactionsParams{
		FilterPkScript: pkScript != nil,
		PkScriptParam:  pkScriptParam,

		FilterRuneID:      runeId != runes.RuneId{},
		RuneIDParam:       runeIdParam,
		RuneID:            []byte(runeId.String()),
		RuneIDBlockHeight: int32(runeId.BlockHeight),
		RuneIDTxIndex:     int32(runeId.TxIndex),

		FromBlock: int32(fromBlock),
		ToBlock:   int32(toBlock),

		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	runeTxs := make([]*entity.RuneTransaction, 0, len(rows))
	for _, row := range rows {
		runeTxModel, runestoneModel, err := extractModelRuneTxAndRunestone(row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to extract rune transaction and runestone from row")
		}

		runeTx, err := mapRuneTransactionModelToType(runeTxModel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse rune transaction model")
		}
		if runestoneModel != nil {
			runestone, err := mapRunestoneModelToType(*runestoneModel)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse runestone model")
			}
			runeTx.Runestone = &runestone
		}
		runeTxs = append(runeTxs, &runeTx)
	}
	return runeTxs, nil
}

func (r *Repository) GetRuneTransaction(ctx context.Context, txHash chainhash.Hash) (*entity.RuneTransaction, error) {
	row, err := r.queries.GetRuneTransaction(ctx, txHash.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	runeTxModel, runestoneModel, err := extractModelRuneTxAndRunestone(gen.GetRuneTransactionsRow(row))
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract rune transaction and runestone from row")
	}

	runeTx, err := mapRuneTransactionModelToType(runeTxModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rune transaction model")
	}

	if runestoneModel != nil {
		runestone, err := mapRunestoneModelToType(*runestoneModel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse runestone model")
		}
		runeTx.Runestone = &runestone
	}

	return &runeTx, nil
}

func (r *Repository) GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]*entity.OutPointBalance, error) {
	balances, err := r.queries.GetOutPointBalancesAtOutPoint(ctx, gen.GetOutPointBalancesAtOutPointParams{
		TxHash: outPoint.Hash.String(),
		TxIdx:  int32(outPoint.Index),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make(map[runes.RuneId]*entity.OutPointBalance, len(balances))
	for _, balanceModel := range balances {
		balance, err := mapOutPointBalanceModelToType(balanceModel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance model")
		}
		result[balance.RuneId] = &balance
	}
	return result, nil
}

func (r *Repository) GetRunesUTXOsByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.RunesUTXO, error) {
	if limit == -1 {
		limit = math.MaxInt32
	}
	if limit < 0 {
		return nil, errors.Wrap(errs.InvalidArgument, "limit must be -1 or non-negative")
	}
	rows, err := r.queries.GetRunesUTXOsByPkScript(ctx, gen.GetRunesUTXOsByPkScriptParams{
		Pkscript:    hex.EncodeToString(pkScript),
		BlockHeight: int32(blockHeight),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make([]*entity.RunesUTXO, 0, len(rows))
	for _, row := range rows {
		utxo, err := mapRunesUTXOModelToType(row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse row model")
		}
		result = append(result, &utxo)
	}
	return result, nil
}

func (r *Repository) GetRunesUTXOsByRuneIdAndPkScript(ctx context.Context, runeId runes.RuneId, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.RunesUTXO, error) {
	if limit == -1 {
		limit = math.MaxInt32
	}
	if limit < 0 {
		return nil, errors.Wrap(errs.InvalidArgument, "limit must be -1 or non-negative")
	}
	rows, err := r.queries.GetRunesUTXOsByRuneIdAndPkScript(ctx, gen.GetRunesUTXOsByRuneIdAndPkScriptParams{
		Pkscript:    hex.EncodeToString(pkScript),
		BlockHeight: int32(blockHeight),
		RuneIds:     []string{runeId.String()},
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make([]*entity.RunesUTXO, 0, len(rows))
	for _, row := range rows {
		utxo, err := mapRunesUTXOModelToType(gen.GetRunesUTXOsByPkScriptRow(row))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse row")
		}
		result = append(result, &utxo)
	}
	return result, nil
}

func (r *Repository) GetRuneIdFromRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error) {
	runeIdStr, err := r.queries.GetRuneIdFromRune(ctx, rune.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return runes.RuneId{}, errors.WithStack(errs.NotFound)
		}
		return runes.RuneId{}, errors.Wrap(err, "error during query")
	}
	runeId, err := runes.NewRuneIdFromString(runeIdStr)
	if err != nil {
		return runes.RuneId{}, errors.Wrap(err, "failed to parse RuneId")
	}
	return runeId, nil
}

func (r *Repository) GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error) {
	runeEntries, err := r.GetRuneEntryByRuneIdBatch(ctx, []runes.RuneId{runeId})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entries by rune id")
	}
	runeEntry, ok := runeEntries[runeId]
	if !ok {
		return nil, errors.WithStack(errs.NotFound)
	}
	return runeEntry, nil
}

func (r *Repository) GetRuneEntryByRuneIdBatch(ctx context.Context, runeIds []runes.RuneId) (map[runes.RuneId]*runes.RuneEntry, error) {
	rows, err := r.queries.GetRuneEntriesByRuneIds(ctx, lo.Map(runeIds, func(runeId runes.RuneId, _ int) string {
		return runeId.String()
	}))
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntries := make(map[runes.RuneId]*runes.RuneEntry, len(rows))
	var errs []error
	for i, runeEntryModel := range rows {
		runeEntry, err := mapRuneEntryModelToType(gen.GetRuneEntriesRow(runeEntryModel))
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to parse rune entry model index %d", i))
			continue
		}
		runeEntries[runeEntry.RuneId] = &runeEntry
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return runeEntries, nil
}

func (r *Repository) GetRuneEntryByRuneIdAndHeight(ctx context.Context, runeId runes.RuneId, blockHeight uint64) (*runes.RuneEntry, error) {
	runeEntries, err := r.GetRuneEntryByRuneIdBatch(ctx, []runes.RuneId{runeId})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entries by rune id")
	}
	runeEntry, ok := runeEntries[runeId]
	if !ok {
		return nil, errors.WithStack(errs.NotFound)
	}
	return runeEntry, nil
}

func (r *Repository) GetRuneEntryByRuneIdAndHeightBatch(ctx context.Context, runeIds []runes.RuneId, blockHeight uint64) (map[runes.RuneId]*runes.RuneEntry, error) {
	rows, err := r.queries.GetRuneEntriesByRuneIdsAndHeight(ctx, gen.GetRuneEntriesByRuneIdsAndHeightParams{
		RuneIds: lo.Map(runeIds, func(runeId runes.RuneId, _ int) string {
			return runeId.String()
		}),
		Height: int32(blockHeight),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntries := make(map[runes.RuneId]*runes.RuneEntry, len(rows))
	var errs []error
	for i, runeEntryModel := range rows {
		runeEntry, err := mapRuneEntryModelToType(gen.GetRuneEntriesRow(runeEntryModel))
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to parse rune entry model index %d", i))
			continue
		}
		runeEntries[runeEntry.RuneId] = &runeEntry
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return runeEntries, nil
}

func (r *Repository) GetRuneEntries(ctx context.Context, search string, blockHeight uint64, limit int32, offset int32) ([]*runes.RuneEntry, error) {
	rows, err := r.queries.GetRuneEntries(ctx, gen.GetRuneEntriesParams{
		Search: search,
		Height: int32(blockHeight),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntries := make([]*runes.RuneEntry, 0, len(rows))
	var errs []error
	for i, model := range rows {
		runeEntry, err := mapRuneEntryModelToType(model)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to parse rune entry model index %d", i))
			continue
		}
		runeEntries = append(runeEntries, &runeEntry)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return runeEntries, nil
}

func (r *Repository) GetOngoingRuneEntries(ctx context.Context, search string, blockHeight uint64, limit int32, offset int32) ([]*runes.RuneEntry, error) {
	rows, err := r.queries.GetOngoingRuneEntries(ctx, gen.GetOngoingRuneEntriesParams{
		Search: search,
		Height: int32(blockHeight),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntries := make([]*runes.RuneEntry, 0, len(rows))
	var errs []error
	for i, model := range rows {
		runeEntry, err := mapRuneEntryModelToType(gen.GetRuneEntriesRow(model))
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to parse rune entry model index %d", i))
			continue
		}
		runeEntries = append(runeEntries, &runeEntry)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return runeEntries, nil
}

func (r *Repository) CountRuneEntries(ctx context.Context) (uint64, error) {
	count, err := r.queries.CountRuneEntries(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "error during query")
	}
	return uint64(count), nil
}

func (r *Repository) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.Balance, error) {
	if limit == -1 {
		limit = math.MaxInt32
	}
	if limit < 0 {
		return nil, errors.Wrap(errs.InvalidArgument, "limit must be -1 or non-negative")
	}
	balances, err := r.queries.GetBalancesByPkScript(ctx, gen.GetBalancesByPkScriptParams{
		Pkscript:    hex.EncodeToString(pkScript),
		BlockHeight: int32(blockHeight),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make([]*entity.Balance, 0, len(balances))
	for _, balanceModel := range balances {
		balance, err := mapBalanceModelToType(gen.RunesBalance(balanceModel))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance model")
		}
		result = append(result, balance)
	}
	return result, nil
}

func (r *Repository) GetBalancesByRuneId(ctx context.Context, runeId runes.RuneId, blockHeight uint64, limit int32, offset int32) ([]*entity.Balance, error) {
	if limit == -1 {
		limit = math.MaxInt32
	}
	if limit < 0 {
		return nil, errors.Wrap(errs.InvalidArgument, "limit must be -1 or non-negative")
	}
	balances, err := r.queries.GetBalancesByRuneId(ctx, gen.GetBalancesByRuneIdParams{
		RuneID:      runeId.String(),
		BlockHeight: int32(blockHeight),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make([]*entity.Balance, 0, len(balances))
	for _, balanceModel := range balances {
		balance, err := mapBalanceModelToType(gen.RunesBalance(balanceModel))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance model")
		}
		result = append(result, balance)
	}
	return result, nil
}

func (r *Repository) GetBalanceByPkScriptAndRuneId(ctx context.Context, pkScript []byte, runeId runes.RuneId, blockHeight uint64) (*entity.Balance, error) {
	balance, err := r.queries.GetBalanceByPkScriptAndRuneId(ctx, gen.GetBalanceByPkScriptAndRuneIdParams{
		Pkscript:    hex.EncodeToString(pkScript),
		RuneID:      runeId.String(),
		BlockHeight: int32(blockHeight),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	result, err := mapBalanceModelToType(balance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse balance model")
	}
	return result, nil
}

func (r *Repository) GetTotalHoldersByRuneIds(ctx context.Context, runeIds []runes.RuneId, blockHeight uint64) (map[runes.RuneId]int64, error) {
	rows, err := r.queries.GetTotalHoldersByRuneIds(ctx, gen.GetTotalHoldersByRuneIdsParams{
		RuneIds:     lo.Map(runeIds, func(runeId runes.RuneId, _ int) string { return runeId.String() }),
		BlockHeight: int32(blockHeight),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}
	holders := make(map[runes.RuneId]int64, len(rows))
	for _, row := range rows {
		runeId, err := runes.NewRuneIdFromString(row.RuneID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse RuneId")
		}
		holders[runeId] = row.Count
	}
	return holders, nil
}

func (r *Repository) CreateRuneTransactions(ctx context.Context, txs []*entity.RuneTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	txParams, err := mapRuneTransactionTypeToParamsBatch(txs)
	if err != nil {
		return errors.Wrap(err, "failed to map rune transactions to params")
	}
	if err := r.queries.BatchCreateRuneTransactions(ctx, txParams); err != nil {
		return errors.Wrap(err, "error during exec BatchCreateRuneTransactions")
	}

	runestoneParams, err := mapRunestoneTypeToParamsBatch(txs)
	if err != nil {
		return errors.Wrap(err, "failed to map runestones to params")
	}
	if err := r.queries.BatchCreateRunestonesPatched(ctx, runestoneParams); err != nil {
		return errors.Wrap(err, "error during exec BatchCreateRunestones")
	}

	return nil
}

func (r *Repository) CreateRuneEntries(ctx context.Context, entries []*runes.RuneEntry) error {
	if len(entries) == 0 {
		return nil
	}

	params, err := mapRuneEntryTypeToParamsBatch(entries)
	if err != nil {
		return errors.Wrap(err, "failed to map rune entries to params")
	}

	if err := r.queries.BatchCreateRuneEntriesPatched(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}

	return nil
}

func (r *Repository) CreateRuneEntryStates(ctx context.Context, entries []*runes.RuneEntry, blockHeight uint64) error {
	if len(entries) == 0 {
		return nil
	}

	params, err := mapRuneEntryStatesTypeToParamsBatch(entries, blockHeight)
	if err != nil {
		return errors.Wrap(err, "failed to map rune entry states to params")
	}

	if err := r.queries.BatchCreateRuneEntryStatesPatched(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}

	return nil
}

func (r *Repository) CreateOutPointBalances(ctx context.Context, outPointBalances []*entity.OutPointBalance) error {
	if len(outPointBalances) == 0 {
		return nil
	}

	params, err := mapOutPointBalanceTypeToParamsBatch(outPointBalances)
	if err != nil {
		return errors.Wrap(err, "failed to map outpoint balances to params")
	}

	if err := r.queries.BatchCreateRunesOutpointBalancesPatched(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}

	return nil
}

func (r *Repository) SpendOutPointBalancesBatch(ctx context.Context, outPoints []wire.OutPoint, blockHeight uint64) error {
	if len(outPoints) == 0 {
		return nil
	}

	params := gen.BatchSpendOutpointBalancesParams{
		TxHashArr:   make([]string, 0, len(outPoints)),
		TxIdxArr:    make([]int32, 0, len(outPoints)),
		SpentHeight: int32(blockHeight),
	}
	for _, outPoint := range outPoints {
		params.TxHashArr = append(params.TxHashArr, outPoint.Hash.String())
		params.TxIdxArr = append(params.TxIdxArr, int32(outPoint.Index))
	}

	if err := r.queries.BatchSpendOutpointBalances(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}

	return nil
}

func (r *Repository) CreateRuneBalances(ctx context.Context, balances []*entity.Balance) error {
	if len(balances) == 0 {
		return nil
	}

	params, err := mapBalanceTypeToParamsBatch(balances)
	if err != nil {
		return errors.Wrap(err, "failed to map rune balances to params")
	}
	if err := r.queries.BatchCreateRunesBalances(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}

	return nil
}

func (r *Repository) CreateIndexedBlock(ctx context.Context, block *entity.IndexedBlock) error {
	if block == nil {
		return nil
	}
	params, err := mapIndexedBlockTypeToParams(*block)
	if err != nil {
		return errors.Wrap(err, "failed to map indexed block to params")
	}
	if err = r.queries.CreateIndexedBlock(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteIndexedBlockSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteIndexedBlockSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneEntriesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteRuneEntriesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneEntryStatesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteRuneEntryStatesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneTransactionsSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteRuneTransactionsSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRunestonesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteRunestonesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteOutPointBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteOutPointBalancesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) UnspendOutPointBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.UnspendOutPointBalancesSinceHeight(ctx, pgtype.Int4{Int32: int32(height), Valid: true}); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.queries.DeleteRuneBalancesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}
