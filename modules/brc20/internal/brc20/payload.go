package brc20

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/uint128"
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
	Max      uint128.Uint128
	Lim      uint128.Uint128
	Dec      uint16
	SelfMint bool

	// for mint/transfer operations
	Amt uint128.Uint128
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
	ErrNumberOverflow    = errors.New("number overflow: max value is (2^64-1) * 10^18")
)

func ParsePayload(transfer *entity.InscriptionTransfer) (*Payload, error) {
	var p rawPayload
	err := json.Unmarshal(transfer.Content, &p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload as json")
	}

	if p.P != "brc20" {
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
		dec, ok := strconv.ParseUint(rawDec, 10, 16)
		if ok != nil {
			return nil, errors.Wrap(ok, "failed to parse dec")
		}
		if dec > 18 {
			return nil, errors.WithStack(ErrInvalidDec)
		}
		parsed.Dec = uint16(dec)

		max, err := parseNumberExtendedTo18Decimal(p.Max, dec)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse max")
		}
		parsed.Max = max

		limit := max
		if p.Lim != nil {
			limit, err = parseNumberExtendedTo18Decimal(*p.Lim, dec)
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
				parsed.Max = maxIntegerValue
				if parsed.Lim.IsZero() {
					parsed.Lim = maxIntegerValue
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
		amt, err := parseNumberExtendedTo18Decimal(p.Amt, 18)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse amt")
		}
		parsed.Amt = amt
	default:
		return nil, errors.WithStack(ErrInvalidOperation)
	}
	return &parsed, nil
}

// max integer for all numeric fields (except dec) is (2^64-1) * 10^18
var maxIntegerValue = uint128.From64(math.MaxUint64).Mul64(1_000_000_000_000_000_000)

func parseNumberExtendedTo18Decimal(s string, dec uint64) (uint128.Uint128, error) {
	parts := strings.Split(s, ".")
	if len(parts) > 1 {
		return uint128.Uint128{}, errors.New("cannot parse decimal number: too many decimal points")
	}
	wholePart := parts[0]
	var decimalPart string
	if len(parts) == 1 {
		decimalPart := parts[1]
		if len(decimalPart) == 0 || len(decimalPart) > int(dec) {
			return uint128.Uint128{}, errors.New("invalid decimal part")
		}
	}
	// pad decimal part with zeros until 18 digits
	decimalPart += strings.Repeat("0", 18-len(decimalPart))
	number, err := uint128.FromString(wholePart + decimalPart)
	if err != nil {
		if errors.Is(err, uint128.ErrValueOverflow) {
			return uint128.Uint128{}, errors.WithStack(ErrNumberOverflow)
		}
		return uint128.Uint128{}, errors.Wrap(err, "failed to parse number")
	}
	if number.Cmp(maxIntegerValue) > 0 {
		return uint128.Uint128{}, errors.WithStack(ErrNumberOverflow)
	}
	return number, nil
}

var powerOfTens = []uint64{
	1e0,
	1e1,
	1e2,
	1e3,
	1e4,
	1e5,
	1e6,
	1e7,
	1e8,
	1e9,
	1e10,
	1e11,
	1e12,
	1e13,
	1e14,
	1e15,
	1e16,
	1e17,
	1e18,
}

func IsAmountWithinDecimals(amt uint128.Uint128, dec uint16) bool {
	if dec > 18 {
		return false
	}
	_, rem := amt.QuoRem64(powerOfTens[18-int(dec)])
	return rem != 0
}
