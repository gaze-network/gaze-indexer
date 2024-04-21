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

func mapBlockTypeToParams(src *types.Block) (gen.InsertBlockParams, gen.BatchInsertTransactionsParams, gen.BatchInsertTransactionTxOutsParams, gen.BatchInsertTransactionTxInsParams) {
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
	txs := gen.BatchInsertTransactionsParams{
		TxHashArr:      []string{},
		VersionArr:     []int32{},
		LocktimeArr:    []int64{},
		BlockHeightArr: []int32{},
		BlockHashArr:   []string{},
		IdxArr:         []int32{},
	}
	txouts := gen.BatchInsertTransactionTxOutsParams{
		TxHashArr:   []string{},
		TxIdxArr:    []int32{},
		PkscriptArr: []string{},
		ValueArr:    []int64{},
	}
	txins := gen.BatchInsertTransactionTxInsParams{
		PrevoutTxHashArr: []string{},
		PrevoutTxIdxArr:  []int32{},
		TxHashArr:        []string{},
		TxIdxArr:         []int32{},
		ScriptsigArr:     []string{},
		WitnessArr:       []string{},
		SequenceArr:      []int32{},
	}

	for txIdx, srcTx := range src.Transactions {
		txHash := srcTx.TxHash.String()
		// Batch insert transactions
		txs.TxHashArr = append(txs.TxHashArr, txHash)
		txs.VersionArr = append(txs.VersionArr, srcTx.Version)
		txs.LocktimeArr = append(txs.LocktimeArr, int64(srcTx.LockTime))
		txs.BlockHeightArr = append(txs.BlockHeightArr, int32(src.Header.Height))
		txs.BlockHashArr = append(txs.BlockHashArr, src.Header.Hash.String())
		txs.IdxArr = append(txs.IdxArr, int32(txIdx))

		// Batch insert txins
		for idx, txin := range srcTx.TxIn {
			var witness string
			if len(txin.Witness) > 0 {
				witness = btcutils.WitnessToString(txin.Witness)
			}
			txins.TxHashArr = append(txins.TxHashArr, txHash)
			txins.TxIdxArr = append(txins.TxIdxArr, int32(idx))
			txins.PrevoutTxHashArr = append(txins.PrevoutTxHashArr, txin.PreviousOutTxHash.String())
			txins.PrevoutTxIdxArr = append(txins.PrevoutTxIdxArr, int32(txin.PreviousOutIndex))
			txins.ScriptsigArr = append(txins.ScriptsigArr, hex.EncodeToString(txin.SignatureScript))
			txins.WitnessArr = append(txins.WitnessArr, witness)
			txins.SequenceArr = append(txins.SequenceArr, int32(txin.Sequence))
		}

		// Batch insert txouts
		for idx, txout := range srcTx.TxOut {
			txouts.TxHashArr = append(txouts.TxHashArr, txHash)
			txouts.TxIdxArr = append(txouts.TxIdxArr, int32(idx))
			txouts.PkscriptArr = append(txouts.PkscriptArr, hex.EncodeToString(txout.PkScript))
			txouts.ValueArr = append(txouts.ValueArr, txout.Value)
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

		witness, err := btcutils.WitnessFromString(txInModel.Witness)
		if err != nil {
			return types.Transaction{}, errors.Wrap(err, "failed to parse witness from hex string")
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
