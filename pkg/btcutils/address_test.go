package btcutils_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/stretchr/testify/assert"
)

func TestGetAddressType(t *testing.T) {
	type Spec struct {
		Address    string
		DefaultNet *chaincfg.Params

		ExpectedError       error
		ExpectedAddressType btcutils.AddressType
	}

	specs := []Spec{
		{
			Address:    "bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczh",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WPKH,
		},
		{
			Address:    "tb1qfpgdxtpl7kz5qdus2pmexyjaza99c28qd6ltey",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WPKH,
		},
		{
			Address:    "bc1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qvz5d38",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2TR,
		},
		{
			Address:    "tb1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qm2zztg",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2TR,
		},
		{
			Address:    "3Ccte7SJz71tcssLPZy3TdWz5DTPeNRbPw",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2SH,
		},
		{
			Address:    "1KrRZSShVkdc8J71CtY4wdw46Rx3BRLKyH",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2PKH,
		},
		{
			Address:    "bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WSH,
		},
		{
			Address:    "migbBPcDajPfffrhoLpYFTQNXQFbWbhpz3",
			DefaultNet: &chaincfg.TestNet3Params,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2PKH,
		},
		{
			Address:    "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2WSH,
		},
		{
			Address:    "2NCxMvHPTduZcCuUeAiWUpuwHga7Y66y9XJ",
			DefaultNet: &chaincfg.TestNet3Params,

			ExpectedError:       nil,
			ExpectedAddressType: btcutils.AddressP2SH,
		},
	}

	for _, spec := range specs {
		t.Run(fmt.Sprintf("address:%s", spec.Address), func(t *testing.T) {
			actualAddressType, actualError := btcutils.GetAddressType(spec.Address, spec.DefaultNet)
			if spec.ExpectedError != nil {
				assert.ErrorIs(t, actualError, spec.ExpectedError)
			} else {
				assert.Equal(t, spec.ExpectedAddressType, actualAddressType)
			}
		})
	}
}

func TestNewAddress(t *testing.T) {
	type Spec struct {
		Address    string
		DefaultNet *chaincfg.Params

		ExpectedAddressType btcutils.AddressType
	}

	specs := []Spec{
		{
			Address: "bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczh",
			// DefaultNet: &chaincfg.MainNetParams, // Optional

			ExpectedAddressType: btcutils.AddressP2WPKH,
		},
		{
			Address: "tb1qfpgdxtpl7kz5qdus2pmexyjaza99c28qd6ltey",
			// DefaultNet: &chaincfg.MainNetParams, // Optional

			ExpectedAddressType: btcutils.AddressP2WPKH,
		},
		{
			Address: "bc1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qvz5d38",
			// DefaultNet: &chaincfg.MainNetParams, // Optional

			ExpectedAddressType: btcutils.AddressP2TR,
		},
		{
			Address: "tb1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qm2zztg",
			// DefaultNet: &chaincfg.MainNetParams, // Optional

			ExpectedAddressType: btcutils.AddressP2TR,
		},
		{
			Address: "bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak",
			// DefaultNet: &chaincfg.MainNetParams, // Optional

			ExpectedAddressType: btcutils.AddressP2WSH,
		},
		{
			Address: "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7",
			// DefaultNet: &chaincfg.MainNetParams, // Optional

			ExpectedAddressType: btcutils.AddressP2WSH,
		},
		{
			Address:    "3Ccte7SJz71tcssLPZy3TdWz5DTPeNRbPw",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedAddressType: btcutils.AddressP2SH,
		},
		{
			Address:    "2NCxMvHPTduZcCuUeAiWUpuwHga7Y66y9XJ",
			DefaultNet: &chaincfg.TestNet3Params,

			ExpectedAddressType: btcutils.AddressP2SH,
		},
		{
			Address:    "1KrRZSShVkdc8J71CtY4wdw46Rx3BRLKyH",
			DefaultNet: &chaincfg.MainNetParams,

			ExpectedAddressType: btcutils.AddressP2PKH,
		},
		{
			Address:    "migbBPcDajPfffrhoLpYFTQNXQFbWbhpz3",
			DefaultNet: &chaincfg.TestNet3Params,

			ExpectedAddressType: btcutils.AddressP2PKH,
		},
	}

	for _, spec := range specs {
		t.Run(fmt.Sprintf("address:%s,type:%s", spec.Address, spec.ExpectedAddressType), func(t *testing.T) {
			addr := btcutils.NewAddress(spec.Address, spec.DefaultNet)

			assert.Equal(t, spec.ExpectedAddressType, addr.Type())
			assert.Equal(t, spec.Address, addr.String())
		})
	}
}

