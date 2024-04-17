package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/core/types"
)

type BitcoinDataGateway interface {
	BitcoinWriterDataDataGateway
	BitcoinReaderDataDataGateway
}

type BitcoinWriterDataDataGateway interface {
	InsertBlock(context.Context, *types.Block) error
	RevertBlocks(context.Context, int64) error
}

type BitcoinReaderDataDataGateway interface {
	GetLatestBlockHeader(context.Context) (types.BlockHeader, error)
	GetBlockHeaderByHeight(ctx context.Context, blockHeight int64) (types.BlockHeader, error)
	GetBlocksByHeightRange(ctx context.Context, from int64, to int64) ([]*types.Block, error)
	GetTransactionByHash(ctx context.Context, txHash chainhash.Hash) (*types.Transaction, error)
}
