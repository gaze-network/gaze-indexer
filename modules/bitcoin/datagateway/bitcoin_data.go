package datagateway

import (
	"context"

	"github.com/gaze-network/indexer-network/core/types"
)

type BitcoinDataGateway interface {
	BitcoinWriterDataDataGateway
	BitcoinReaderDataDataGateway
}

type BitcoinWriterDataDataGateway interface {
	InsertBlock(context.Context, *types.Block) error
	ReverBlocks(context.Context, int64) error
}

type BitcoinReaderDataDataGateway interface {
	GetLatestBlockHeader(context.Context) (types.BlockHeader, error)
}
