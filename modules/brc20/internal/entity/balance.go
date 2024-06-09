package entity

import "github.com/shopspring/decimal"

type Balance struct {
	PkScript         []byte
	Tick             string
	BlockHeight      uint64
	OverallBalance   decimal.Decimal
	AvailableBalance decimal.Decimal
}
