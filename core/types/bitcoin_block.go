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

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
}

func ParseMsgBlock(src *wire.MsgBlock) *Block {
	return &Block{
		Header: BlockHeader{
			Hash:       src.Header.BlockHash(),
			Version:    src.Header.Version,
			PrevBlock:  src.Header.PrevBlock,
			MerkleRoot: src.Header.MerkleRoot,
			Timestamp:  src.Header.Timestamp,
			Bits:       src.Header.Bits,
			Nonce:      src.Header.Nonce,
		},
		Transactions: lo.Map(src.Transactions, func(item *wire.MsgTx, _ int) *Transaction { return ParseMsgTx(item) }),
	}
}
