package datagateway

import (
	"context"
	"time"

	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
)

type NodesaleDataGateway interface {
	BeginNodesaleTx(ctx context.Context) (NodesaleDataGatewayWithTx, error)
	AddBlock(ctx context.Context, arg entity.Block) error
	GetBlock(ctx context.Context, blockHeight int64) (*entity.Block, error)
	GetLastProcessedBlock(ctx context.Context) (*entity.Block, error)
	RemoveBlockFrom(ctx context.Context, fromBlock int64) (int64, error)
	RemoveEventsFromBlock(ctx context.Context, fromBlock int64) (int64, error)
	ClearDelegate(ctx context.Context) (int64, error)
	GetNodes(ctx context.Context, arg GetNodesParams) ([]entity.Node, error)
	AddEvent(ctx context.Context, arg AddEventParams) error
	SetDelegates(ctx context.Context, arg SetDelegatesParams) (int64, error)
	AddNodesale(ctx context.Context, arg AddNodesaleParams) error
	GetNodesale(ctx context.Context, arg GetNodesaleParams) ([]entity.NodeSale, error)
	GetNodesByOwner(ctx context.Context, arg GetNodesByOwnerParams) ([]entity.Node, error)
	AddNode(ctx context.Context, arg AddNodeParams) error
	GetNodeCountByTierIndex(ctx context.Context, arg GetNodeCountByTierIndexParams) ([]GetNodeCountByTierIndexRow, error)
	GetNodesByPubkey(ctx context.Context, arg GetNodesByPubkeyParams) ([]entity.Node, error)
	GetEventsByWallet(ctx context.Context, walletAddress string) ([]entity.Event, error)
}

type NodesaleDataGatewayWithTx interface {
	NodesaleDataGateway
	Tx
}

type GetNodesParams struct {
	SaleBlock   int64
	SaleTxIndex int32
	NodeIds     []int32
}

type AddEventParams struct {
	TxHash         string
	BlockHeight    int64
	TxIndex        int32
	WalletAddress  string
	Valid          bool
	Action         int32
	RawMessage     []byte
	ParsedMessage  []byte
	BlockTimestamp time.Time
	BlockHash      string
	Metadata       []byte
}

type SetDelegatesParams struct {
	SaleBlock   int64
	SaleTxIndex int32
	Delegatee   string
	NodeIds     []int32
}

type AddNodesaleParams struct {
	BlockHeight           int64
	TxIndex               int32
	Name                  string
	StartsAt              time.Time
	EndsAt                time.Time
	Tiers                 [][]byte
	SellerPublicKey       string
	MaxPerAddress         int32
	DeployTxHash          string
	MaxDiscountPercentage int32
	SellerWallet          string
}

type GetNodesaleParams struct {
	BlockHeight int64
	TxIndex     int32
}

type GetNodesByOwnerParams struct {
	SaleBlock      int64
	SaleTxIndex    int32
	OwnerPublicKey string
}

type AddNodeParams struct {
	SaleBlock      int64
	SaleTxIndex    int32
	NodeID         int32
	TierIndex      int32
	DelegatedTo    string
	OwnerPublicKey string
	PurchaseTxHash string
	DelegateTxHash string
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
