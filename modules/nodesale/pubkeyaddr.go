package nodesale

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
)

func (p *Processor) pubkeyToAddress(pubkey string) btcutil.Address {
	pubKeyBytes, _ := hex.DecodeString(pubkey)
	pubKey, _ := btcec.ParsePubKey(pubKeyBytes)
	sellerAddr, _ := btcutil.NewAddressTaproot(schnorr.SerializePubKey(pubKey), p.network.ChainParams())
	return sellerAddr
}
