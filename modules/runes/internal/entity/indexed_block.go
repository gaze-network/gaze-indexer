package entity

import "github.com/btcsuite/btcd/chaincfg/chainhash"

type IndexedBlock struct {
	Height              int64
	Hash                chainhash.Hash
	PrevHash            chainhash.Hash
	EventHash           chainhash.Hash
	CumulativeEventHash chainhash.Hash
}
