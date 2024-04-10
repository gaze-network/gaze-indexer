package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/samber/lo"
)

type Transaction struct {
	BlockHeight int64
	BlockHash   chainhash.Hash
	TxHash      chainhash.Hash
	Version     int32
	LockTime    uint32
	TxIn        []*TxIn
	TxOut       []*TxOut
}

type TxIn struct {
	SignatureScript   []byte
	Witness           [][]byte
	Sequence          uint32
	PreviousOutIndex  uint32
	PreviousOutTxHash chainhash.Hash
}

type TxOut struct {
	PkScript []byte
	Value    int64
}

// ParseMsgTx parses btcd/wire.MsgTx to Transaction.
func ParseMsgTx(src *wire.MsgTx, blockHeight int64, blockHash chainhash.Hash) *Transaction {
	return &Transaction{
		BlockHeight: blockHeight,
		BlockHash:   blockHash,
		TxHash:      src.TxHash(),
		Version:     src.Version,
		LockTime:    src.LockTime,
		TxIn: lo.Map(src.TxIn, func(item *wire.TxIn, _ int) *TxIn {
			return ParseTxIn(item)
		}),
		TxOut: lo.Map(src.TxOut, func(item *wire.TxOut, _ int) *TxOut {
			return ParseTxOut(item)
		}),
	}
}

// ParseTxIn parses btcd/wire.TxIn to TxIn.
func ParseTxIn(src *wire.TxIn) *TxIn {
	return &TxIn{
		SignatureScript:   src.SignatureScript,
		Witness:           src.Witness,
		Sequence:          src.Sequence,
		PreviousOutIndex:  src.PreviousOutPoint.Index,
		PreviousOutTxHash: src.PreviousOutPoint.Hash,
	}
}

// ParseTxOut parses btcd/wire.TxOut to TxOut.
func ParseTxOut(src *wire.TxOut) *TxOut {
	return &TxOut{
		PkScript: src.PkScript,
		Value:    src.Value,
	}
}
