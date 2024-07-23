package usecase

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
)

// Use limit = -1 as no limit.
func (u *Usecase) GetRuneTransactions(ctx context.Context, pkScript []byte, runeId runes.RuneId, fromBlock, toBlock uint64, limit int32, offset int32) ([]*entity.RuneTransaction, error) {
	txs, err := u.runesDg.GetRuneTransactions(ctx, pkScript, runeId, fromBlock, toBlock, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetTransactionsByHeight")
	}
	return txs, nil
}

func (u *Usecase) GetRuneTransactionByHash(ctx context.Context, hash chainhash.Hash) (*entity.RuneTransaction, error) {
	tx, err := u.runesDg.GetRuneTransactionByHash(ctx, hash)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetRuneTransactionByHash")
	}
	return tx, nil
}
