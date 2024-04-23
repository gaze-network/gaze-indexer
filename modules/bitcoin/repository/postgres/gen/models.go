// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package gen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type BitcoinBlock struct {
	BlockHeight   int32
	BlockHash     string
	Version       int32
	MerkleRoot    string
	PrevBlockHash string
	Timestamp     pgtype.Timestamptz
	Bits          int64
	Nonce         int64
}

type BitcoinIndexerDbVersion struct {
	Id        int64
	Version   int32
	CreatedAt pgtype.Timestamptz
}

type BitcoinIndexerStat struct {
	Id            int64
	ClientVersion string
	Network       string
	CreatedAt     pgtype.Timestamptz
}

type BitcoinTransaction struct {
	TxHash      string
	Version     int32
	Locktime    int64
	BlockHeight int32
	BlockHash   string
	Idx         int32
}

type BitcoinTransactionTxin struct {
	TxHash          string
	TxIdx           int32
	PrevoutTxHash   string
	PrevoutTxIdx    int32
	PrevoutPkscript pgtype.Text
	Scriptsig       string
	Witness         string
	Sequence        int64
}

type BitcoinTransactionTxout struct {
	TxHash   string
	TxIdx    int32
	Pkscript string
	Value    int64
	IsSpent  bool
}
