package nodesale

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
)

func (p *Processor) PubkeyToPkHashAddress(pubKey *btcec.PublicKey) btcutil.Address {
	addrPubKey, _ := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(), p.Network.ChainParams())
	addrPubKeyHash := addrPubKey.AddressPubKeyHash()
	return addrPubKeyHash
}
