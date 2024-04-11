package runes

import (
	"math/bits"
	"strings"

	"github.com/cockroachdb/errors"
)

type SpacedRune struct {
	Rune    Rune
	Spacers uint32
}

func NewSpacedRune(rune Rune, spacers uint32) SpacedRune {
	return SpacedRune{
		Rune:    rune,
		Spacers: spacers,
	}
}

var (
	ErrLeadingSpacer              = errors.New("runes cannot start with a spacer")
	ErrTrailingSpacer             = errors.New("runes cannot end with a spacer")
	ErrDoubleSpacer               = errors.New("runes cannot have more than one spacer between characters")
	ErrInvalidSpacedRuneCharacter = errors.New("invalid spaced rune character: must satisfy regex [A-Z•.]")
)

func NewSpacedRuneFromString(input string) (SpacedRune, error) {
	var sb strings.Builder
	var spacers uint32

	for _, c := range input {
		if c >= 'A' && c <= 'Z' {
			sb.WriteRune(c)
			continue
		}
		if c == '•' || c == '.' {
			if sb.Len() == 0 {
				return SpacedRune{}, errors.WithStack(ErrLeadingSpacer)
			}
			flag := 1 << (sb.Len() - 1)
			if spacers&uint32(flag) != 0 {
				return SpacedRune{}, errors.WithStack(ErrDoubleSpacer)
			}
			spacers |= 1 << (sb.Len() - 1)
			continue
		}
		return SpacedRune{}, errors.WithStack(ErrInvalidSpacedRuneCharacter)
	}

	if 32-bits.LeadingZeros32(spacers) >= sb.Len() {
		return SpacedRune{}, errors.WithStack(ErrTrailingSpacer)
	}
	rune, err := NewRuneFromString(sb.String())
	if err != nil {
		return SpacedRune{}, errors.Wrap(err, "failed to parse rune from string")
	}
	return NewSpacedRune(rune, spacers), nil
}

func (r SpacedRune) String() string {
	runeStr := r.Rune.String()
	var sb strings.Builder
	for i, c := range runeStr {
		sb.WriteRune(c)
		if i < len(runeStr)-1 && r.Spacers&(1<<i) != 0 {
			sb.WriteRune('•')
		}
	}
	return sb.String()
}
