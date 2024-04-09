package runes

import (
	"math"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

type Message struct {
	Fields Fields
	Edicts []Edict
	Flaws  Flaws
}

type Fields map[Tag][]uint128.Uint128

func (fields Fields) Take(tag Tag) *uint128.Uint128 {
	values, ok := fields[tag]
	if !ok {
		return nil
	}
	first := values[0]
	values = values[1:]
	if len(values) == 0 {
		delete(fields, tag)
	} else {
		fields[tag] = values
	}
	return &first
}

func MessageFromIntegers(tx *wire.MsgTx, payload []uint128.Uint128) Message {
	flaws := Flaws(0)
	var edicts []Edict
	fields := make(map[Tag][]uint128.Uint128)

	for i := 0; i < len(payload); i += 2 {
		tag, err := ParseTag(payload[i])
		if err != nil {
			continue
		}

		// If tag is Body, treat all remaining integers are edicts
		if tag == TagBody {
			runeId := RuneId{}
			for _, chunk := range lo.Chunk(payload[i+1:], 4) {
				if len(chunk) != 4 {
					flaws |= FlawFlagTrailingIntegers.Mask()
					break
				}
				blockDelta, txIndexDelta, amount, output := chunk[0], chunk[1], chunk[2], chunk[3]
				if blockDelta.Cmp64(math.MaxUint64) > 0 || blockDelta.Cmp64(math.MaxUint32) > 0 {
					flaws |= FlawFlagEdictRuneId.Mask()
					break
				}
				if output.Cmp64(uint64(len(tx.TxOut))) > 0 {
					flaws |= FlawFlagEdictOutput.Mask()
					break
				}
				runeId, err = runeId.Next(blockDelta.Uint64(), txIndexDelta.Uint32()) // safe to cast as uint32 because we checked
				if err != nil {
					flaws |= FlawFlagEdictRuneId.Mask()
					break
				}
				edict := Edict{
					Id:     runeId,
					Amount: amount,
					Output: int(output.Uint64()),
				}
				edicts = append(edicts, edict)
			}
			break
		}

		// append tag value to fields
		if i+1 >= len(payload) {
			flaws |= Flaws(FlawFlagTruncatedField)
			break
		}
		value := payload[i+1]
		if _, ok := fields[tag]; !ok {
			fields[tag] = make([]uint128.Uint128, 0)
		}
		fields[tag] = append(fields[tag], value)
	}

	return Message{
		Flaws:  flaws,
		Edicts: edicts,
		Fields: fields,
	}
}
