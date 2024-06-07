// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package gen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Brc20Balance struct {
	Pkscript         string
	BlockHeight      int32
	Tick             string
	OverallBalance   pgtype.Numeric
	AvailableBalance pgtype.Numeric
}

type Brc20EventDeploy struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	Pkscript          string
	Satpoint          string
	TotalSupply       pgtype.Numeric
	Decimals          int16
	LimitPerMint      pgtype.Numeric
	IsSelfMint        bool
}

type Brc20EventInscribeTransfer struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	Pkscript          string
	Satpoint          string
	OutputIndex       int32
	SatsAmount        int64
	Amount            pgtype.Numeric
}

type Brc20EventMint struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	Pkscript          string
	Satpoint          string
	Amount            pgtype.Numeric
	ParentID          pgtype.Text
}

type Brc20EventTransferTransfer struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	FromPkscript      string
	FromSatpoint      string
	FromInputIndex    int32
	ToPkscript        string
	ToSatpoint        string
	ToOutputIndex     int32
	Amount            pgtype.Numeric
}

type Brc20IndexedBlock struct {
	Height              int32
	Hash                string
	EventHash           string
	CumulativeEventHash string
}

type Brc20IndexerState struct {
	Id               int64
	ClientVersion    string
	Network          string
	DbVersion        int32
	EventHashVersion int32
	CreatedAt        pgtype.Timestamptz
}

type Brc20InscriptionEntry struct {
	Id              string
	Number          int64
	SequenceNumber  int64
	Delegate        pgtype.Text
	Metadata        []byte
	Metaprotocol    pgtype.Text
	Parents         []string
	Pointer         pgtype.Int8
	Content         []byte
	ContentEncoding pgtype.Text
	ContentType     pgtype.Text
	Cursed          bool
	CursedForBrc20  bool
	CreatedAt       pgtype.Timestamp
	CreatedAtHeight int32
}

type Brc20InscriptionEntryState struct {
	Id            string
	BlockHeight   int32
	TransferCount int32
}

type Brc20InscriptionTransfer struct {
	InscriptionID     string
	BlockHeight       int32
	TxIndex           int32
	OldSatpointTxHash pgtype.Text
	OldSatpointOutIdx pgtype.Int4
	OldSatpointOffset pgtype.Int8
	NewSatpointTxHash pgtype.Text
	NewSatpointOutIdx pgtype.Int4
	NewSatpointOffset pgtype.Int8
	NewPkscript       string
	NewOutputValue    int64
	SentAsFee         bool
}

type Brc20ProcessorStat struct {
	BlockHeight             int32
	CursedInscriptionCount  int32
	BlessedInscriptionCount int32
	LostSats                int64
}

type Brc20TickEntry struct {
	Tick                string
	OriginalTick        string
	TotalSupply         pgtype.Numeric
	Decimals            int16
	LimitPerMint        pgtype.Numeric
	IsSelfMint          bool
	DeployInscriptionID string
	DeployedAt          pgtype.Timestamp
	DeployedAtHeight    int32
}

type Brc20TickEntryState struct {
	Tick              string
	BlockHeight       int32
	MintedAmount      pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}
