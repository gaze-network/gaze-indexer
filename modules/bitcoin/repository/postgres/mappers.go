package postgres

import (
	"cmp"
	"encoding/hex"
	"slices"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
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
		Bits:  int64(src.Bits),
		Nonce: int64(src.Nonce),
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
		Bits:  int64(src.Header.Bits),
		Nonce: int64(src.Header.Nonce),
	}
	for txIdx, srcTx := range src.Transactions {
		tx := gen.InsertTransactionParams{
			TxHash:      srcTx.TxHash.String(),
			Version:     srcTx.Version,
			Locktime:    int64(srcTx.LockTime),
			BlockHeight: int32(src.Header.Height),
			BlockHash:   src.Header.Hash.String(),
			Idx:         int16(txIdx),
		}
		txs = append(txs, tx)

		for idx, txin := range srcTx.TxIn {
			var witness pgtype.Text
			if len(txin.Witness) > 0 {
				witness = pgtype.Text{
					String: btcutils.WitnessToString(txin.Witness),
					Valid:  true,
				}
			}
			txins = append(txins, gen.InsertTransactionTxInParams{
				TxHash:        tx.TxHash,
				TxIdx:         int16(idx),
				PrevoutTxHash: txin.PreviousOutTxHash.String(),
				PrevoutTxIdx:  int16(txin.PreviousOutIndex),
				Scriptsig:     hex.EncodeToString(txin.SignatureScript),
				Witness:       witness,
				Sequence:      int64(txin.Sequence),
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

func mapTransactionModelToType(src gen.BitcoinTransaction, txInModel []gen.BitcoinTransactionTxin, txOutModels []gen.BitcoinTransactionTxout) (types.Transaction, error) {
	blockHash, err := chainhash.NewHashFromStr(src.BlockHash)
	if err != nil {
		return types.Transaction{}, errors.Wrap(err, "failed to parse block hash")
	}

	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return types.Transaction{}, errors.Wrap(err, "failed to parse tx hash")
	}

	// Sort txins and txouts by index (Asc)
	slices.SortFunc(txOutModels, func(i, j gen.BitcoinTransactionTxout) int {
		return cmp.Compare(i.TxIdx, j.TxIdx)
	})
	slices.SortFunc(txInModel, func(i, j gen.BitcoinTransactionTxin) int {
		return cmp.Compare(i.TxIdx, j.TxIdx)
	})

	txIns := make([]*types.TxIn, 0, len(txInModel))
	txOuts := make([]*types.TxOut, 0, len(txOutModels))
	for _, txInModel := range txInModel {
		scriptsig, err := hex.DecodeString(txInModel.Scriptsig)
		if err != nil {
			return types.Transaction{}, errors.Wrap(err, "failed to decode scriptsig")
		}

		prevoutTxHash, err := chainhash.NewHashFromStr(txInModel.PrevoutTxHash)
		if err != nil {
			return types.Transaction{}, errors.Wrap(err, "failed to parse prevout tx hash")
		}

		var witness [][]byte
		if txInModel.Witness.Valid {
			w, err := btcutils.WitnessFromString(txInModel.Witness.String)
			if err != nil {
				return types.Transaction{}, errors.Wrap(err, "failed to parse witness from hex string")
			}
			witness = w
		}

		txIns = append(txIns, &types.TxIn{
			SignatureScript:   scriptsig,
			Witness:           witness,
			Sequence:          uint32(txInModel.Sequence),
			PreviousOutIndex:  uint32(txInModel.PrevoutTxIdx),
			PreviousOutTxHash: *prevoutTxHash,
		})
	}
	for _, txOutModel := range txOutModels {
		pkscript, err := hex.DecodeString(txOutModel.Pkscript)
		if err != nil {
			return types.Transaction{}, errors.Wrap(err, "failed to decode pkscript")
		}
		txOuts = append(txOuts, &types.TxOut{
			PkScript: pkscript,
			Value:    txOutModel.Value,
		})
	}

	return types.Transaction{
		BlockHeight: int64(src.BlockHeight),
		BlockHash:   *blockHash,
		Index:       uint32(src.Idx),
		TxHash:      *txHash,
		Version:     src.Version,
		LockTime:    uint32(src.Locktime),
		TxIn:        txIns,
		TxOut:       txOuts,
	}, nil
}
