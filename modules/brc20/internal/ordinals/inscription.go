package ordinals

import "time"

type Inscription struct {
	Content         []byte
	ContentEncoding string
	ContentType     string
	Delegate        *InscriptionId
	Metadata        []byte
	Metaprotocol    string
	Parent          *InscriptionId
	Pointer         *uint64
}

type InscriptionEntry struct {
	Id              InscriptionId
	Number          int64
	SequenceNumber  uint64
	Cursed          bool
	Vindicated      bool
	CreatedAt       time.Time
	CreatedAtHeight uint64
	TransferCount   uint32
	Inscription     Inscription
}
