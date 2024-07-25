package usecase

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
)

func (u *Usecase) GetRunesUTXOsByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.RunesUTXO, error) {
	balances, err := u.runesDg.GetRunesUTXOsByPkScript(ctx, pkScript, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}

func (u *Usecase) GetRunesUTXOsByRuneIdAndPkScript(ctx context.Context, runeId runes.RuneId, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.RunesUTXO, error) {
	balances, err := u.runesDg.GetRunesUTXOsByRuneIdAndPkScript(ctx, runeId, pkScript, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}

func (u *Usecase) GetUTXOsOutputByLocation(ctx context.Context, txHash chainhash.Hash, outputIdx uint32) (*entity.RunesUTXO, error) {
	tx, err := u.bitcoinClient.GetRawTransactionByTxHash(ctx, txHash)
	if err != nil {
		return nil, errors.WithStack(ErrUTXONotFound)
	}

	// If the output index is out of range, return an error
	if len(tx.TxOut) <= int(outputIdx) {
		return nil, errors.WithStack(ErrUTXONotFound)
	}

	rune := &entity.RunesUTXO{
		PkScript: tx.TxOut[0].PkScript,
		OutPoint: wire.OutPoint{
			Hash:  txHash,
			Index: outputIdx,
		},
	}

	transaction, err := u.runesDg.GetRuneTransaction(ctx, txHash)
	// If Bitcoin transaction is not found in the database, return the PkScript and OutPoint
	if errors.Is(err, errs.NotFound) {
		return rune, nil
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	runeBalance := make([]entity.RunesUTXOBalance, 0, len(transaction.Outputs))
	for _, output := range transaction.Outputs {
		if output.Index == outputIdx {
			runeBalance = append(runeBalance, entity.RunesUTXOBalance{
				RuneId: output.RuneId,
				Amount: output.Amount,
			})
		}
	}

	rune.RuneBalances = runeBalance
	return rune, nil
}
