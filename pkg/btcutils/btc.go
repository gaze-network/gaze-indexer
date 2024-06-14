package btcutils

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
)

var (
	// NullAddress is an address that script address is all zeros.
	NullAddress = NewAddress("1111111111111111111114oLvT2", &chaincfg.MainNetParams)

	// NullHash is a hash that all bytes are zero.
	NullHash = utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000"))
)

// TransactionType is the type of bitcoin transaction
// It's an alias of txscript.ScriptClass
type TransactionType = txscript.ScriptClass

// AddressType is the type of bitcoin address.
// It's an alias of txscript.ScriptClass
type AddressType = txscript.ScriptClass

// Types of bitcoin transaction
const (
	TransactionP2WPKH  = txscript.WitnessV0PubKeyHashTy
	TransactionP2TR    = txscript.WitnessV1TaprootTy
	TransactionTaproot = TransactionP2TR // Alias of P2TR
	TransactionP2SH    = txscript.ScriptHashTy
	TransactionP2PKH   = txscript.PubKeyHashTy
	TransactionP2WSH   = txscript.WitnessV0ScriptHashTy
)

// Types of bitcoin address
const (
	AddressP2WPKH  = txscript.WitnessV0PubKeyHashTy
	AddressP2TR    = txscript.WitnessV1TaprootTy
	AddressTaproot = AddressP2TR // Alias of P2TR
	AddressP2SH    = txscript.ScriptHashTy
	AddressP2PKH   = txscript.PubKeyHashTy
	AddressP2WSH   = txscript.WitnessV0ScriptHashTy
)
