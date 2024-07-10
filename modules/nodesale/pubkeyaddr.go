package nodesale

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
)

func (p *Processor) pubkeyToPkHashAddress(pubKey *btcec.PublicKey) btcutil.Address {
	// pubKeyBytes, _ := hex.DecodeString(pubkey)
	// pubKey, _ := btcec.ParsePubKey(pubKeyBytes)
	addrPubKey, _ := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(), p.network.ChainParams())
	addrPubKeyHash := addrPubKey.AddressPubKeyHash()
	return addrPubKeyHash
}
