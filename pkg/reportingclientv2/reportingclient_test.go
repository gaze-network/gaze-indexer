package reportingclientv2

import (
	"context"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/stretchr/testify/assert"
)

// Key for test only. DO NOT USE IN PRODUCTION
const (
	privateKeyStr = "ce9c2fd75623e82a83ed743518ec7749f6f355f7301dd432400b087717fed2f2"
	mainnetKey    = "L49LKamtrPZxty5TG7jaFPHMRZbrvAr4Dvn5BHGdvmvbcTDNAbZj"
	pubKeyStr     = "0251e2dfcdeea17cc9726e4be0855cd0bae19e64f3e247b10760cd76851e7df47e"
)

func TestReport(t *testing.T) {
	config := Config{
		ReportCenter: ReportCenter{
			BaseURL:   "http://localhost:8000",
			PublicKey: defaultPublicKey,
		},
		NodeInfo: NodeInfo{
			Name: "tests",
		},
	}
	client, err := New(config, privateKeyStr)
	assert.NoError(t, err)

	err = client.SubmitNodeReport(context.Background(), "runes", "mainnet", "v0.0.1")
	assert.NoError(t, err)

	err = client.SubmitBlockReport(context.Background(), SubmitBlockReportPayloadData{
		Type:                "runes",
		ClientVersion:       "v0.0.1",
		DBVersion:           1,
		EventHashVersion:    1,
		Network:             common.NetworkMainnet,
		BlockHeight:         2,
		BlockHash:           chainhash.Hash{},
		EventHash:           chainhash.Hash{},
		CumulativeEventHash: chainhash.Hash{},
	})
}
