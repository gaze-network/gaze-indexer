// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package gen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Block struct {
	BlockHeight int32
	BlockHash   string
	Module      string
}

type Event struct {
	TxHash         string
	BlockHeight    int32
	TxIndex        int32
	WalletAddress  string
	Valid          bool
	Action         int32
	RawMessage     []byte
	ParsedMessage  []byte
	BlockTimestamp pgtype.Timestamp
	BlockHash      string
	Metadata       []byte
}

type Node struct {
	SaleBlock      int32
	SaleTxIndex    int32
	NodeID         int32
	TierIndex      int32
	DelegatedTo    string
	OwnerPublicKey string
	PurchaseTxHash string
	DelegateTxHash string
}

type NodeSale struct {
	BlockHeight           int32
	TxIndex               int32
	Name                  string
	StartsAt              pgtype.Timestamp
	EndsAt                pgtype.Timestamp
	Tiers                 [][]byte
	SellerPublicKey       string
	MaxPerAddress         int32
	DeployTxHash          string
	MaxDiscountPercentage int32
	SellerWallet          string
}
