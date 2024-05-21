package ordinals

import (
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
)

// PushScriptBuilder is a helper to build scripts that requires data pushes to use OP_PUSHDATA* or OP_DATA_* opcodes only.
// Empty data pushes are still encoded as OP_0.
type PushScriptBuilder struct {
	script []byte
	err    error
}

func NewPushScriptBuilder() *PushScriptBuilder {
	return &PushScriptBuilder{}
}

// canonicalDataSize returns the number of bytes the canonical encoding of the
// data will take.
func canonicalDataSize(data []byte) int {
	dataLen := len(data)

	// When the data consists of a single number that can be represented
	// by one of the "small integer" opcodes, that opcode will be instead
	// of a data push opcode followed by the number.
	if dataLen == 0 {
		return 1
	}

	if dataLen < txscript.OP_PUSHDATA1 {
		return 1 + dataLen
	} else if dataLen <= 0xff {
		return 2 + dataLen
	} else if dataLen <= 0xffff {
		return 3 + dataLen
	}

	return 5 + dataLen
}

func pushDataToBytes(data []byte) []byte {
	if len(data) == 0 {
		return []byte{txscript.OP_0}
	}
	script := make([]byte, 0)
	dataLen := len(data)
	if dataLen < txscript.OP_PUSHDATA1 {
		script = append(script, byte(txscript.OP_DATA_1-1+dataLen))
	} else if dataLen <= 0xff {
		script = append(script, txscript.OP_PUSHDATA1, byte(dataLen))
	} else if dataLen <= 0xffff {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(dataLen))
		script = append(script, txscript.OP_PUSHDATA2)
		script = append(script, buf...)
	} else {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(dataLen))
		script = append(script, txscript.OP_PUSHDATA4)
		script = append(script, buf...)
	}
	// Append the actual data.
	script = append(script, data...)
	return script
}

// AddData pushes the passed data to the end of the script.  It automatically
// chooses canonical opcodes depending on the length of the data.  A zero length
// buffer will lead to a push of empty data onto the stack (OP_0) and any push
// of data greater than MaxScriptElementSize will not modify the script since
// that is not allowed by the script engine.  Also, the script will not be
// modified if pushing the data would cause the script to exceed the maximum
// allowed script engine size.
func (b *PushScriptBuilder) AddData(data []byte) *PushScriptBuilder {
	if b.err != nil {
		return b
	}
	// Pushes that would cause the script to exceed the largest allowed
	// script size would result in a non-canonical script.
	dataSize := canonicalDataSize(data)
	if len(b.script)+dataSize > txscript.MaxScriptSize {
		str := fmt.Sprintf("adding %d bytes of data would exceed the "+
			"maximum allowed canonical script length of %d",
			dataSize, txscript.MaxScriptSize)
		b.err = txscript.ErrScriptNotCanonical(str)
		return b
	}

	// Pushes larger than the max script element size would result in a
	// script that is not canonical.
	dataLen := len(data)
	if dataLen > txscript.MaxScriptElementSize {
		str := fmt.Sprintf("adding a data element of %d bytes would "+
			"exceed the maximum allowed script element size of %d",
			dataLen, txscript.MaxScriptElementSize)
		b.err = txscript.ErrScriptNotCanonical(str)
		return b
	}

	b.script = append(b.script, pushDataToBytes(data)...)
	return b
}

// AddFullData should not typically be used by ordinary users as it does not
// include the checks which prevent data pushes larger than the maximum allowed
// sizes which leads to scripts that can't be executed.  This is provided for
// testing purposes such as regression tests where sizes are intentionally made
// larger than allowed.
//
// Use AddData instead.
func (b *PushScriptBuilder) AddFullData(data []byte) *PushScriptBuilder {
	if b.err != nil {
		return b
	}

	b.script = append(b.script, pushDataToBytes(data)...)
	return b
}

// AddOp pushes the passed opcode to the end of the script.  The script will not
// be modified if pushing the opcode would cause the script to exceed the
// maximum allowed script engine size.
func (b *PushScriptBuilder) AddOp(opcode byte) *PushScriptBuilder {
	if b.err != nil {
		return b
	}

	// Pushes that would cause the script to exceed the largest allowed
	// script size would result in a non-canonical script.
	if len(b.script)+1 > txscript.MaxScriptSize {
		str := fmt.Sprintf("adding an opcode would exceed the maximum "+
			"allowed canonical script length of %d", txscript.MaxScriptSize)
		b.err = txscript.ErrScriptNotCanonical(str)
		return b
	}

	b.script = append(b.script, opcode)
	return b
}

// AddOps pushes the passed opcodes to the end of the script.  The script will
// not be modified if pushing the opcodes would cause the script to exceed the
// maximum allowed script engine size.
func (b *PushScriptBuilder) AddOps(opcodes []byte) *PushScriptBuilder {
	if b.err != nil {
		return b
	}

	// Pushes that would cause the script to exceed the largest allowed
	// script size would result in a non-canonical script.
	if len(b.script)+len(opcodes) > txscript.MaxScriptSize {
		str := fmt.Sprintf("adding opcodes would exceed the maximum "+
			"allowed canonical script length of %d", txscript.MaxScriptSize)
		b.err = txscript.ErrScriptNotCanonical(str)
		return b
	}

	b.script = append(b.script, opcodes...)
	return b
}

// Script returns the currently built script.  When any errors occurred while
// building the script, the script will be returned up the point of the first
// error along with the error.
func (b *PushScriptBuilder) Script() ([]byte, error) {
	return b.script, b.err
}
