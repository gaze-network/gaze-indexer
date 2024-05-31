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

type Brc20DeployEvent struct {
	Id            int64
	InscriptionID string
	Tick          string
	OriginalTick  string
	TxHash        string
	BlockHeight   int32
	TxIndex       int32
	Timestamp     pgtype.Timestamp
	Pkscript      string
	TotalSupply   pgtype.Numeric
	Decimals      int16
	LimitPerMint  pgtype.Numeric
	IsSelfMint    bool
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

type Brc20MintEvent struct {
	Id            int64
	InscriptionID string
	Tick          string
	OriginalTick  string
	TxHash        string
	BlockHeight   int32
	TxIndex       int32
	Timestamp     pgtype.Timestamp
	Pkscript      string
	Amount        pgtype.Numeric
	ParentID      pgtype.Text
}

type Brc20ProcessorStat struct {
	BlockHeight             int32
	CursedInscriptionCount  int32
	BlessedInscriptionCount int32
	LostSats                int64
}

type Brc20Tick struct {
	Tick                string
	OriginalTick        string
	TotalSupply         pgtype.Numeric
	Decimals            int16
	LimitPerMint        pgtype.Numeric
	IsSelfMint          bool
	DeployInscriptionID string
	CreatedAt           pgtype.Timestamp
	CreatedAtHeight     int32
}

type Brc20TickState struct {
	Tick              string
	BlockHeight       int32
	MintedAmount      pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

type Brc20TransferEvent struct {
	Id            int64
	InscriptionID string
	Tick          string
	OriginalTick  string
	TxHash        string
	BlockHeight   int32
	TxIndex       int32
	Timestamp     pgtype.Timestamp
	FromPkscript  pgtype.Text
	ToPkscript    string
	Amount        pgtype.Numeric
}
