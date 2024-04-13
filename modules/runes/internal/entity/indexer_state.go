package entity

import "time"

type IndexerState struct {
	CreatedAt        time.Time
	DBVersion        int32
	EventHashVersion int32
}
