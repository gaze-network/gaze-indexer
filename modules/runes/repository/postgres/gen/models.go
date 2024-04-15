// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package gen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type RunesBalance struct {
	Pkscript    string
	BlockHeight int32
	RuneID      string
	Amount      pgtype.Numeric
}

type RunesEntry struct {
	RuneID           string
	Rune             string
	Spacers          int32
	Premine          pgtype.Numeric
	Symbol           int32
	Divisibility     int16
	Terms            bool
	TermsAmount      pgtype.Numeric
	TermsCap         pgtype.Numeric
	TermsHeightStart pgtype.Int4
	TermsHeightEnd   pgtype.Int4
	TermsOffsetStart pgtype.Int4
	TermsOffsetEnd   pgtype.Int4
	Turbo            bool
	EtchingBlock     int32
	EtchingTxHash    string
}

type RunesEntryState struct {
	RuneID            string
	BlockHeight       int32
	Mints             pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

type RunesIndexedBlock struct {
	Height              int32
	Hash                string
	PrevHash            string
	EventHash           string
	CumulativeEventHash string
}

type RunesIndexerStat struct {
	Id            int64
	ClientVersion string
	Network       string
	CreatedAt     pgtype.Timestamptz
}

type RunesIndexerState struct {
	Id               int64
	DbVersion        int32
	EventHashVersion int32
	CreatedAt        pgtype.Timestamptz
}

type RunesOutpointBalance struct {
	RuneID      string
	TxHash      string
	TxIdx       int32
	Amount      pgtype.Numeric
	BlockHeight int32
	SpentHeight pgtype.Int4
}

type RunesRunestone struct {
	TxHash                  string
	BlockHeight             int32
	Etching                 bool
	EtchingDivisibility     pgtype.Int2
	EtchingPremine          pgtype.Numeric
	EtchingRune             pgtype.Text
	EtchingSpacers          pgtype.Int4
	EtchingSymbol           pgtype.Int4
	EtchingTerms            pgtype.Bool
	EtchingTermsAmount      pgtype.Numeric
	EtchingTermsCap         pgtype.Numeric
	EtchingTermsHeightStart pgtype.Int4
	EtchingTermsHeightEnd   pgtype.Int4
	EtchingTermsOffsetStart pgtype.Int4
	EtchingTermsOffsetEnd   pgtype.Int4
	EtchingTurbo            pgtype.Bool
	Edicts                  []byte
	Mint                    pgtype.Text
	Pointer                 pgtype.Int4
	Cenotaph                bool
	Flaws                   int32
}

type RunesTransaction struct {
	Hash        string
	BlockHeight int32
	Index       int32
	Timestamp   pgtype.Timestamp
	Inputs      []byte
	Outputs     []byte
	Mints       []byte
	Burns       []byte
}
