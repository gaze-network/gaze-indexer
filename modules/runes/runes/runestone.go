package runes

import (
	"fmt"
	"log"
	"slices"
	"unicode/utf8"

	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/leb128"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

const (
	RUNESTONE_PAYLOAD_MAGIC_NUMBER = txscript.OP_13
	RUNE_COMMIT_BLOCKS             = 6
)

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

// Encipher encodes a runestone into a scriptPubKey, ready to be put into a transaction output.
func (r Runestone) Encipher() ([]byte, error) {
	var payload []byte

	encodeUint128 := func(value uint128.Uint128) {
		payload = append(payload, leb128.EncodeUint128(value)...)
	}
	encodeTagValues := func(tag Tag, values ...uint128.Uint128) {
		for _, value := range values {
			// encode tag key
			encodeUint128(tag.Uint128())
			// encode tag value
			encodeUint128(value)
		}
	}
	encodeEdict := func(previousRuneId RuneId, edict Edict) {
		blockHeight, txIndex := previousRuneId.Delta(edict.Id)
		encodeUint128(uint128.From64(blockHeight))
		encodeUint128(uint128.From64(uint64(txIndex)))
		encodeUint128(edict.Amount)
		encodeUint128(uint128.From64(uint64(edict.Output)))
	}

	if r.Etching != nil {
		etching := r.Etching
		flags := Flags(uint128.Zero)
		flags.Set(FlagEtching)
		if etching.Terms != nil {
			flags.Set(FlagTerms)
		}
		if etching.Turbo {
			flags.Set(FlagTurbo)
		}
		encodeTagValues(TagFlags, flags.Uint128())

		if etching.Rune != nil {
			encodeTagValues(TagRune, etching.Rune.Uint128())
		}
		if etching.Divisibility != nil {
			encodeTagValues(TagDivisibility, uint128.From64(uint64(*etching.Divisibility)))
		}
		if etching.Spacers != nil {
			encodeTagValues(TagSpacers, uint128.From64(uint64(*etching.Spacers)))
		}
		if etching.Symbol != nil {
			encodeTagValues(TagSymbol, uint128.From64(uint64(*etching.Symbol)))
		}
		if etching.Premine != nil {
			encodeTagValues(TagPremine, *etching.Premine)
		}
		if etching.Terms != nil {
			terms := etching.Terms
			if terms.Amount != nil {
				encodeTagValues(TagAmount, *terms.Amount)
			}
			if terms.Cap != nil {
				encodeTagValues(TagCap, *terms.Cap)
			}
			if terms.HeightStart != nil {
				encodeTagValues(TagHeightStart, uint128.From64(*terms.HeightStart))
			}
			if terms.HeightEnd != nil {
				encodeTagValues(TagHeightEnd, uint128.From64(*terms.HeightEnd))
			}
			if terms.OffsetStart != nil {
				encodeTagValues(TagOffsetStart, uint128.From64(*terms.OffsetStart))
			}
			if terms.OffsetEnd != nil {
				encodeTagValues(TagOffsetEnd, uint128.From64(*terms.OffsetEnd))
			}
		}
	}

	if r.Mint != nil {
		encodeTagValues(TagMint, uint128.From64(r.Mint.BlockHeight), uint128.From64(uint64(r.Mint.TxIndex)))
	}
	if r.Pointer != nil {
		encodeTagValues(TagPointer, uint128.From64(*r.Pointer))
	}
	if len(r.Edicts) > 0 {
		encodeUint128(TagBody.Uint128())
		edicts := make([]Edict, len(r.Edicts))
		copy(edicts, r.Edicts)
		// sort by block height first, then by tx index
		slices.SortFunc(edicts, func(i, j Edict) int {
			if i.Id.BlockHeight != j.Id.BlockHeight {
				return int(i.Id.BlockHeight) - int(j.Id.BlockHeight)
			}
			return int(i.Id.TxIndex) - int(j.Id.TxIndex)
		})
		var previousRuneId RuneId
		for _, edict := range edicts {
			encodeEdict(previousRuneId, edict)
			previousRuneId = edict.Id
		}
	}

	sb := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).
		AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER)

	// chunk payload to MaxScriptElementSize
	for _, chunk := range lo.Chunk(payload, txscript.MaxScriptElementSize) {
		sb.AddData(chunk)
	}

	scriptPubKey, err := sb.Script()
	if err != nil {
		return nil, errors.Wrap(err, "cannot build scriptPubKey")
	}
	return scriptPubKey, nil
}

