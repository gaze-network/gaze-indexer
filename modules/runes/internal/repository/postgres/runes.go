package postgres

import (
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/internal/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

var _ datagateway.RunesDataGateway = (*Repository)(nil)

// warning: GetLatestBlock currently returns a types.BlockHeader with only Height, Hash, and PrevBlock fields populated.
// This is because it is known that all usage of this function only requires these fields. In the future, we may want to populate all fields for type safety.
func (r *Repository) GetLatestBlock(ctx context.Context) (types.BlockHeader, error) {
	state, err := r.queries.GetRunesProcessorState(ctx)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "error during query")
	}
	hash, err := chainhash.NewHashFromStr(state.LatestBlockHash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to parse block hash")
	}
	prevHash, err := chainhash.NewHashFromStr(state.LatestPrevBlockHash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to parse prev block hash")
	}
	return types.BlockHeader{
		Height:    int64(state.LatestBlockHeight),
		Hash:      *hash,
		PrevBlock: *prevHash,
	}, nil
}

func (r *Repository) GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error) {
	indexedBlockModel, err := r.queries.GetIndexedBlockByHeight(ctx, int32(height))
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	indexedBlock, err := mapIndexedBlockModelToType(indexedBlockModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse indexed block model")
	}
	return indexedBlock, nil
}

func (r *Repository) GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error) {
	balances, err := r.queries.GetOutPointBalances(ctx, gen.GetOutPointBalancesParams{
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

func (r *Repository) GetRuneEntryByRune(ctx context.Context, rune runes.Rune) (*runes.RuneEntry, error) {
	runeEntryModels, err := r.queries.GetRuneEntriesByRunes(ctx, []string{rune.String()})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}
	if len(runeEntryModels) == 0 {
		return nil, errors.WithStack(errs.NotFound)
	}

	runeEntry, err := mapRuneEntryModelToType(runeEntryModels[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rune entry model")
	}
	return &runeEntry, nil
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
	runeEntryModels, err := r.queries.GetRuneEntriesByRuneIds(ctx, lo.Map(runeIds, func(runeId runes.RuneId, _ int) string {
		return runeId.String()
	}))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntries := make(map[runes.RuneId]*runes.RuneEntry, len(runeEntryModels))
	var errs []error
	for i, runeEntryModel := range runeEntryModels {
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

func (r *Repository) SetRuneEntry(ctx context.Context, entry *runes.RuneEntry) error {
	if entry == nil {
		return nil
	}
	params, err := mapRuneEntryTypeToParams(*entry)
	if err != nil {
		return errors.Wrap(err, "failed to map rune entry to params")
	}
	if err = r.queries.SetRuneEntry(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) CreateRuneBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint, balances map[runes.RuneId]uint128.Uint128) error {
	params := make([]gen.CreateRuneBalancesAtOutPointParams, 0, len(balances))
	for runeId, balance := range balances {
		balance := balance
		amount, err := numericFromUint128(&balance)
		if err != nil {
			return errors.Wrap(err, "failed to convert balance to numeric")
		}
		params = append(params, gen.CreateRuneBalancesAtOutPointParams{
			RuneID: runeId.String(),
			TxHash: outPoint.Hash.String(),
			TxIdx:  int32(outPoint.Index),
			Amount: amount,
		})
	}
	result := r.queries.CreateRuneBalancesAtOutPoint(ctx, params)
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

func (r *Repository) CreateRuneBalancesAtBlock(ctx context.Context, params []datagateway.CreateRuneBalancesAtBlockParams) error {
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
	result := r.queries.CreateRuneBalanceAtBlock(ctx, insertParams)
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

func (r *Repository) UpdateLatestBlock(ctx context.Context, blockHeader types.BlockHeader) error {
	params := gen.UpdateLatestBlockParams{
		LatestBlockHeight:   int32(blockHeader.Height),
		LatestBlockHash:     blockHeader.Hash.String(),
		LatestPrevBlockHash: blockHeader.PrevBlock.String(),
	}
	if err := r.queries.UpdateLatestBlock(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) (map[runes.RuneId]*entity.Balance, error) {
	balances, err := r.queries.GetBalancesByPkScript(ctx, gen.GetBalancesByPkScriptParams{
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
	balances, err := r.queries.GetBalancesByRuneId(ctx, gen.GetBalancesByRuneIdParams{
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

func (r *Repository) DeleteIndexedBlockByHash(ctx context.Context, hash chainhash.Hash) error {
	if err := r.queries.DeleteIndexedBlockByHash(ctx, hash.String()); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}
