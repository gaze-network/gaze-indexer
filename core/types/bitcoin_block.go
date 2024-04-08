package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type BlockHeader struct {
	Hash   *chainhash.Hash
	Height int64
	wire.BlockHeader
}

type Block struct {
	BlockHeader
	Transactions []*wire.MsgTx
}

func (b *Block) MsgBlock() *wire.MsgBlock {
	return &wire.MsgBlock{
		Header:       b.BlockHeader.BlockHeader,
		Transactions: b.Transactions,
	}
}
