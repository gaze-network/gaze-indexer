package datagateway

import (
	"context"

	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
)

type NodeSaleDataGateway interface {
	BeginNodeSaleTx(ctx context.Context) (NodeSaleDataGatewayWithTx, error)
	CreateBlock(ctx context.Context, arg entity.Block) error
	GetBlock(ctx context.Context, blockHeight int64) (*entity.Block, error)
	GetLastProcessedBlock(ctx context.Context) (*entity.Block, error)
	RemoveBlockFrom(ctx context.Context, fromBlock int64) (int64, error)
	RemoveEventsFromBlock(ctx context.Context, fromBlock int64) (int64, error)
	ClearDelegate(ctx context.Context) (int64, error)
	GetNodesByIds(ctx context.Context, arg GetNodesByIdsParams) ([]entity.Node, error)
	CreateEvent(ctx context.Context, arg entity.NodeSaleEvent) error
	SetDelegates(ctx context.Context, arg SetDelegatesParams) (int64, error)
	CreateNodeSale(ctx context.Context, arg entity.NodeSale) error
	GetNodeSale(ctx context.Context, arg GetNodeSaleParams) ([]entity.NodeSale, error)
	GetNodesByOwner(ctx context.Context, arg GetNodesByOwnerParams) ([]entity.Node, error)
	CreateNode(ctx context.Context, arg entity.Node) error
	GetNodeCountByTierIndex(ctx context.Context, arg GetNodeCountByTierIndexParams) ([]GetNodeCountByTierIndexRow, error)
	GetNodesByPubkey(ctx context.Context, arg GetNodesByPubkeyParams) ([]entity.Node, error)
	GetEventsByWallet(ctx context.Context, walletAddress string) ([]entity.NodeSaleEvent, error)
}

type NodeSaleDataGatewayWithTx interface {
	NodeSaleDataGateway
	Tx
}

type GetNodesByIdsParams struct {
	SaleBlock   int64
	SaleTxIndex int32
	NodeIds     []int32
}

type SetDelegatesParams struct {
	SaleBlock      int64
	SaleTxIndex    int32
	Delegatee      string
	DelegateTxHash string
	NodeIds        []int32
}

type GetNodeSaleParams struct {
	BlockHeight int64
	TxIndex     int32
}

type GetNodesByOwnerParams struct {
	SaleBlock      int64
	SaleTxIndex    int32
	OwnerPublicKey string
}

type GetNodeCountByTierIndexParams struct {
	SaleBlock   int64
	SaleTxIndex int32
	FromTier    int32
	ToTier      int32
}

type GetNodeCountByTierIndexRow struct {
	TierIndex int32
	Count     int64
}

type GetNodesByPubkeyParams struct {
	SaleBlock      int64
	SaleTxIndex    int32
	OwnerPublicKey string
	DelegatedTo    string
}
