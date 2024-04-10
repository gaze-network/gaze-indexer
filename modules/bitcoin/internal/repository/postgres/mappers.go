package postgres

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/internal/repository/postgres/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapBlockHeaderTypeToModel(src *types.BlockHeader) gen.BitcoinBlock {
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

func mapBlockHeaderModelToType(src gen.BitcoinBlock) (*types.BlockHeader, error) {
	hash, err := chainhash.NewHashFromStr(src.BlockHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse block hash")
	}
	prevHash, err := chainhash.NewHashFromStr(src.PrevBlockHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse prev block hash")
	}
	merkleRoot, err := chainhash.NewHashFromStr(src.MerkleRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse merkle root")
	}
	return &types.BlockHeader{
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

func mapBlockTypeToModel(src *types.Block) (gen.BitcoinBlock, []gen.BitcoinTransaction, []gen.BitcoinTransactionTxin, []gen.BitcoinTransactionTxout) {
	block := mapBlockHeaderTypeToModel(&src.Header)
	txs := make([]gen.BitcoinTransaction, 0, len(src.Transactions))
	txins := make([]gen.BitcoinTransactionTxin, 0)
	txouts := make([]gen.BitcoinTransactionTxout, 0)
	for txIdx, srcTx := range src.Transactions {
		tx := gen.BitcoinTransaction{
			TxHash:      srcTx.TxHash.String(),
			Version:     srcTx.Version,
			Locktime:    int32(srcTx.LockTime),
			BlockHeight: int32(src.Header.Height),
			BlockHash:   src.Header.Hash.String(),
			Idx:         int16(txIdx),
		}
		txs = append(txs, tx)

		for idx, txin := range srcTx.TxIn {
			txins = append(txins, gen.BitcoinTransactionTxin{
				TxHash:          tx.TxHash,
				TxIdx:           int16(idx),
				PrevoutTxHash:   txin.PreviousOutTxHash.String(),
				PrevoutTxIdx:    int16(txin.PreviousOutIndex),
				PrevoutPkscript: "", // Note: this will be updated while inserting to the database
				Scriptsig:       hex.EncodeToString(txin.SignatureScript),
				// TODO: should figure out how to store this
				// Witness: pgtype.Text{
				// 	String: string(txin.Witness),
				// 	Valid:  true,
				// },
				Sequence: int32(txin.Sequence),
			})
		}

		for idx, txout := range srcTx.TxOut {
			txouts = append(txouts, gen.BitcoinTransactionTxout{
				TxHash:   tx.TxHash,
				TxIdx:    int16(idx),
				Pkscript: hex.EncodeToString(txout.PkScript),
				Value:    txout.Value,
				IsSpent:  false,
			})
		}
	}
	return block, txs, txins, txouts
}
