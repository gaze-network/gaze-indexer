package entity

import "time"

type Block struct {
	BlockHeight int64
	BlockHash   string
	Module      string
}

type Node struct {
	SaleBlock      int64
	SaleTxIndex    int32
	NodeID         int32
	TierIndex      int32
	DelegatedTo    string
	OwnerPublicKey string
	PurchaseTxHash string
	DelegateTxHash string
}

type NodeSale struct {
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

type NodeSaleEvent struct {
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
	Metadata       *MetadataEventPurchase
	Reason         string
}

type MetadataEventPurchase struct {
	ExpectedTotalAmountDiscounted int64
	ReportedTotalAmount           int64
	PaidTotalAmount               int64
}
