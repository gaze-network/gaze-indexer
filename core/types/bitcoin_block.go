package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type UTXO struct {
	TxHash   string
	Index    uint64
	PkScript []byte
	Amount   float64
}

type TxIn struct {
	PreviousUTXO UTXO
}

type TxOut struct {
	PkScript []byte
}

type Transaction struct {
	Hash   string
	TxIns  []TxIn
	TxOuts []TxOut
}

type Block struct {
	BlockHeader
	Transactions []Transaction
}

type BlockHeader struct {
	Hash   *chainhash.Hash
	Height int64
	wire.BlockHeader
}
