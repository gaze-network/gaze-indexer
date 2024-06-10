package brc20

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/shopspring/decimal"
)

type rawPayload struct {
	P    string // required
	Op   string `json:"op"`   // required
	Tick string `json:"tick"` // required

	// for deploy operations
	Max      string  `json:"max"` // required
	Lim      *string `json:"lim"`
	Dec      *string `json:"dec"`
	SelfMint *string `json:"self_mint"`

	// for mint/transfer operations
	Amt string `json:"amt"` // required
}

type Payload struct {
	Transfer     *entity.InscriptionTransfer
	P            string
	Op           Operation
	Tick         string // lower-cased tick
	OriginalTick string // original tick before lower-cased

	// for deploy operations
	Max      decimal.Decimal
	Lim      decimal.Decimal
	Dec      uint16
	SelfMint bool

	// for mint/transfer operations
	Amt decimal.Decimal
}

var (
	ErrInvalidProtocol   = errors.New("invalid protocol: must be 'brc20'")
	ErrInvalidOperation  = errors.New("invalid operation for brc20: must be one of 'deploy', 'mint', or 'transfer'")
	ErrInvalidTickLength = errors.New("invalid tick length: must be 4 or 5 bytes")
	ErrEmptyTick         = errors.New("empty tick")
	ErrEmptyMax          = errors.New("empty max")
	ErrInvalidMax        = errors.New("invalid max")
	ErrInvalidDec        = errors.New("invalid dec")
	ErrInvalidSelfMint   = errors.New("invalid self_mint")
	ErrInvalidAmt        = errors.New("invalid amt")
	ErrNumberOverflow    = errors.New("number overflow: max value is (2^64-1)")
)

func ParsePayload(transfer *entity.InscriptionTransfer) (*Payload, error) {
	var p rawPayload
	err := json.Unmarshal(transfer.Content, &p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload as json")
	}

	if p.P != "brc-20" {
		return nil, errors.WithStack(ErrInvalidProtocol)
	}
	if !Operation(p.Op).IsValid() {
		return nil, errors.WithStack(ErrInvalidOperation)
	}
	if p.Tick == "" {
		return nil, errors.WithStack(ErrEmptyTick)
	}
	if len(p.Tick) != 4 && len(p.Tick) != 5 {
		return nil, errors.WithStack(ErrInvalidTickLength)
	}

	parsed := Payload{
		Transfer:     transfer,
		P:            p.P,
		Op:           Operation(p.Op),
		Tick:         strings.ToLower(p.Tick),
		OriginalTick: p.Tick,
	}

	switch parsed.Op {
	case OperationDeploy:
		if p.Max == "" {
			return nil, errors.WithStack(ErrEmptyMax)
		}
		var rawDec string
		if p.Dec != nil {
			rawDec = *p.Dec
		}
		if rawDec == "" {
			rawDec = "18"
		}
		dec, ok := strconv.ParseUint(rawDec, 10, 16)
		if ok != nil {
			return nil, errors.Wrap(ok, "failed to parse dec")
		}
		if dec > 18 {
			return nil, errors.WithStack(ErrInvalidDec)
		}
		parsed.Dec = uint16(dec)

		max, err := parseNumericString(p.Max, dec)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse max")
		}
		parsed.Max = max

		limit := max
		if p.Lim != nil {
			limit, err = parseNumericString(*p.Lim, dec)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse lim")
			}
		}
		parsed.Lim = limit

		// 5-bytes ticks are self-mint only
		if len(parsed.OriginalTick) == 5 {
			if p.SelfMint == nil || *p.SelfMint != "true" {
				return nil, errors.WithStack(ErrInvalidSelfMint)
			}
			// infinite mints if tick is self-mint, and max is set to 0
			if parsed.Max.IsZero() {
				parsed.Max = maxNumber
				if parsed.Lim.IsZero() {
					parsed.Lim = maxNumber
				}
			}
		}
		if parsed.Max.IsZero() {
			return nil, errors.WithStack(ErrInvalidMax)
		}
	case OperationMint, OperationTransfer:
		if p.Amt == "" {
			return nil, errors.WithStack(ErrInvalidAmt)
		}
		// NOTE: check tick decimals after parsing payload
		amt, err := parseNumericString(p.Amt, 18)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse amt")
		}
		parsed.Amt = amt
	default:
		return nil, errors.WithStack(ErrInvalidOperation)
	}
	return &parsed, nil
}

// max number for all numeric fields (except dec) is (2^64-1)
var (
	maxNumber = decimal.NewFromBigInt(new(big.Int).SetUint64(math.MaxUint64), 0)
)

func parseNumericString(s string, maxDec uint64) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, errors.Wrap(err, "failed to parse decimal number")
	}
	if -d.Exponent() > int32(maxDec) {
		return decimal.Decimal{}, errors.Errorf("cannot parse decimal number: too many decimal points: expected %d got %d", maxDec, d.Exponent())
	}
	if d.GreaterThan(maxNumber) {
		return decimal.Decimal{}, errors.WithStack(ErrNumberOverflow)
	}
	return d, nil
}
