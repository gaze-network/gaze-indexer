package nodesale

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
)

/*
func (p *Processor) pubkeyToTaprootAddress(pubkey string, script []byte) (btcutil.Address, error) {
	pubKeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode string : %w", err)
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse pubkey : %w", err)
	}

	tapleaf := txscript.NewBaseTapLeaf(script)

	scriptTree := txscript.AssembleTaprootScriptTree(tapleaf)
	rootHash := scriptTree.RootNode.TapHash()

	tapkey := txscript.ComputeTaprootOutputKey(pubKey, rootHash[:])

	sellerAddr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(tapkey), p.network.ChainParams())
	if err != nil {
		return nil, fmt.Errorf("invalid taproot address: %w", err)
	}
	return sellerAddr, nil
}*/

func (p *Processor) pubkeyToPkHashAddress(pubKey *btcec.PublicKey) btcutil.Address {
	// pubKeyBytes, _ := hex.DecodeString(pubkey)
	// pubKey, _ := btcec.ParsePubKey(pubKeyBytes)
	addrPubKey, _ := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(), p.network.ChainParams())
	addrPubKeyHash := addrPubKey.AddressPubKeyHash()
	return addrPubKeyHash
}
