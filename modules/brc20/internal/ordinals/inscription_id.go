package ordinals

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
)

type InscriptionId struct {
	TxHash chainhash.Hash
	Index  uint32
}

func (i InscriptionId) String() string {
	return fmt.Sprintf("%si%d", i.TxHash.String(), i.Index)
}

func NewInscriptionId(txHash chainhash.Hash, index uint32) InscriptionId {
	return InscriptionId{
		TxHash: txHash,
		Index:  index,
	}
}

var ErrInscriptionIdInvalidSeparator = fmt.Errorf("invalid inscription id: must contain exactly one separator")

func NewInscriptionIdFromString(s string) (InscriptionId, error) {
	parts := strings.SplitN(s, "i", 2)
	if len(parts) != 2 {
		return InscriptionId{}, errors.WithStack(ErrInscriptionIdInvalidSeparator)
	}
	txHash, err := chainhash.NewHashFromStr(parts[0])
	if err != nil {
		return InscriptionId{}, errors.Wrap(err, "invalid inscription id: cannot parse txHash")
	}
	index, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return InscriptionId{}, errors.Wrap(err, "invalid inscription id: cannot parse index")
	}
	return InscriptionId{
		TxHash: *txHash,
		Index:  uint32(index),
	}, nil
}

// MarshalJSON implements json.Marshaler
func (r InscriptionId) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (r *InscriptionId) UnmarshalJSON(data []byte) error {
	// data must be quoted
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("must be string")
	}
	data = data[1 : len(data)-1]
	parsed, err := NewInscriptionIdFromString(string(data))
	if err != nil {
		return errors.WithStack(err)
	}
	*r = parsed
	return nil
}
