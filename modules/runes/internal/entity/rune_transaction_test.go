package entity

import (
	"encoding/json"
	"testing"

	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

func TestMarshal(t *testing.T) {
	burns := make(map[string]uint128.Uint128)
	burns["1:2"] = uint128.From64(100)
	burns["3:4"] = uint128.From64(200)
	bytes, err := json.Marshal(burns)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}

func TestUnmarshal(t *testing.T) {
	burns := make(map[runes.RuneId]uint128.Uint128)
	bytes := []byte(`{"1:2":"100","3:4":"200"}`)
	err := json.Unmarshal(bytes, &burns)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(burns)
}
