package btcutils_test

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
)

/*
NOTE:

# Compare this benchmark to go-ethereum/common.Address utils
- go-ethereum/common.HexToAddress speed: 45 ns/op, 48 B/op, 1 allocs/op
- go-ethereum/common.IsHexAddress speed: 25 ns/op, 0 B/op, 0 allocs/op

It's slower than go-ethereum/common.Address utils because ethereum wallet address is Hex string 20 bytes,
but Bitcoin has many types of address and each type has complex algorithm to solve (can't solve and validate address type directly from address string)

20/Jan/2024 @Planxnx Macbook Air M1 16GB
BenchmarkIsAddress/specific-network/mainnet/P2WPKH-8         	 1776146	       625.6 ns/op	     120 B/op	       3 allocs/op
BenchmarkIsAddress/specific-network/testnet3/P2WPKH-8        	 1917876	       623.2 ns/op	     120 B/op	       3 allocs/op
BenchmarkIsAddress/specific-network/mainnet/P2TR-8           	 1330348	       915.4 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/specific-network/testnet3/P2TR-8          	 1235806	       931.1 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/specific-network/mainnet/P2WSH-8          	 1261730	       960.9 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/specific-network/testnet3/P2WSH-8         	 1307851	       916.1 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/specific-network/mainnet/P2SH-8           	 3081762	       402.0 ns/op	     192 B/op	       8 allocs/op
BenchmarkIsAddress/specific-network/testnet3/P2SH-8          	 3245838	       344.9 ns/op	     176 B/op	       7 allocs/op
BenchmarkIsAddress/specific-network/mainnet/P2PKH-8          	 2904252	       410.4 ns/op	     184 B/op	       8 allocs/op
BenchmarkIsAddress/specific-network/testnet3/P2PKH-8         	 3522332	       342.8 ns/op	     176 B/op	       7 allocs/op
BenchmarkIsAddress/automate-network/mainnet/P2WPKH-8         	 1882059	       637.6 ns/op	     120 B/op	       3 allocs/op
BenchmarkIsAddress/automate-network/testnet3/P2WPKH-8        	 1626151	       664.8 ns/op	     120 B/op	       3 allocs/op
BenchmarkIsAddress/automate-network/mainnet/P2TR-8           	 1250253	       952.1 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/automate-network/testnet3/P2TR-8          	 1257901	       993.7 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/automate-network/mainnet/P2WSH-8          	 1000000	      1005 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/automate-network/testnet3/P2WSH-8         	 1209108	       971.2 ns/op	     160 B/op	       3 allocs/op
BenchmarkIsAddress/automate-network/mainnet/P2SH-8           	 1869075	       625.0 ns/op	     268 B/op	       9 allocs/op
BenchmarkIsAddress/automate-network/testnet3/P2SH-8          	  779496	      1609 ns/op	     694 B/op	      17 allocs/op
BenchmarkIsAddress/automate-network/mainnet/P2PKH-8          	 1924058	       650.6 ns/op	     259 B/op	       9 allocs/op
BenchmarkIsAddress/automate-network/testnet3/P2PKH-8         	  721510	      1690 ns/op	     694 B/op	      17 allocs/op
*/
func BenchmarkIsAddress(b *testing.B) {
	cases := []btcutils.Address{
		/* P2WPKH */ btcutils.NewAddress("bc1qfpgdxtpl7kz5qdus2pmexyjaza99c28q8uyczh", &chaincfg.MainNetParams),
		/* P2WPKH */ btcutils.NewAddress("tb1qfpgdxtpl7kz5qdus2pmexyjaza99c28qd6ltey", &chaincfg.TestNet3Params),
		/* P2TR */ btcutils.NewAddress("bc1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qvz5d38", &chaincfg.MainNetParams),
		/* P2TR */ btcutils.NewAddress("tb1p7h87kqsmpzatddzhdhuy9gmxdpvn5kvar6hhqlgau8d2ffa0pa3qm2zztg", &chaincfg.TestNet3Params),
		/* P2WSH */ btcutils.NewAddress("bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak", &chaincfg.MainNetParams),
		/* P2WSH */ btcutils.NewAddress("tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7", &chaincfg.TestNet3Params),
		/* P2SH */ btcutils.NewAddress("3Ccte7SJz71tcssLPZy3TdWz5DTPeNRbPw", &chaincfg.MainNetParams),
		/* P2SH */ btcutils.NewAddress("2NCxMvHPTduZcCuUeAiWUpuwHga7Y66y9XJ", &chaincfg.TestNet3Params),
		/* P2PKH */ btcutils.NewAddress("1KrRZSShVkdc8J71CtY4wdw46Rx3BRLKyH", &chaincfg.MainNetParams),
		/* P2PKH */ btcutils.NewAddress("migbBPcDajPfffrhoLpYFTQNXQFbWbhpz3", &chaincfg.TestNet3Params),
	}

	b.Run("specific-network", func(b *testing.B) {
		for _, c := range cases {
			b.Run(c.NetworkName()+"/"+c.Type().String(), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = btcutils.IsAddress(c.String(), c.Net())
				}
			})
		}
	})

	b.Run("automate-network", func(b *testing.B) {
		for _, c := range cases {
			b.Run(c.NetworkName()+"/"+c.Type().String(), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					ok := btcutils.IsAddress(c.String())
					if !ok {
						b.Error("IsAddress returned false")
					}
				}
			})
		}
	})
}
