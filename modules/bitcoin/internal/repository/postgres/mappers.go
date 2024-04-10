package postgres

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/internal/repository/postgres/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapBlockHeaderTypeToModel(src types.BlockHeader) gen.BitcoinBlock {
	return gen.BitcoinBlock{
		BlockHeight:   int32(src.Height),
		BlockHash:     src.Hash.String(),
		Version:       src.Version,
		MerkleRoot:    src.MerkleRoot.String(),
		PrevBlockHash: src.PrevBlock.String(),
		Timestamp: pgtype.Timestamptz{
			Time:  src.Timestamp,
			Valid: true,
		},
		Bits:  int32(src.Bits),
		Nonce: int32(src.Nonce),
	}
}

func mapBlockHeaderModelToType(src gen.BitcoinBlock) (types.BlockHeader, error) {
	hash, err := chainhash.NewHashFromStr(src.BlockHash)
	if err != nil {
		return types.BlockHeader{}, errors.Join(errors.Wrap(err, "failed to parse block hash"), errs.InternalError)
	}
	prevHash, err := chainhash.NewHashFromStr(src.PrevBlockHash)
	if err != nil {
		return types.BlockHeader{}, errors.Join(errors.Wrap(err, "failed to parse prev block hash"), errs.InternalError)
	}
	merkleRoot, err := chainhash.NewHashFromStr(src.MerkleRoot)
	if err != nil {
		return types.BlockHeader{}, errors.Join(errors.Wrap(err, "failed to parse merkle root"), errs.InternalError)
	}
	return types.BlockHeader{
		Hash:       *hash,
		Height:     int64(src.BlockHeight),
		Version:    src.Version,
		PrevBlock:  *prevHash,
		MerkleRoot: *merkleRoot,
		Timestamp:  src.Timestamp.Time,
		Bits:       uint32(src.Bits),
		Nonce:      uint32(src.Nonce),
	}, nil
}

func mapBlockTypeToParams(src *types.Block) (gen.InsertBlockParams, []gen.InsertTransactionParams, []gen.InsertTransactionTxOutParams, []gen.InsertTransactionTxInParams) {
	txs := make([]gen.InsertTransactionParams, 0, len(src.Transactions))
	txouts := make([]gen.InsertTransactionTxOutParams, 0)
	txins := make([]gen.InsertTransactionTxInParams, 0)
	block := gen.InsertBlockParams{
		BlockHeight:   int32(src.Header.Height),
		BlockHash:     src.Header.Hash.String(),
		Version:       src.Header.Version,
		MerkleRoot:    src.Header.MerkleRoot.String(),
		PrevBlockHash: src.Header.PrevBlock.String(),
		Timestamp: pgtype.Timestamptz{
			Time:  src.Header.Timestamp,
			Valid: true,
		},
		Bits:  int32(src.Header.Bits),
		Nonce: int32(src.Header.Nonce),
	}
	for txIdx, srcTx := range src.Transactions {
		tx := gen.InsertTransactionParams{
			TxHash:      srcTx.TxHash.String(),
			Version:     srcTx.Version,
			Locktime:    int32(srcTx.LockTime),
			BlockHeight: int32(src.Header.Height),
			BlockHash:   src.Header.Hash.String(),
			Idx:         int16(txIdx),
		}
		txs = append(txs, tx)

		for idx, txin := range srcTx.TxIn {
			txins = append(txins, gen.InsertTransactionTxInParams{
				TxHash:        tx.TxHash,
				TxIdx:         int16(idx),
				PrevoutTxHash: txin.PreviousOutTxHash.String(),
				PrevoutTxIdx:  int16(txin.PreviousOutIndex),
				Scriptsig:     hex.EncodeToString(txin.SignatureScript),
				// TODO: should figure out how to store this
				// Witness: pgtype.Text{
				// 	String: string(txin.Witness),
				// 	Valid:  true,
				// },
				Sequence: int32(txin.Sequence),
			})
		}

		for idx, txout := range srcTx.TxOut {
			txouts = append(txouts, gen.InsertTransactionTxOutParams{
				TxHash:   tx.TxHash,
				TxIdx:    int16(idx),
				Pkscript: hex.EncodeToString(txout.PkScript),
				Value:    txout.Value,
			})
		}
	}
	return block, txs, txouts, txins
}
