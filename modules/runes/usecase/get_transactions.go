package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
)

func (u *Usecase) GetRuneTransactions(ctx context.Context, pkScript []byte, runeId runes.RuneId, fromBlock, toBlock uint64) ([]*entity.RuneTransaction, error) {
	txs, err := u.runesDg.GetRuneTransactions(ctx, pkScript, runeId, fromBlock, toBlock)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetTransactionsByHeight")
	}
	return txs, nil
}
