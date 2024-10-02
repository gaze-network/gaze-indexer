package btcutils

import (
	"encoding/json"
	"reflect"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

const (
	// MaxSupportedPkScriptSize is the maximum supported size of a pkScript.
	MaxSupportedPkScriptSize = 40
)

// IsAddress returns whether or not the passed string is a valid bitcoin address and valid supported type.
//
// NetParams is optional. If provided, we only check for that network,
// otherwise, we check for all supported networks.
func IsAddress(address string, defaultNet ...*chaincfg.Params) bool {
	if len(address) == 0 {
		return false
	}

	// If defaultNet is provided, we only check for that network.
	net, ok := utils.Optional(defaultNet)
	if ok {
		_, _, err := parseAddress(address, net)
		return err == nil
	}

	// Otherwise, we check for all supported networks.
	for _, net := range supportedNetworks {
		_, _, err := parseAddress(address, net)
		if err == nil {
			return true
		}
	}
	return false
}

// TODO: create GetAddressNetwork
// check `Bech32HRPSegwit` prefix or netID for P2SH/P2PKH is equal to `PubKeyHashAddrID/ScriptHashAddrID`

// GetAddressType returns the address type of the passed address.
func GetAddressType(address string, net *chaincfg.Params) (AddressType, error) {
	_, addrType, err := parseAddress(address, net)
	return addrType, errors.WithStack(err)
}

type Address struct {
	decoded          btcutil.Address
	net              *chaincfg.Params
	encoded          string
	encodedType      AddressType
	scriptPubKey     [MaxSupportedPkScriptSize]byte
	scriptPubKeySize int
}

// NewAddress creates a new address from the given address string.
//
// defaultNet is required if your address is P2SH or P2PKH (legacy or nested segwit)
// If your address is P2WSH, P2WPKH or P2TR, defaultNet is not required.
func NewAddress(address string, defaultNet ...*chaincfg.Params) Address {
	addr, err := SafeNewAddress(address, defaultNet...)
	if err != nil {
		logger.Panic("can't create parse address", slogx.Error(err), slogx.String("package", "btcutils"))
	}
	return addr
}

// SafeNewAddress creates a new address from the given address string.
// It returns an error if the address is invalid.
//
// defaultNet is required if your address is P2SH or P2PKH (legacy or nested segwit)
// If your address is P2WSH, P2WPKH or P2TR, defaultNet is not required.
func SafeNewAddress(address string, defaultNet ...*chaincfg.Params) (Address, error) {
	net := utils.DefaultOptional(defaultNet, &chaincfg.MainNetParams)

	decoded, addrType, err := parseAddress(address, net)
	if err != nil {
		return Address{}, errors.Wrap(err, "can't parse address")
	}

	scriptPubkey, err := txscript.PayToAddrScript(decoded)
	if err != nil {
		return Address{}, errors.Wrap(err, "can't get script pubkey")
	}

	fixedPkScript := [MaxSupportedPkScriptSize]byte{}
	copy(fixedPkScript[:], scriptPubkey)
	return Address{
		decoded:          decoded,
		net:              net,
		encoded:          decoded.EncodeAddress(),
		encodedType:      addrType,
		scriptPubKey:     fixedPkScript,
		scriptPubKeySize: len(scriptPubkey),
	}, nil
}

// String returns the address string.
func (a Address) String() string {
	return a.encoded
}

// Type returns the address type.
func (a Address) Type() AddressType {
	return a.encodedType
}

// Decoded returns the btcutil.Address
func (a Address) Decoded() btcutil.Address {
	return a.decoded
}

// IsForNet returns whether or not the address is associated with the passed bitcoin network.
func (a Address) IsForNet(net *chaincfg.Params) bool {
	return a.decoded.IsForNet(net)
}

// ScriptAddress returns the raw bytes of the address to be used when inserting the address into a txout's script.
func (a Address) ScriptAddress() []byte {
	return a.decoded.ScriptAddress()
}

// Net returns the address network params.
func (a Address) Net() *chaincfg.Params {
	return a.net
}

// NetworkName
func (a Address) NetworkName() string {
	return a.net.Name
}

// ScriptPubKey or pubkey script
func (a Address) ScriptPubKey() []byte {
	return a.scriptPubKey[:a.scriptPubKeySize]
}

// Equal return true if addresses are equal
func (a Address) Equal(b Address) bool {
	return a.encoded == b.encoded
}

// DustLimit returns the output dust limit (lowest possible satoshis in a UTXO) for the address type.
func (a Address) DustLimit() uint64 {
	switch a.encodedType {
	case AddressP2TR:
		return 330
	case AddressP2WPKH:
		return 294
	default:
		return 546
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.encoded), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (a *Address) UnmarshalText(input []byte) error {
	address := string(input)
	addr, err := SafeNewAddress(address)
	if err == nil {
		*a = addr
		return nil
	}
	return errors.Wrapf(errs.InvalidArgument, "invalid address `%s`", address)
}

// MarshalJSON implements the json.Marshaler interface.
func (a Address) MarshalJSON() ([]byte, error) {
	t, err := a.MarshalText()
	if err != nil {
		return nil, &json.MarshalerError{Type: reflect.TypeOf(a), Err: err}
	}
	b := make([]byte, len(t)+2)
	b[0], b[len(b)-1] = '"', '"' // add quotes
	copy(b[1:], t)
	return b, nil
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	if !(len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"') {
		return &json.UnmarshalTypeError{Value: "non-string", Type: reflect.TypeOf(Address{})}
	}
	if err := a.UnmarshalText(input[1 : len(input)-1]); err != nil {
		return err
	}
	return nil
}

func parseAddress(address string, params *chaincfg.Params) (btcutil.Address, AddressType, error) {
	decoded, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "can't decode address `%s` for network `%s`", address, params.Name)
	}

	switch decoded.(type) {
	case *btcutil.AddressWitnessPubKeyHash:
		return decoded, AddressP2WPKH, nil
	case *btcutil.AddressTaproot:
		return decoded, AddressP2TR, nil
	case *btcutil.AddressScriptHash:
		return decoded, AddressP2SH, nil
	case *btcutil.AddressPubKeyHash:
		return decoded, AddressP2PKH, nil
	case *btcutil.AddressWitnessScriptHash:
		return decoded, AddressP2WSH, nil
	default:
		return nil, 0, errors.Wrap(errs.Unsupported, "unsupported address type")
	}
}
