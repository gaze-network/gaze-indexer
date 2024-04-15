package postgres

import (
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/indexer-network/modules/runes/repository/postgres/gen"
	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
)

var _ datagateway.RunesDataGateway = (*Repository)(nil)

func (r *Repository) GetLatestIndexerState(ctx context.Context) (entity.IndexerState, error) {
	indexerStateModel, err := r.getQueries().GetLatestIndexerState(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.IndexerState{}, errors.WithStack(errs.NotFound)
		}
		return entity.IndexerState{}, errors.Wrap(err, "error during query")
	}
	indexerState := mapIndexerStateModelToType(indexerStateModel)
	return indexerState, nil
}

// warning: GetLatestBlock currently returns a types.BlockHeader with only Height, Hash, and PrevBlock fields populated.
// This is because it is known that all usage of this function only requires these fields. In the future, we may want to populate all fields for type safety.
func (r *Repository) GetLatestBlock(ctx context.Context) (types.BlockHeader, error) {
	block, err := r.getQueries().GetLatestIndexedBlock(ctx)
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
	indexedBlockModel, err := r.getQueries().GetIndexedBlockByHeight(ctx, int32(height))
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

func (r *Repository) GetRuneTransactionsByHeight(ctx context.Context, height uint64) ([]*entity.RuneTransaction, error) {
	rows, err := r.getQueries().GetRuneTransactionsByHeight(ctx, int32(height))
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

func (r *Repository) GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error) {
	balances, err := r.getQueries().GetOutPointBalances(ctx, gen.GetOutPointBalancesParams{
		TxHash: outPoint.Hash.String(),
		TxIdx:  int32(outPoint.Index),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make(map[runes.RuneId]uint128.Uint128, len(balances))
	for _, balance := range balances {
		runeId, err := runes.NewRuneIdFromString(balance.RuneID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse RuneId")
		}
		amount, err := uint128FromNumeric(balance.Amount)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance")
		}
		result[runeId] = lo.FromPtr(amount)
	}
	return result, nil
}

func (r *Repository) GetRuneIdFromRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error) {
	runeIdStr, err := r.getQueries().GetRuneIdFromRune(ctx, rune.String())
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
	rows, err := r.getQueries().GetRuneEntriesByRuneIds(ctx, lo.Map(runeIds, func(runeId runes.RuneId, _ int) string {
		return runeId.String()
	}))
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntries := make(map[runes.RuneId]*runes.RuneEntry, len(rows))
	var errs []error
	for i, runeEntryModel := range rows {
		runeEntry, err := mapRuneEntryModelToType(runeEntryModel)
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

func (r *Repository) CountRuneEntries(ctx context.Context) (uint64, error) {
	count, err := r.getQueries().CountRuneEntries(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "error during query")
	}
	return uint64(count), nil
}

func (r *Repository) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) (map[runes.RuneId]*entity.Balance, error) {
	balances, err := r.getQueries().GetBalancesByPkScript(ctx, gen.GetBalancesByPkScriptParams{
		Pkscript:    hex.EncodeToString(pkScript),
		BlockHeight: int32(blockHeight),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make(map[runes.RuneId]*entity.Balance, len(balances))
	for _, balanceModel := range balances {
		balance, err := mapBalanceModelToType(balanceModel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance model")
		}
		result[balance.RuneId] = balance
	}
	return result, nil
}

func (r *Repository) GetBalancesByRuneId(ctx context.Context, runeId runes.RuneId, blockHeight uint64) ([]*entity.Balance, error) {
	balances, err := r.getQueries().GetBalancesByRuneId(ctx, gen.GetBalancesByRuneIdParams{
		RuneID:      runeId.String(),
		BlockHeight: int32(blockHeight),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make([]*entity.Balance, 0, len(balances))
	for _, balanceModel := range balances {
		balance, err := mapBalanceModelToType(balanceModel)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance model")
		}
		result = append(result, balance)
	}
	return result, nil
}

func (r *Repository) GetBalanceByPkScriptAndRuneId(ctx context.Context, pkScript []byte, runeId runes.RuneId, blockHeight uint64) (*entity.Balance, error) {
	balance, err := r.getQueries().GetBalanceByPkScriptAndRuneId(ctx, gen.GetBalanceByPkScriptAndRuneIdParams{
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

func (r *Repository) SetIndexerState(ctx context.Context, state entity.IndexerState) error {
	params := mapIndexerStateTypeToParams(state)
	if err := r.getQueries().SetIndexerState(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) CreateRuneTransaction(ctx context.Context, tx *entity.RuneTransaction) error {
	if tx == nil {
		return nil
	}
	txParams, runestoneParams, err := mapRuneTransactionTypeToParams(*tx)
	if err != nil {
		return errors.Wrap(err, "failed to map rune transaction to params")
	}
	if err = r.getQueries().CreateRuneTransaction(ctx, txParams); err != nil {
		return errors.Wrap(err, "error during exec CreateRuneTransaction")
	}
	if runestoneParams != nil {
		if err = r.getQueries().CreateRunestone(ctx, *runestoneParams); err != nil {
			return errors.Wrap(err, "error during exec CreateRunestone")
		}
	}
	return nil
}

func (r *Repository) CreateRuneEntry(ctx context.Context, entry *runes.RuneEntry, blockHeight uint64) error {
	if entry == nil {
		return nil
	}
	createParams, _, err := mapRuneEntryTypeToParams(*entry, blockHeight)
	if err != nil {
		return errors.Wrap(err, "failed to map rune entry to params")
	}
	if err = r.getQueries().CreateRuneEntry(ctx, createParams); err != nil {
		return errors.Wrap(err, "error during exec CreateRuneEntry")
	}
	return nil
}

func (r *Repository) CreateRuneEntryState(ctx context.Context, entry *runes.RuneEntry, blockHeight uint64) error {
	if entry == nil {
		return nil
	}
	_, createStateParams, err := mapRuneEntryTypeToParams(*entry, blockHeight)
	if err != nil {
		return errors.Wrap(err, "failed to map rune entry to params")
	}
	if err = r.getQueries().CreateRuneEntryState(ctx, createStateParams); err != nil {
		return errors.Wrap(err, "error during exec CreateRuneEntryState")
	}
	return nil
}

func (r *Repository) CreateOutPointBalances(ctx context.Context, outPoint wire.OutPoint, balances map[runes.RuneId]uint128.Uint128, blockHeight uint64) error {
	params := make([]gen.CreateOutPointBalancesParams, 0, len(balances))
	for runeId, balance := range balances {
		balance := balance
		amount, err := numericFromUint128(&balance)
		if err != nil {
			return errors.Wrap(err, "failed to convert balance to numeric")
		}
		params = append(params, gen.CreateOutPointBalancesParams{
			RuneID:      runeId.String(),
			TxHash:      outPoint.Hash.String(),
			TxIdx:       int32(outPoint.Index),
			Amount:      amount,
			BlockHeight: int32(blockHeight),
			SpentHeight: pgtype.Int4{Valid: false},
		})
	}
	result := r.getQueries().CreateOutPointBalances(ctx, params)
	var execErrors []error
	result.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
	}
	return nil
}

func (r *Repository) SpendOutPointBalances(ctx context.Context, outPoint wire.OutPoint, blockHeight uint64) error {
	if err := r.getQueries().SpendOutPointBalances(ctx, gen.SpendOutPointBalancesParams{
		TxHash:      outPoint.Hash.String(),
		TxIdx:       int32(outPoint.Index),
		SpentHeight: pgtype.Int4{Int32: int32(blockHeight), Valid: true},
	}); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) CreateRuneBalances(ctx context.Context, params []datagateway.CreateRuneBalancesParams) error {
	insertParams := make([]gen.CreateRuneBalanceAtBlockParams, 0, len(params))
	for _, param := range params {
		param := param
		amount, err := numericFromUint128(&param.Balance)
		if err != nil {
			return errors.Wrap(err, "failed to convert balance to numeric")
		}
		insertParams = append(insertParams, gen.CreateRuneBalanceAtBlockParams{
			Pkscript:    hex.EncodeToString(param.PkScript),
			BlockHeight: int32(param.BlockHeight),
			RuneID:      param.RuneId.String(),
			Amount:      amount,
		})
	}
	result := r.getQueries().CreateRuneBalanceAtBlock(ctx, insertParams)
	var execErrors []error
	result.Exec(func(i int, err error) {
		if err != nil {
			execErrors = append(execErrors, err)
		}
	})
	if len(execErrors) > 0 {
		return errors.Wrap(errors.Join(execErrors...), "error during exec")
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
	if err = r.getQueries().CreateIndexedBlock(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteIndexedBlockSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteIndexedBlockSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneEntriesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteRuneEntriesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneEntryStatesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteRuneEntryStatesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneTransactionsSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteRuneTransactionsSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRunestonesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteRunestonesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteOutPointBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteOutPointBalancesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) UnspendOutPointBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().UnspendOutPointBalancesSinceHeight(ctx, pgtype.Int4{Int32: int32(height), Valid: true}); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) DeleteRuneBalancesSinceHeight(ctx context.Context, height uint64) error {
	if err := r.getQueries().DeleteRuneBalancesSinceHeight(ctx, int32(height)); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}
