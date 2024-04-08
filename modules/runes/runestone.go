package runes

import (
	"math"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/lib/leb128"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

const PAYLOAD_MAGIC_NUMBER = txscript.OP_13

type Runestone struct {
	// Rune to etch in this transaction
	Etching *Etching
	// The rune ID of the runestone to mint in this transaction
	Mint *RuneId
	// Denotes the transaction output to allocate leftover runes to. If nil, use the first non-OP_RETURN output. If target output is OP_RETURN, those runes are burned.
	Pointer *uint64
	// List of edicts to execute in this transaction
	Edicts []Edict
	// If true, the runestone is a cenotaph. All minted runes in a cenotaph are burned. Runes etched in a cenotaph are not mintable.
	Cenotaph bool
	// Bitmask of flaws that caused the runestone to be a cenotaph
	Flaws Flaws
}

// DecipherRunestone deciphers a runestone from a transaction. If the runestone is a cenotaph, the runestone is returned with Cenotaph set to true and Flaws set to the bitmask of flaws that caused the runestone to be a cenotaph.
// If no runestone is found, nil is returned.
func DecipherRunestone(tx *wire.MsgTx) (*Runestone, error) {
	payload, flaws := runestonePayloadFromTx(tx)
	if flaws != 0 {
		return &Runestone{
			Cenotaph: true,
			Flaws:    flaws,
		}, nil
	}
	if payload == nil {
		return nil, nil
	}

	integers, err := decodeLEB128VarIntsFromPayload(payload)
	if err != nil {
		flaws |= FlawFlagVarInt.Mask()
		return &Runestone{
			Cenotaph: true,
			Flaws:    flaws,
		}, nil
	}
	message := MessageFromIntegers(tx, integers)
	edicts, fields := message.Edicts, message.Fields

	flags, err := ParseFlags(lo.FromPtr(fields.Take(TagFlags)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse flags")
	}

	var etching *Etching
	if flags.Take(FlagEtching) {
		divisibility := fields.Take(TagDivisibility)
		if divisibility != nil && divisibility.Cmp64(uint64(maxDivisibility)) > 0 {
			divisibility = nil
		}
		spacers := fields.Take(TagSpacers)
		if spacers != nil && spacers.Cmp64(uint64(maxSpacers)) > 0 {
			spacers = nil
		}
		symbol := fields.Take(TagSymbol)
		if symbol != nil && symbol.Cmp64(math.MaxInt32) > 0 {
			symbol = nil
		}

		var terms *Terms
		if flags.Take(FlagTerms) {
			var heightStart, heightEnd, offsetStart, offsetEnd *uint64
			if value := fields.Take(TagHeightStart); value != nil && value.IsUint64() {
				heightStart = lo.ToPtr(value.Uint64())
			}
			if value := fields.Take(TagHeightEnd); value != nil && value.IsUint64() {
				heightEnd = lo.ToPtr(value.Uint64())
			}
			if value := fields.Take(TagOffsetStart); value != nil && value.IsUint64() {
				offsetStart = lo.ToPtr(value.Uint64())
			}
			if value := fields.Take(TagOffsetEnd); value != nil && value.IsUint64() {
				offsetEnd = lo.ToPtr(value.Uint64())
			}
			terms = &Terms{
				Amount:      fields.Take(TagAmount),
				Cap:         fields.Take(TagCap),
				HeightStart: heightStart,
				HeightEnd:   heightEnd,
				OffsetStart: offsetStart,
				OffsetEnd:   offsetEnd,
			}
		}

		etching = &Etching{
			Divisibility: lo.ToPtr(divisibility.Uint8()),
			Premine:      fields.Take(TagPremine),
			Rune:         (*Rune)(fields.Take(TagRune)),
			Spacers:      lo.ToPtr(spacers.Uint32()),
			Symbol:       lo.ToPtr(rune(symbol.Uint32())),
			Terms:        terms,
		}
	}

	var mint *RuneId
	mintValues := fields[TagMint]
	if len(mintValues) >= 2 {
		mintRuneIdBlock := lo.FromPtr(fields.Take(TagMint))
		mintRuneIdTx := lo.FromPtr(fields.Take(TagMint))
		if mintRuneIdBlock.IsUint64() && mintRuneIdTx.IsUint32() {
			runeId, err := NewRuneId(mintRuneIdBlock.Uint64(), mintRuneIdTx.Uint32())
			if err == nil {
				mint = &runeId
			}
		}
	}
	var pointer *uint64
	pointerU128 := fields.Take(TagPointer)
	if pointerU128 != nil && pointerU128.Cmp64(uint64(len(tx.TxOut))) < 0 {
		pointer = lo.ToPtr(pointerU128.Uint64())
	}

	_, err = etching.Supply()
	if err != nil {
		if errors.Is(err, errs.OverflowUint128) {
			flaws |= FlawFlagSupplyOverflow.Mask()
		} else {
			return nil, errors.Wrap(err, "cannot calculate supply")
		}
	}

	if !flags.Uint128().IsZero() {
		flaws |= FlawFlagUnrecognizedFlag.Mask()
	}
	leftoverEvenTags := lo.Filter(lo.Keys(fields), func(tag Tag, _ int) bool {
		return tag.Uint128().Mod64(2) == 0
	})
	if len(leftoverEvenTags) != 0 {
		flaws |= FlawFlagUnrecognizedEvenTag.Mask()
	}
	if flaws != 0 {
		return &Runestone{
			Cenotaph: true,
			Flaws:    flaws,
			Mint:     mint,
			Etching:  etching,
		}, nil
	}

	return &Runestone{
		Etching: etching,
		Mint:    mint,
		Edicts:  edicts,
		Pointer: pointer,
	}, nil
}

func runestonePayloadFromTx(tx *wire.MsgTx) ([]byte, Flaws) {
	for _, output := range tx.TxOut {
		tokenizer := txscript.MakeScriptTokenizer(0, output.PkScript)

		// payload must start with OP_RETURN
		tokenizer.Next()
		if opCode := tokenizer.Opcode(); opCode != txscript.OP_RETURN {
			continue
		}

		// next opcode must be the magic number
		tokenizer.Next()
		if opCode := tokenizer.Opcode(); opCode != PAYLOAD_MAGIC_NUMBER {
			continue
		}

		// construct the payload by concatenating the remaining data pushes
		payload := make([]byte, 0)
		for tokenizer.Next() {
			data := tokenizer.Data()
			if data == nil {
				return nil, FlawFlagOpCode.Mask()
			}
			payload = append(payload, data...)
		}
		if tokenizer.Err() != nil {
			return nil, FlawFlagInvalidScript.Mask()
		}

		return payload, Flaws(0)
	}

	// if not found, return nil
	return nil, 0
}

func decodeLEB128VarIntsFromPayload(payload []byte) ([]uint128.Uint128, error) {
	integers := make([]uint128.Uint128, 0)
	i := 0

	for i < len(payload) {
		n, length, err := leb128.DecodeLEB128(payload[i:])
		if err != nil {
			return nil, errors.Wrap(err, "cannot decode LEB128 varint")
		}

		integers = append(integers, n)
		i += length
	}

	return integers, nil
}
