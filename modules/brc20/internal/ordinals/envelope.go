package ordinals

import (
	"bytes"
	"encoding/binary"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/samber/lo"
)

type Envelope struct {
	Inscription           Inscription
	InputIndex            uint32 // Index of input that contains the envelope
	Offset                int    // Number of envelope in the input
	PushNum               bool   // True if envelope contains pushnum opcodes
	Stutter               bool   // True if envelope matches stuttering curse structure
	IncompleteField       bool   // True if payload is incomplete
	DuplicateField        bool   // True if payload contains duplicated field
	UnrecognizedEvenField bool   // True if payload contains unrecognized even field
}

func ParseEnvelopesFromTx(tx *wire.MsgTx) []*Envelope {
	envelopes := make([]*Envelope, 0)

	for i, txIn := range tx.TxIn {
		tapScript, ok := extractTapScript(txIn.Witness)
		if !ok {
			continue
		}

		newEnvelopes := envelopesFromTapScript(tapScript, i)
		envelopes = append(envelopes, newEnvelopes...)
	}

	return envelopes
}

const protocolId = "ord"

func envelopesFromTapScript(tokenizer txscript.ScriptTokenizer, inputIndex int) []*Envelope {
	envelopes := make([]*Envelope, 0)

	var stuttered bool
	for tokenizer.Next() {
		if tokenizer.Err() != nil {
			break
		}
		if tokenizer.Opcode() == txscript.OP_FALSE {
			envelope, stutter := envelopeFromTokenizer(tokenizer, inputIndex, len(envelopes), stuttered)
			if envelope != nil {
				envelopes = append(envelopes, envelope)
			} else {
				stuttered = stutter
			}
		}
	}
	if tokenizer.Err() != nil {
		logger.Warn("failed to parse tapscript", slogx.Error(tokenizer.Err()))
		return envelopes
	}
	return envelopes
}

