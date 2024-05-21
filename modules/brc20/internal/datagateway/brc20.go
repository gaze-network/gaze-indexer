package datagateway

import (
	"context"
)

type BRC20DataGateway interface {
	BRC20ReaderDataGateway
	BRC20WriterDataGateway

	// BeginBRC20Tx returns a new BRC20DataGateway with transaction enabled. All write operations performed in this datagateway must be committed to persist changes.
	BeginBRC20Tx(ctx context.Context) (BRC20DataGatewayWithTx, error)
}

type BRC20DataGatewayWithTx interface {
	BRC20DataGateway
	Tx
}

type BRC20ReaderDataGateway interface {
	// TODO: add methods
}

type BRC20WriterDataGateway interface {
	// TODO: add methods
}
