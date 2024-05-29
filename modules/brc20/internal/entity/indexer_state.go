package entity

import (
	"time"

	"github.com/gaze-network/indexer-network/common"
)

type IndexerState struct {
	CreatedAt        time.Time
	ClientVersion    string
	DBVersion        int32
	EventHashVersion int32
	Network          common.Network
}