func envelopeFromTokenizer(tokenizer txscript.ScriptTokenizer, inputIndex int, offset int, stuttered bool) (*Envelope, bool) {
	tokenizer.Next()
	if tokenizer.Opcode() != txscript.OP_IF {
		return nil, tokenizer.Opcode() == txscript.OP_FALSE
	}

	tokenizer.Next()
	if !bytes.Equal(tokenizer.Data(), []byte(protocolId)) {
		return nil, tokenizer.Opcode() == txscript.OP_FALSE
	}

	var pushNum bool
	payload := make([][]byte, 0)
	for tokenizer.Next() {
		if tokenizer.Err() != nil {
			return nil, false
		}
		opCode := tokenizer.Opcode()
		if opCode == txscript.OP_ENDIF {
			break
		}
		switch opCode {
		case txscript.OP_1NEGATE:
			pushNum = true
			payload = append(payload, []byte{0x81})
		case txscript.OP_1:
			pushNum = true
			payload = append(payload, []byte{0x01})
		case txscript.OP_2:
			pushNum = true
			payload = append(payload, []byte{0x02})
		case txscript.OP_3:
			pushNum = true
			payload = append(payload, []byte{0x03})
		case txscript.OP_4:
			pushNum = true
			payload = append(payload, []byte{0x04})
		case txscript.OP_5:
			pushNum = true
			payload = append(payload, []byte{0x05})
		case txscript.OP_6:
			pushNum = true
			payload = append(payload, []byte{0x06})
		case txscript.OP_7:
			pushNum = true
			payload = append(payload, []byte{0x07})
		case txscript.OP_8:
			pushNum = true
			payload = append(payload, []byte{0x08})
		case txscript.OP_9:
			pushNum = true
			payload = append(payload, []byte{0x09})
		case txscript.OP_10:
			pushNum = true
			payload = append(payload, []byte{0x10})
		case txscript.OP_11:
			pushNum = true
			payload = append(payload, []byte{0x11})
		case txscript.OP_12:
			pushNum = true
			payload = append(payload, []byte{0x12})
		case txscript.OP_13:
			pushNum = true
			payload = append(payload, []byte{0x13})
		case txscript.OP_14:
			pushNum = true
			payload = append(payload, []byte{0x14})
		case txscript.OP_15:
			pushNum = true
			payload = append(payload, []byte{0x15})
		case txscript.OP_16:
			pushNum = true
			payload = append(payload, []byte{0x16})
		case txscript.OP_0:
			// OP_0 is a special case, it is accepted in ord's implementation
			payload = append(payload, []byte{})
		default:
			data := tokenizer.Data()
			if data == nil {
				return nil, false
			}
			payload = append(payload, data)
		}
	}
	// incomplete envelope
	if tokenizer.Done() {
		return nil, false
	}

	_, bodyIndex, ok := lo.FindIndexOf(payload, func(value []byte) bool {
		return len(value) == 0
	})
	var fieldPayloads, bodyPayloads [][]byte
	if ok {
		fieldPayloads = payload[:]
	} else {
		fieldPayloads = payload[:bodyIndex]
		bodyPayloads = payload[bodyIndex+1:]
	}

	var incompleteField bool
	fields := make(Fields)
	for _, chunk := range lo.Chunk(fieldPayloads, 2) {
		if len(chunk) != 2 {
			incompleteField = true
			break
		}
		key := chunk[0]
		value := chunk[1]
		// key cannot be empty, as checked by bodyIndex above
		// if key exceeds 1 byte, it would not match any tags
		if len(key) > 1 {
			incompleteField = true
			continue
		}
		tag := Tag(key[0])
		fields[tag] = append(fields[tag], value)
	}

	var duplicateField bool
	for _, values := range fields {
		if len(values) > 1 {
			duplicateField = true
			break
		}
	}

	rawContentEncoding := fields.Take(TagContentEncoding)
	rawContentType := fields.Take(TagContentType)
	rawDelegate := fields.Take(TagDelegate)
	rawMetadata := fields.Take(TagMetadata)
	rawMetaprotocol := fields.Take(TagMetaprotocol)
	rawParent := fields.Take(TagParent)
	rawPointer := fields.Take(TagPointer)

	unrecognizedEvenField := lo.SomeBy(lo.Keys(fields), func(key Tag) bool {
		return key%2 == 0
	})

	body := lo.Flatten(bodyPayloads)

	var delegate, parent *InscriptionId
	inscriptionId, err := NewInscriptionIdFromString(string(rawDelegate))
	if err == nil {
		delegate = &inscriptionId
	}
	inscriptionId, err = NewInscriptionIdFromString(string(rawParent))
	if err == nil {
		parent = &inscriptionId
	}

	var pointer *uint64
	// if rawPointer fits in uint64
	if len(rawPointer) <= 8 || lo.EveryBy(rawPointer[8:], func(value byte) bool {
		return value != 0
	}) {
		pointer = lo.ToPtr(binary.LittleEndian.Uint64(rawPointer))
	}

	inscription := Inscription{
		Content:         body,
		ContentEncoding: string(rawContentEncoding),
		ContentType:     string(rawContentType),
		Delegate:        delegate,
		Metadata:        rawMetadata,
		Metaprotocol:    string(rawMetaprotocol),
		Parent:          parent,
		Pointer:         pointer,
	}
	return &Envelope{
		Inscription:           inscription,
		InputIndex:            uint32(inputIndex),
		Offset:                offset,
		PushNum:               pushNum,
		Stutter:               stuttered,
		IncompleteField:       incompleteField,
		DuplicateField:        duplicateField,
		UnrecognizedEvenField: unrecognizedEvenField,
	}, false
}

type Fields map[Tag][][]byte

func (fields Fields) Take(tag Tag) []byte {
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

func isDataPushOpCode(opCode byte) bool {
	// includes OP_0, OP_DATA_1 to OP_DATA_75, OP_PUSHDATA1, OP_PUSHDATA2, OP_PUSHDATA4
	return opCode <= txscript.OP_PUSHDATA4
}

func extractTapScript(witness [][]byte) (txscript.ScriptTokenizer, bool) {
	witness = removeAnnexFromWitness(witness)
	if len(witness) < 2 {
		return txscript.ScriptTokenizer{}, false
	}
	script := witness[len(witness)-2]

	return txscript.MakeScriptTokenizer(0, script), true
}

func removeAnnexFromWitness(witness [][]byte) [][]byte {
	if len(witness) >= 2 && len(witness[len(witness)-1]) > 0 && witness[len(witness)-1][0] == txscript.TaprootAnnexTag {
		return witness[:len(witness)-1]
	}
	return witness
}
