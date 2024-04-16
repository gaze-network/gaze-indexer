package postgres

import (
	"cmp"
	"context"
	"encoding/hex"
	"slices"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres/gen"
	"github.com/samber/lo"
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

func (r *Repository) RevertBlocks(ctx context.Context, from int64) error {
	if err := r.queries.RevertData(ctx, int32(from)); err != nil {
		return errors.Wrap(err, "failed to revert data")
	}
	return nil
}

func (r *Repository) GetBlocksByHeightRange(ctx context.Context, from int64, to int64) ([]*types.Block, error) {
	blocks, err := r.queries.GetBlocksByHeightRange(ctx, gen.GetBlocksByHeightRangeParams{
		FromHeight: int32(from),
		ToHeight:   int32(to),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get blocks by height range")
	}

	txs, err := r.queries.GetTransactionsByHeightRange(ctx, gen.GetTransactionsByHeightRangeParams{
		FromHeight: int32(from),
		ToHeight:   int32(to),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transactions by height range")
	}

	txHashes := lo.Map(txs, func(tx gen.BitcoinTransaction, _ int) string { return tx.TxHash })

	txOuts, err := r.queries.GetTransactionTxOutsByTxHashes(ctx, txHashes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction txouts by tx hashes")
	}

	txIns, err := r.queries.GetTransactionTxInsByTxHashes(ctx, txHashes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction txins by tx hashes")
	}

	// Grouping result by block height and tx hash
	groupedTxs := lo.GroupBy(txs, func(tx gen.BitcoinTransaction) int32 { return tx.BlockHeight })
	groupedTxOuts := lo.GroupBy(txOuts, func(txOut gen.BitcoinTransactionTxout) string { return txOut.TxHash })
	groupedTxIns := lo.GroupBy(txIns, func(txIn gen.BitcoinTransactionTxin) string { return txIn.TxHash })

	var errs []error
	result := lo.Map(blocks, func(blockModel gen.BitcoinBlock, _ int) *types.Block {
		header, err := mapBlockHeaderModelToType(blockModel)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "failed to map block header model to type"))
			return nil
		}

		txsModel := groupedTxs[blockModel.BlockHeight]
		return &types.Block{
			Header: header,
			Transactions: lo.Map(txsModel, func(txModel gen.BitcoinTransaction, _ int) *types.Transaction {
				blockHash, err := chainhash.NewHashFromStr(txModel.BlockHash)
				if err != nil {
					errs = append(errs, errors.Wrap(err, "failed to parse block hash"))
					return nil
				}

				txHash, err := chainhash.NewHashFromStr(txModel.TxHash)
				if err != nil {
					errs = append(errs, errors.Wrap(err, "failed to parse tx hash"))
					return nil
				}

				txOutsModel := groupedTxOuts[txModel.TxHash]
				txInsModel := groupedTxIns[txModel.TxHash]

				// Sort txins and txouts by index (Asc)
				slices.SortFunc(txOutsModel, func(i, j gen.BitcoinTransactionTxout) int {
					return cmp.Compare(i.TxIdx, j.TxIdx)
				})
				slices.SortFunc(txInsModel, func(i, j gen.BitcoinTransactionTxin) int {
					return cmp.Compare(i.TxIdx, j.TxIdx)
				})

				return &types.Transaction{
					BlockHeight: int64(txModel.BlockHeight),
					BlockHash:   *blockHash,
					Index:       uint32(txModel.Idx),
					TxHash:      *txHash,
					Version:     txModel.Version,
					LockTime:    uint32(txModel.Locktime),
					TxIn: lo.Map(txInsModel, func(src gen.BitcoinTransactionTxin, _ int) *types.TxIn {
						scriptsig, err := hex.DecodeString(src.Scriptsig)
						if err != nil {
							errs = append(errs, errors.Wrap(err, "failed to decode scriptsig"))
							return nil
						}

						prevoutTxHash, err := chainhash.NewHashFromStr(src.PrevoutTxHash)
						if err != nil {
							errs = append(errs, errors.Wrap(err, "failed to parse prevout tx hash"))
							return nil
						}

						return &types.TxIn{
							SignatureScript:   scriptsig,
							Witness:           [][]byte{}, // TODO: implement witness
							Sequence:          uint32(src.Sequence),
							PreviousOutIndex:  uint32(src.PrevoutTxIdx),
							PreviousOutTxHash: *prevoutTxHash,
						}
					}),
					TxOut: lo.Map(txOutsModel, func(src gen.BitcoinTransactionTxout, _ int) *types.TxOut {
						pkscript, err := hex.DecodeString(src.Pkscript)
						if err != nil {
							errs = append(errs, errors.Wrap(err, "failed to decode pkscript"))
							return nil
						}
						return &types.TxOut{
							PkScript: pkscript,
							Value:    src.Value,
						}
					}),
				}
			}),
		}
	})
	if len(errs) > 0 {
		return nil, errors.Wrap(errors.Join(errs...), "failed while mapping result")
	}
	return result, nil
}
