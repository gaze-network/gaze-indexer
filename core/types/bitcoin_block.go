package types

type UTXO struct {
	TxHash   string
	Index    uint64
	PkScript []byte
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
	Hash   string
	Height uint64
}
