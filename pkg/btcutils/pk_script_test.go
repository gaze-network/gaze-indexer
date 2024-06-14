package btcutils_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/stretchr/testify/assert"
)

func TestNewPkScript(t *testing.T) {
	anyError := errors.New("any error")

	type Spec struct {
		Address          string
		DefaultNet       *chaincfg.Params
		ExpectedError    error
		ExpectedPkScript string // hex encoded
	}

	specs := []Spec{
		{
			Address:          "some_invalid_address",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    anyError,
			ExpectedPkScript: "",
		},
		{
			// P2WPKH
			Address:          "bc1qdx72th7e3z8zc5wdrdxweswfcne974pjneyjln",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "001469bca5dfd9888e2c51cd1b4cecc1c9c4f25f5432",
		},
		{
			// P2WPKH
			Address:          "bc1q7cj6gz6t3d28qg7kxhrc7h5t3h0re34fqqalga",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "0014f625a40b4b8b547023d635c78f5e8b8dde3cc6a9",
		},
		{
			// P2TR
			Address:          "bc1pfd0zw2jwlpn4xckpr3dxpt7x0gw6wetuftxvrc4dt2qgn9azjuus65fug6",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "51204b5e272a4ef8675362c11c5a60afc67a1da7657c4accc1e2ad5a808997a29739",
		},
		{
			// P2TR
			Address:          "bc1pxpumml545tqum5afarzlmnnez2npd35nvf0j0vnrp88nemqsn54qle05sm",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "51203079bdfe95a2c1cdd3a9e8c5fdce7912a616c693625f27b26309cf3cec109d2a",
		},
		{
			// P2SH
			Address:          "3Ccte7SJz71tcssLPZy3TdWz5DTPeNRbPw",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "a91477e1a3d54f545d83869ae3a6b28b071422801d7b87",
		},
		{
			// P2PKH
			Address:          "1KrRZSShVkdc8J71CtY4wdw46Rx3BRLKyH",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "76a914cecb25b53809991c7beef2d27bc2be49e78c684388ac",
		},
		{
			// P2WSH
			Address:          "bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak",
			DefaultNet:       &chaincfg.MainNetParams,
			ExpectedError:    nil,
			ExpectedPkScript: "0020cdbf909e935c855d3e8d1b61aeb9c5e3c03ae8021b286839b1a72f2e48fdba70",
		},
	}

	for _, spec := range specs {
		t.Run(fmt.Sprintf("address:%s", spec.Address), func(t *testing.T) {
			// Validate Expected PkScript
			if spec.ExpectedError == nil {
				{
					expectedPkScriptRaw, err := hex.DecodeString(spec.ExpectedPkScript)
					if err != nil {
						t.Fatalf("can't decode expected pkscript %s, Reason: %s", spec.ExpectedPkScript, err)
					}
					expectedPkScript, err := txscript.ParsePkScript(expectedPkScriptRaw)
					if err != nil {
						t.Fatalf("invalid expected pkscript %s, Reason: %s", spec.ExpectedPkScript, err)
					}

					expectedAddress, err := expectedPkScript.Address(spec.DefaultNet)
					if err != nil {
						t.Fatalf("can't get address from expected pkscript %s, Reason: %s", spec.ExpectedPkScript, err)
					}
					assert.Equal(t, spec.Address, expectedAddress.EncodeAddress())
				}
				{
					address, err := btcutil.DecodeAddress(spec.Address, spec.DefaultNet)
					if err != nil {
						t.Fatalf("can't decode address %s(%s),Reason: %s", spec.Address, spec.DefaultNet.Name, err)
					}

					pkScript, err := txscript.PayToAddrScript(address)
					if err != nil {
						t.Fatalf("can't get pkscript from address %s(%s),Reason: %s", spec.Address, spec.DefaultNet.Name, err)
					}

					pkScriptStr := hex.EncodeToString(pkScript)
					assert.Equal(t, spec.ExpectedPkScript, pkScriptStr)
				}
			}

			pkScript, err := btcutils.NewPkScript(spec.Address, spec.DefaultNet)
			if spec.ExpectedError == anyError {
				assert.Error(t, err)
			} else if spec.ExpectedError != nil {
				assert.ErrorIs(t, err, spec.ExpectedError)
			} else {
				address, err := btcutils.SafeNewAddress(spec.Address, spec.DefaultNet)
				if err != nil {
					t.Fatalf("can't create address %s(%s),Reason: %s", spec.Address, spec.DefaultNet.Name, err)
				}

				// ScriptPubKey from address and from NewPkScript should be the same
				assert.Equal(t, address.ScriptPubKey(), pkScript)

				// Expected PkScript and New PkScript should be the same
				pkScriptStr := hex.EncodeToString(pkScript)
				assert.Equal(t, spec.ExpectedPkScript, pkScriptStr)

				// Can convert PkScript back to same address
				acualPkScript, err := txscript.ParsePkScript(address.ScriptPubKey())
				if !assert.NoError(t, err) {
					t.Fail()
				}
				assert.Equal(t, address.Decoded().String(), utils.Must(acualPkScript.Address(spec.DefaultNet)).String())
			}
		})
	}
}

func TestGetAddressTypeFromPkScript(t *testing.T) {
	type Spec struct {
		PubkeyScript string

		ExpectedError       error
		ExpectedAddressType btcutils.AddressType
	}

	specs := []Spec{
		{
			PubkeyScript: "0014602181cc89f7c9f54cb6d7607a3445e3e022895d",

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WPKH,
		},
		{
			PubkeyScript: "5120ef8d59038dd51093fbfff794f658a07a3697b94d9e6d24e45b28abd88f10e33d",

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2TR,
		},
		{
			PubkeyScript: "a91416eef7e84fb9821db1341b6ccef1c4a4e5ec21e487",

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2SH,
		},
		{
			PubkeyScript: "76a914cecb25b53809991c7beef2d27bc2be49e78c684388ac",

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2PKH,
		},
		{
			PubkeyScript: "0020cdbf909e935c855d3e8d1b61aeb9c5e3c03ae8021b286839b1a72f2e48fdba70",

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WSH,
		},
		{
			PubkeyScript: "0020cdbf909e935c855d3e8d1b61aeb9c5e3c03ae8021b286839b1a72f2e48fdba70",

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WSH,
		},
		{
			PubkeyScript: "6a5d0614c0a2331441",

			ExpectedError:       nil,
			ExpectedAddressType: txscript.NonStandardTy,
		},
	}

	for _, spec := range specs {
		t.Run(fmt.Sprintf("PkScript:%s", spec.PubkeyScript), func(t *testing.T) {
			pkScript, err := hex.DecodeString(spec.PubkeyScript)
			if err != nil {
				t.Fail()
			}
			actualAddressType, actualError := btcutils.GetAddressTypeFromPkScript(pkScript)
			if spec.ExpectedError != nil {
				assert.ErrorIs(t, actualError, spec.ExpectedError)
			} else {
				assert.Equal(t, spec.ExpectedAddressType, actualAddressType)
			}
		})
	}
}