func TestIsAddress(t *testing.T) {
	type Spec struct {
		Address  string
		Expected bool
	}

	specs := []Spec{
		{
			Address: "bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczh",

			Expected: true,
		},
		{
			Address: "tb1qfpgdxtpl7kz5qdus2pmexyjaza99c28qd6ltey",

			Expected: true,
		},
		{
			Address: "bc1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qvz5d38",

			Expected: true,
		},
		{
			Address: "tb1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qm2zztg",

			Expected: true,
		},
		{
			Address: "bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak",

			Expected: true,
		},
		{
			Address: "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7",

			Expected: true,
		},
		{
			Address: "3Ccte7SJz71tcssLPZy3TdWz5DTPeNRbPw",

			Expected: true,
		},
		{
			Address: "2NCxMvHPTduZcCuUeAiWUpuwHga7Y66y9XJ",

			Expected: true,
		},
		{
			Address: "1KrRZSShVkdc8J71CtY4wdw46Rx3BRLKyH",

			Expected: true,
		},
		{
			Address: "migbBPcDajPfffrhoLpYFTQNXQFbWbhpz3",

			Expected: true,
		},
		{
			Address: "",

			Expected: false,
		},
		{
			Address: "migbBPcDajPfffrhoLpYFTQNXQFbWbhpz2",

			Expected: false,
		},
		{
			Address: "bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczz",

			Expected: false,
		},
	}

	for _, spec := range specs {
		t.Run(fmt.Sprintf("address:%s", spec.Address), func(t *testing.T) {
			ok := btcutils.IsAddress(spec.Address)
			assert.Equal(t, spec.Expected, ok)
		})
	}
}

func TestAddressEncoding(t *testing.T) {
	rawAddress := "bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczh"
	address := btcutils.NewAddress(rawAddress, &chaincfg.MainNetParams)

	type Spec struct {
		Data     interface{}
		Expected string
	}

	specs := []Spec{
		{
			Data:     address,
			Expected: fmt.Sprintf(`"%s"`, rawAddress),
		},
		{
			Data: map[string]interface{}{
				"address": rawAddress,
			},
			Expected: fmt.Sprintf(`{"address":"%s"}`, rawAddress),
		},
	}

	for i, spec := range specs {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			actual, err := json.Marshal(spec.Data)
			assert.NoError(t, err)
			assert.Equal(t, spec.Expected, string(actual))
		})
	}
}

func TestAddressDecoding(t *testing.T) {
	rawAddress := "bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczh"
	address := btcutils.NewAddress(rawAddress, &chaincfg.MainNetParams)

	// Case #1: address is a string
	t.Run("from_string", func(t *testing.T) {
		input := fmt.Sprintf(`"%s"`, rawAddress)
		expected := address
		actual := btcutils.Address{}

		err := json.Unmarshal([]byte(input), &actual)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.Equal(t, expected, actual)
	})

	// Case #2: address is a field of a struct
	t.Run("from_field_string", func(t *testing.T) {
		type Data struct {
			Address btcutils.Address `json:"address"`
		}
		input := fmt.Sprintf(`{"address":"%s"}`, rawAddress)
		expected := Data{Address: address}
		actual := Data{}
		err := json.Unmarshal([]byte(input), &actual)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.Equal(t, expected, actual)
	})

	// Case #3: address is an element of an array
	t.Run("from_array", func(t *testing.T) {
		input := fmt.Sprintf(`["%s"]`, rawAddress)
		expected := []btcutils.Address{address}
		actual := []btcutils.Address{}
		err := json.Unmarshal([]byte(input), &actual)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.Equal(t, expected, actual)
	})

	// Case #4: not supported address type
	t.Run("from_string/not_address", func(t *testing.T) {
		input := fmt.Sprintf(`"%s"`, "THIS_IS_NOT_SUPPORTED_ADDRESS")
		actual := btcutils.Address{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.Error(t, err)
	})

	// Case #5: invalid field type
	t.Run("from_number", func(t *testing.T) {
		type Data struct {
			Address btcutils.Address `json:"address"`
		}
		input := fmt.Sprintf(`{"address":%d}`, 123)
		actual := Data{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.Error(t, err)
	})
}
