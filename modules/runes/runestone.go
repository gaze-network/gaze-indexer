package runes

import (
	"math"
	"math/big"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/lib/leb128"
	"github.com/samber/lo"
)

const PAYLOAD_MAGIC_NUMBER = txscript.OP_13

type Runestone struct {
	// Rune to etch in this transaction
	Etching *Etching
	// The rune ID of the runestone to mint in this transaction
	Mint *RuneId
	// List of edicts to execute in this transaction
	Edicts []Edict
	// Denotes the transaction output to allocate leftover runes to. If nil, use the first non-OP_RETURN output. If target output is OP_RETURN, those runes are burned.
	Pointer *uint64
	// If true, the runestone is a cenotaph. All minted runes in a cenotaph are burned. Runes etched in a cenotaph are not mintable.
	Cenotaph bool
	// Bitmask of flaws that caused the runestone to be a cenotaph
	Flaws Flaws
}

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

	flags, err := ParseFlags(uint64OrDefault(fields.Take(TagFlags)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse flags")
	}

	var etching *Etching
	if flags.Take(FlagEtching) {
		divisibilityBI := fields.Take(TagDivisibility)
		var divisibility uint8 = 0
		if divisibilityBI.Cmp(big.NewInt(int64(maxDivisibility))) <= 0 {
			divisibility = uint8(divisibilityBI.Uint64())
		}
		spacersBI := fields.Take(TagSpacers)
		var spacers uint32 = 0
		if spacersBI.Cmp(big.NewInt(int64(maxSpacers))) <= 0 {
			spacers = uint32(spacersBI.Uint64())
		}
		symbolBI := fields.Take(TagSymbol)
		var symbol rune
		if symbolBI.Cmp(big.NewInt(math.MaxInt32)) <= 0 {
			symbol = rune(symbolBI.Int64())
		}

		var terms *Terms
		if flags.Take(FlagTerms) {
			terms = &Terms{
				Amount:      fields.Take(TagAmount),
				Cap:         fields.Take(TagCap),
				HeightStart: uint64OrDefault(fields.Take(TagHeightStart)),
				HeightEnd:   uint64OrDefault(fields.Take(TagHeightEnd)),
				OffsetStart: uint64OrDefault(fields.Take(TagOffsetStart)),
				OffsetEnd:   uint64OrDefault(fields.Take(TagOffsetEnd)),
			}
		}

		etching = &Etching{
			Divisibility: divisibility,
			Premine:      fields.Take(TagPremine),
			Rune:         (*Rune)(fields.Take(TagRune)),
			Spacers:      spacers,
			Symbol:       symbol,
			Terms:        terms,
		}
	}

	var mint *RuneId
	mintValues := fields[TagMint]
	if len(mintValues) >= 2 {
		// TODO: switch to uint128 for rune ids
		mintRuneIdBlock := uint64OrDefault(fields.Take(TagMint))
		mintRuneIdTx := uint64OrDefault(fields.Take(TagMint))
		runeId, err := NewRuneId(mintRuneIdBlock, mintRuneIdTx)
		if err == nil {
			mint = &runeId
		}
	}
	var pointer *uint64
	pointerBI := fields.Take(TagPointer)
	if pointerBI.IsUint64() && pointerBI.Cmp(big.NewInt(int64(len(tx.TxOut)))) < 0 {
		pointer = lo.ToPtr(pointerBI.Uint64())
	}

	// TODO: Original implementation in Rust checks if supply overflows uint258. Need to check if we need to do the same here.

	if flags != 0 {
		flaws |= FlawFlagUnrecognizedFlag.Mask()
	}
	leftoverEvenTags := lo.Filter(lo.Keys(fields), func(tag Tag, _ int) bool {
		return tag%2 == 0
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

func decodeLEB128VarIntsFromPayload(payload []byte) ([]*big.Int, error) {
	integers := make([]*big.Int, 0)
	i := 0

	for i < len(payload) {
		n, length, err := leb128.DecodeLEB128(payload[i:])
		if err != nil {
			return nil, errors.Wrap(err, "cannot decode LEB128 varint")
		}
		if new(big.Int).Rsh(n, 128).Sign() != 0 {
			return nil, errors.WithStack(errs.OverflowUint128)
		}

		integers = append(integers, n)
		i += length
	}

	return integers, nil
}
