package entity

import "github.com/btcsuite/btcd/chaincfg/chainhash"

type IndexedBlock struct {
	Height              uint64
	Hash                chainhash.Hash
	EventHash           chainhash.Hash
	CumulativeEventHash chainhash.Hash
}
