package entity

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type OutPointBalance struct {
	PkScript   []byte
	Id         runes.RuneId
	Amount     uint128.Uint128
	Index      uint32
	TxHash     chainhash.Hash
	TxOutIndex uint32
}

type outpointBalanceJSON struct {
	PkScript   string          `json:"pkScript"`
	Id         runes.RuneId    `json:"id"`
	Amount     uint128.Uint128 `json:"amount"`
	Index      uint32          `json:"index"`
	TxHash     chainhash.Hash  `json:"txHash"`
	TxOutIndex uint32          `json:"txOutIndex"`
}

func (o OutPointBalance) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(outpointBalanceJSON{
		PkScript:   hex.EncodeToString(o.PkScript),
		Id:         o.Id,
		Amount:     o.Amount,
		Index:      o.Index,
		TxHash:     o.TxHash,
		TxOutIndex: o.TxOutIndex,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return bytes, nil
}

func (o *OutPointBalance) UnmarshalJSON(data []byte) error {
	var aux outpointBalanceJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return errors.WithStack(err)
	}
	pkScript, err := hex.DecodeString(aux.PkScript)
	if err != nil {
		return errors.WithStack(err)
	}
	o.PkScript = pkScript
	o.Id = aux.Id
	o.Amount = aux.Amount
	o.Index = aux.Index
	o.TxHash = aux.TxHash
	o.TxOutIndex = aux.TxOutIndex
	return nil
}

// TODO: add tx index
type RuneTransaction struct {
	Hash        chainhash.Hash
	BlockHeight uint64
	Index       uint32
	Timestamp   time.Time
	Inputs      []*OutPointBalance
	Outputs     []*OutPointBalance
	Mints       map[runes.RuneId]uint128.Uint128
	Burns       map[runes.RuneId]uint128.Uint128
	Runestone   *runes.Runestone
}
