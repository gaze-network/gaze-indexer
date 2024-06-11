// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package gen

import (
	"context"
)

type Querier interface {
	AddBlock(ctx context.Context, arg AddBlockParams) error
	AddEvent(ctx context.Context, arg AddEventParams) error
	AddNode(ctx context.Context, arg AddNodeParams) error
	AddNodesale(ctx context.Context, arg AddNodesaleParams) error
	ClearDelegate(ctx context.Context) (int64, error)
	ClearEvents(ctx context.Context) error
	GetBlock(ctx context.Context, blockHeight int32) (Block, error)
	GetEventsByWallet(ctx context.Context, walletAddress string) ([]Event, error)
	GetLastProcessedBlock(ctx context.Context) (Block, error)
	GetNodeCountByTierIndex(ctx context.Context, arg GetNodeCountByTierIndexParams) ([]GetNodeCountByTierIndexRow, error)
	GetNodes(ctx context.Context, arg GetNodesParams) ([]Node, error)
	GetNodesByOwner(ctx context.Context, arg GetNodesByOwnerParams) ([]Node, error)
	GetNodesByPubkey(ctx context.Context, arg GetNodesByPubkeyParams) ([]GetNodesByPubkeyRow, error)
	GetNodesale(ctx context.Context, arg GetNodesaleParams) ([]NodeSale, error)
	RemoveBlockFrom(ctx context.Context, fromBlock int32) (int64, error)
	RemoveEventsFromBlock(ctx context.Context, fromBlock int32) (int64, error)
	SetDelegates(ctx context.Context, arg SetDelegatesParams) (int64, error)
}

var _ Querier = (*Queries)(nil)
