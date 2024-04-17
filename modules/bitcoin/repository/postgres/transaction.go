package postgres

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
)

func (r *Repository) GetTransactionByHash(ctx context.Context, txHash chainhash.Hash) (*types.Transaction, error) {
	model, err := r.queries.GetTransactionByHash(ctx, txHash.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction by hash")
	}
	txIns, err := r.queries.GetTransactionTxInsByTxHashes(ctx, []string{txHash.String()})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction txins by tx hashes")
	}
	txOuts, err := r.queries.GetTransactionTxOutsByTxHashes(ctx, []string{txHash.String()})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction txouts by tx hashes")
	}

	tx, err := mapTransactionModelToType(model, txIns, txOuts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to map transaction model to type")
	}
	return &tx, nil
}
