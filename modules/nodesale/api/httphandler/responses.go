package httphandler

import (
	"encoding/json"
	"time"
)

type infoResponse struct {
	IndexedBlockHeight int32  `json:"indexedBlockHeight"`
	IndexedBlockHash   string `json:"indexedBlockHash"`
}

type deployResponse struct {
	Id              string         `json:"id"`
	Name            string         `json:"name"`
	StartAt         time.Time      `json:"startAt"`
	EndAt           time.Time      `json:"EndAt"`
	Tiers           []tierResponse `json:"tiers"`
	SellerPublicKey string         `json:"sellerPublicKey"`
	MaxPerAddress   int32          `json:"maxPerAddress"`
	DeployTxHash    string         `json:"deployTxHash"`
}

type tierResponse struct {
	PriceSat      uint32 `json:"priceSat"`
	Limit         uint32 `json:"limit"`
	MaxPerAddress uint32 `json:"maxPerAddress"`
	Sold          int64  `json:"sold"`
}

type nodeResponse struct {
	DeployId            string `json:"deployId"`
	NodeId              int32  `json:"nodeId"`
	TierIndex           int32  `json:"tierIndex"`
	DelegatedTo         string `json:"delegatedTo"`
	OwnerPublicKey      string `json:"ownerPublicKey"`
	PurchaseTxHash      string `json:"purchaseTxHash"`
	DelegateTxHash      string `json:"delegateTxHash"`
	PurchaseBlockHeight int32  `json:"purchaseBlockHeight"`
}

type eventResposne struct {
	TxHash         string          `json:"txHash"`
	BlockHeight    int32           `json:"blockHeight"`
	TxIndex        int32           `json:"txIndex"`
	WalletAddress  string          `json:"walletAddress"`
	Action         string          `json:"action"`
	ParsedMessage  json.RawMessage `json:"parsedMessage"`
	BlockTimestamp time.Time       `json:"blockTimestamp"`
	BlockHash      string          `json:"blockHash"`
}
