package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
)

func (r *Repository) GetLatestBlockHeader(ctx context.Context) (types.BlockHeader, error) {
	model, err := r.queries.GetLatestBlockHeader(ctx)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get latest block header")
	}

	data, err := mapBlockHeaderModelToType(model)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to map block header model to type")
	}

	return data, nil
}

func (r *Repository) InsertBlock(ctx context.Context, block *types.Block) error {
	blockParams, txParams, txoutParams, txinParams := mapBlockTypeToParams(block)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r) // re-throw panic after rollback
		}
	}()

	queries := r.queries.WithTx(tx)

	if err := queries.InsertBlock(ctx, blockParams); err != nil {
		_ = tx.Rollback(ctx)
		return errors.Wrapf(err, "failed to insert block, height: %d, hash: %s", blockParams.BlockHeight, blockParams.BlockHash)
	}

	for _, params := range txParams {
		if err := queries.InsertTransaction(ctx, params); err != nil {
			_ = tx.Rollback(ctx)
			return errors.Wrapf(err, "failed to insert transaction, hash: %s", params.TxHash)
		}
	}

	// Should insert txout first, then txin
	// Because txin references txout
	for _, params := range txoutParams {
		if err := queries.InsertTransactionTxOut(ctx, params); err != nil {
			_ = tx.Rollback(ctx)
			return errors.Wrapf(err, "failed to insert transaction txout, %v:%v", params.TxHash, params.TxIdx)
		}
	}

	for _, params := range txinParams {
		if err := queries.InsertTransactionTxIn(ctx, params); err != nil {
			_ = tx.Rollback(ctx)
			return errors.Wrapf(err, "failed to insert transaction txin, %v:%v", params.TxHash, params.TxIdx)
		}
	}

	return nil
}
