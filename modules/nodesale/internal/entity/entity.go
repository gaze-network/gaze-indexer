package entity

import "time"

type Block struct {
	BlockHeight int64
	BlockHash   string
	Module      string
}

type Node struct {
	SaleBlock      uint64
	SaleTxIndex    uint32
	NodeID         uint32
	TierIndex      int32
	DelegatedTo    string
	OwnerPublicKey string
	PurchaseTxHash string
	DelegateTxHash string
}

type NodeSale struct {
	BlockHeight           uint64
	TxIndex               uint32
	Name                  string
	StartsAt              time.Time
	EndsAt                time.Time
	Tiers                 [][]byte
	SellerPublicKey       string
	MaxPerAddress         uint32
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
	ExpectedTotalAmountDiscounted uint64
	ReportedTotalAmount           uint64
	PaidTotalAmount               uint64
}
