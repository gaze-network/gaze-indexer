package entity

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
	"github.com/stretchr/testify/assert"
)

func TestOutPointBalanceJSON(t *testing.T) {
	ob := OutPointBalance{
		PkScript:   utils.Must(hex.DecodeString("51203daaca9b82a51aca960c1491588246029d7e0fc49e0abdbcc8fd17574be5c74b")),
		Id:         runes.RuneId{BlockHeight: 1, TxIndex: 2},
		Amount:     uint128.From64(100),
		Index:      1,
		TxHash:     *utils.Must(chainhash.NewHashFromStr("3ea1b497b25993adf3f2c8dae1470721316a45c82600798c14d0425039c410ad")),
		TxOutIndex: 2,
	}
	bytes, err := json.Marshal(ob)
	assert.NoError(t, err)
	t.Log(string(bytes))

	var parsedOB OutPointBalance
	err = json.Unmarshal(bytes, &parsedOB)
	assert.NoError(t, err)
	assert.Equal(t, ob, parsedOB)
}
