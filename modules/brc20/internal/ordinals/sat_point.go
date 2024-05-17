package ordinals

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
)

type SatPoint struct {
	OutPoint wire.OutPoint
	Offset   uint64
}

func (s SatPoint) String() string {
	return fmt.Sprintf("%s:%d", s.OutPoint.String(), s.Offset)
}

var ErrSatPointInvalidSeparator = fmt.Errorf("invalid sat point: must contain exactly two separators")

func NewSatPointFromString(s string) (SatPoint, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return SatPoint{}, errors.WithStack(ErrSatPointInvalidSeparator)
	}
	txHash, err := chainhash.NewHashFromStr(parts[0])
	if err != nil {
		return SatPoint{}, errors.Wrap(err, "invalid inscription id: cannot parse txHash")
	}
	index, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return SatPoint{}, errors.Wrap(err, "invalid inscription id: cannot parse index")
	}
	offset, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return SatPoint{}, errors.Wrap(err, "invalid sat point: cannot parse offset")
	}
	return SatPoint{
		OutPoint: wire.OutPoint{
			Hash:  *txHash,
			Index: uint32(index),
		},
		Offset: offset,
	}, nil
}

// MarshalJSON implements json.Marshaler
func (r SatPoint) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (r *SatPoint) UnmarshalJSON(data []byte) error {
	// data must be quoted
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("must be string")
	}
	data = data[1 : len(data)-1]
	parsed, err := NewSatPointFromString(string(data))
	if err != nil {
		return errors.WithStack(err)
	}
	*r = parsed
	return nil
}
