package reportingclientv2

import (
	"context"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/stretchr/testify/assert"
)

// Key for test only. DO NOT USE IN PRODUCTION
const (
	privateKeyStrOfficial = "ce9c2fd75623e82a83ed743518ec7749f6f355f7301dd432400b087717fed2f2"
	mainnetKeyOfficial    = "L2hqSEYLjRQHHSiSgfNSFwaZ1dYpwn4PUTt2nQT8AefzYGQCwsGY"
	pubKeyStrOfficial     = "0251e2dfcdeea17cc9726e4be0855cd0bae19e64f3e247b10760cd76851e7df47e"
)

const (
	privateKeyStr1 = "a3a7d2c40c8bb7a4b4afa9a6b3ed613da1de233913ed07a017e6dd44ef542d80"
	mainnetKey1    = "L49LKamtrPZxty5TG7jaFPHMRZbrvAr4Dvn5BHGdvmvbcTDNAbZj"
	pubKeyStr1     = "02596e11b5a2104533d2732a3df35eadaeb61b189b9069715106e72e27c1de7775"
)

const (
	privateKeyStr2 = "a3a7d2c40c8bb7a4b4afa9a6b3ed613da1de233913ed07a017e6dd44ef542d81"
	// mainnetKey1    = "L49LKamtrPZxty5TG7jaFPHMRZbrvAr4Dvn5BHGdvmvbcTDNAbZj"
	// pubKeyStr1     = "02596e11b5a2104533d2732a3df35eadaeb61b189b9069715106e72e27c1de7775"
)

func TestReport1(t *testing.T) {
	block := 18
	hash, err := chainhash.NewHashFromStr(fmt.Sprintf("%d", block))
	assert.NoError(t, err)

	configOfficial := Config{
		ReportCenter: ReportCenter{
			BaseURL:   "https://indexer-dev.api.gaze.network",
			PublicKey: defaultPublicKey,
		},
		NodeInfo: NodeInfo{
			Name:   "Official Node",
			APIURL: "http://localhost:2000",
		},
	}
	config1 := Config{
		ReportCenter: ReportCenter{
			BaseURL:   "https://indexer-dev.api.gaze.network",
			PublicKey: defaultPublicKey,
		},
		NodeInfo: NodeInfo{
			Name:   "Node 1",
			APIURL: "http://localhost:2000",
		},
	}
	config2 := Config{
		ReportCenter: ReportCenter{
			BaseURL:   "https://indexer-dev.api.gaze.network",
			PublicKey: defaultPublicKey,
		},
		NodeInfo: NodeInfo{
			Name:   "Node 2",
			APIURL: "http://localhost:2000",
		},
	}
	clientOfficial, err := New(configOfficial, privateKeyStrOfficial)
	assert.NoError(t, err)
	client1, err := New(config1, privateKeyStr1)
	assert.NoError(t, err)
	client2, err := New(config2, privateKeyStr2)
	assert.NoError(t, err)

	err = clientOfficial.SubmitNodeReport(context.Background(), "runes", "mainnet", "v0.0.1")
	assert.NoError(t, err)
	err = client1.SubmitNodeReport(context.Background(), "runes", "mainnet", "v0.0.1")
	assert.NoError(t, err)
	err = client2.SubmitNodeReport(context.Background(), "runes", "mainnet", "v0.0.1")
	assert.NoError(t, err)

	blockReport := SubmitBlockReportPayloadData{
		Type:                "runes",
		ClientVersion:       "v0.0.1",
		DBVersion:           1,
		EventHashVersion:    1,
		Network:             common.NetworkMainnet,
		BlockHeight:         uint64(block),
		BlockHash:           *hash,
		EventHash:           *hash,
		CumulativeEventHash: *hash,
	}

	err = clientOfficial.SubmitBlockReport(context.Background(), blockReport)
	err = client1.SubmitBlockReport(context.Background(), blockReport)
	err = client2.SubmitBlockReport(context.Background(), blockReport)
}
