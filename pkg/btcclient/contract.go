package btcclient

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type Contract interface {
	GetRawTransactionAndHeightByTxHash(ctx context.Context, txHash chainhash.Hash) (*wire.MsgTx, int64, error)
}
