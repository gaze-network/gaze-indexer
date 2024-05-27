package btcclient

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/core/types"
)

type Contract interface {
	GetTransactionByHash(ctx context.Context, txHash chainhash.Hash) (*types.Transaction, error)
	GetTransactionOutputs(ctx context.Context, txHash chainhash.Hash) ([]*types.TxOut, error)
}