// DecipherRunestone deciphers a runestone from a transaction. If the runestone is a cenotaph, the runestone is returned with Cenotaph set to true and Flaws set to the bitmask of flaws that caused the runestone to be a cenotaph.
// If no runestone is found, nil is returned.
func DecipherRunestone(tx *types.Transaction) (*Runestone, error) {
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
		log.Printf("warning: %v\n", err)
		flaws |= FlawFlagVarInt.Mask()
		return &Runestone{
			Cenotaph: true,
			Flaws:    flaws,
		}, nil
	}
	message := MessageFromIntegers(tx, integers)
	edicts, fields := message.Edicts, message.Fields
	flaws |= message.Flaws

	flags, err := ParseFlags(lo.FromPtr(fields.Take(TagFlags)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse flags")
	}

	var etching *Etching
	if flags.Take(FlagEtching) {
		divisibilityU128 := fields.Take(TagDivisibility)
		if divisibilityU128 != nil && divisibilityU128.Cmp64(uint64(maxDivisibility)) > 0 {
			divisibilityU128 = nil
		}
		spacersU128 := fields.Take(TagSpacers)
		if spacersU128 != nil && spacersU128.Cmp64(uint64(maxSpacers)) > 0 {
			spacersU128 = nil
		}
		symbolU128 := fields.Take(TagSymbol)
		if symbolU128 != nil && symbolU128.Cmp64(utf8.MaxRune) > 0 {
			symbolU128 = nil
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

		var divisibility *uint8
		if divisibilityU128 != nil {
			divisibility = lo.ToPtr(divisibilityU128.Uint8())
		}
		var spacers *uint32
		if spacersU128 != nil {
			spacers = lo.ToPtr(spacersU128.Uint32())
		}
		var symbol *rune
		if symbolU128 != nil {
			symbol = lo.ToPtr(rune(symbolU128.Uint32()))
		}

		etching = &Etching{
			Divisibility: divisibility,
			Premine:      fields.Take(TagPremine),
			Rune:         (*Rune)(fields.Take(TagRune)),
			Spacers:      spacers,
			Symbol:       symbol,
			Terms:        terms,
			Turbo:        flags.Take(FlagTurbo),
		}
	}

	var mint *RuneId
	mintValues := fields[TagMint]
	if len(mintValues) >= 2 {
		mintRuneIdBlock := lo.FromPtr(fields.Take(TagMint))
		mintRuneIdTx := lo.FromPtr(fields.Take(TagMint))
		if mintRuneIdBlock.IsUint64() && mintRuneIdTx.IsUint32() {
			runeId, err := NewRuneId(mintRuneIdBlock.Uint64(), mintRuneIdTx.Uint32())
			if err != nil {
				// invalid mint
				flaws |= FlawFlagUnrecognizedEvenTag.Mask()
			} else {
				mint = &runeId
			}
		}
	}
	var pointer *uint64
	pointerU128 := fields.Take(TagPointer)
	if pointerU128 != nil {
		if pointerU128.Cmp64(uint64(len(tx.TxOut))) < 0 {
			pointer = lo.ToPtr(pointerU128.Uint64())
		} else {
			// invalid pointer
			flaws |= FlawFlagUnrecognizedEvenTag.Mask()
		}
	}

	if etching != nil {
		_, err = etching.Supply()
		if err != nil {
			if errors.Is(err, errs.OverflowUint128) {
				flaws |= FlawFlagSupplyOverflow.Mask()
			} else {
				return nil, errors.Wrap(err, "cannot calculate supply")
			}
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
		var cenotaphEtching *Etching
		if etching != nil && etching.Rune != nil {
			cenotaphEtching = &Etching{
				Rune: etching.Rune,
			}
		}
		return &Runestone{
			Cenotaph: true,
			Flaws:    flaws,
			Mint:     mint,
			Etching:  cenotaphEtching, // return etching with only Rune field if runestone is cenotaph
		}, nil
	}

	return &Runestone{
		Etching: etching,
		Mint:    mint,
		Edicts:  edicts,
		Pointer: pointer,
	}, nil
}

func runestonePayloadFromTx(tx *types.Transaction) ([]byte, Flaws) {
	for _, output := range tx.TxOut {
		tokenizer := txscript.MakeScriptTokenizer(0, output.PkScript)

		// payload must start with OP_RETURN
		if ok := tokenizer.Next(); !ok {
			// script ended
			continue
		}
		if err := tokenizer.Err(); err != nil {
			continue
		}
		if opCode := tokenizer.Opcode(); opCode != txscript.OP_RETURN {
			continue
		}

		// next opcode must be the magic number
		if ok := tokenizer.Next(); !ok {
			// script ended
			continue
		}
		if err := tokenizer.Err(); err != nil {
			fmt.Println(err.Error())
			continue
		}
		if opCode := tokenizer.Opcode(); opCode != RUNESTONE_PAYLOAD_MAGIC_NUMBER {
			continue
		}

		// this output is now selected to be the runestone output. Any errors from now on will be considered a flaw.

		// construct the payload by concatenating the remaining data pushes
		payload := make([]byte, 0)
		for tokenizer.Next() {
			if tokenizer.Err() != nil {
				return nil, FlawFlagInvalidScript.Mask()
			}
			if !IsDataPushOpCode(tokenizer.Opcode()) {
				return nil, FlawFlagOpCode.Mask()
			}
			payload = append(payload, tokenizer.Data()...)
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
		n, length, err := leb128.DecodeUint128(payload[i:])
		if err != nil {
			return nil, errors.Wrap(err, "cannot decode LEB128 varint")
		}

		integers = append(integers, n)
		i += length
	}

	return integers, nil
}

func IsDataPushOpCode(opCode byte) bool {
	// includes OP_0, OP_DATA_1 to OP_DATA_75, OP_PUSHDATA1, OP_PUSHDATA2, OP_PUSHDATA4
	return opCode <= txscript.OP_PUSHDATA4
}
