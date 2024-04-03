package runes

import (
	"math/big"

	"github.com/btcsuite/btcd/wire"
	"github.com/samber/lo"
)

type Message struct {
	Flaws  Flaws
	Edicts []Edict
	Fields Fields
}

type Fields map[Tag][]*big.Int

func (fields Fields) Take(tag Tag) *big.Int {
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
	return first
}

func MessageFromIntegers(tx *wire.MsgTx, payload []*big.Int) Message {
	flaws := Flaws(0)
	edicts := make([]Edict, 0)
	fields := make(map[Tag][]*big.Int)

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
				if !blockDelta.IsUint64() || !txIndexDelta.IsUint64() {
					flaws |= FlawFlagEdictRuneId.Mask()
					break
				}
				if output.Cmp(big.NewInt(int64(len(tx.TxOut)))) > 0 {
					flaws |= FlawFlagEdictOutput.Mask()
					break
				}
				nextRuneId, err := runeId.Next(blockDelta.Uint64(), txIndexDelta.Uint64())
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
				runeId = nextRuneId
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
			fields[tag] = make([]*big.Int, 0)
		}
		fields[tag] = append(fields[tag], value)
	}

	return Message{
		Flaws:  flaws,
		Edicts: edicts,
		Fields: fields,
	}
}
