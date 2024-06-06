package ordinals

import "time"

type Inscription struct {
	Content         []byte
	ContentEncoding string
	ContentType     string
	Delegate        *InscriptionId
	Metadata        []byte
	Metaprotocol    string
	Parent          *InscriptionId // in 0.14, inscription has only one parent
	Pointer         *uint64
}

// TODO: refactor ordinals.InscriptionEntry to entity.InscriptionEntry
type InscriptionEntry struct {
	Id              InscriptionId
	Number          int64
	SequenceNumber  uint64
	Cursed          bool
	CursedForBRC20  bool
	CreatedAt       time.Time
	CreatedAtHeight uint64
	Inscription     Inscription
	TransferCount   uint32
}
