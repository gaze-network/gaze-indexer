package types

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/samber/lo"
)

type BlockHeader struct {
	Hash       chainhash.Hash
	Height     int64
	Version    int32
	PrevBlock  chainhash.Hash
	MerkleRoot chainhash.Hash
	Timestamp  time.Time
	Bits       uint32
	Nonce      uint32
}

func ParseMsgBlockHeader(src wire.BlockHeader, height int64) BlockHeader {
	hash := src.BlockHash()
	return BlockHeader{
		Hash:       hash,
		Height:     height,
		Version:    src.Version,
		PrevBlock:  src.PrevBlock,
		MerkleRoot: src.MerkleRoot,
		Timestamp:  src.Timestamp,
		Bits:       src.Bits,
		Nonce:      src.Nonce,
	}
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
}

func ParseMsgBlock(src *wire.MsgBlock, height int64) *Block {
	hash := src.Header.BlockHash()
	return &Block{
		Header:       ParseMsgBlockHeader(src.Header, height),
		Transactions: lo.Map(src.Transactions, func(item *wire.MsgTx, index int) *Transaction { return ParseMsgTx(item, height, hash, uint32(index)) }),
	}
}
