package ordinals

import (
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestParseEnvelopesFromTx(t *testing.T) {
	testTx := func(t *testing.T, tx *types.Transaction, expected []*Envelope) {
		t.Helper()

		envelopes := ParseEnvelopesFromTx(tx)
		assert.Equal(t, expected, envelopes)
	}
	testParseWitness := func(t *testing.T, tapScript []byte, expected []*Envelope) {
		t.Helper()

		tx := &types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn: []*types.TxIn{
				{
					Witness: wire.TxWitness{
						tapScript,
						{},
					},
				},
			},
		}
		testTx(t, tx, expected)
	}
	testEnvelope := func(t *testing.T, payload [][]byte, expected []*Envelope) {
		t.Helper()

		builder := NewPushScriptBuilder().
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF)
		for _, data := range payload {
			builder.AddData(data)
		}
		builder.AddOp(txscript.OP_ENDIF)
		script, err := builder.Script()
		assert.NoError(t, err)

		testParseWitness(
			t,
			script,
			expected,
		)
	}

	t.Run("empty_witness", func(t *testing.T) {
		testTx(t, &types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn: []*types.TxIn{{
				Witness: wire.TxWitness{},
			}},
		}, []*Envelope{})
	})
	t.Run("ignore_key_path_spends", func(t *testing.T) {
		testTx(t, &types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn: []*types.TxIn{{
				Witness: wire.TxWitness{
					utils.Must(NewPushScriptBuilder().
						AddOp(txscript.OP_FALSE).
						AddOp(txscript.OP_IF).
						AddData(protocolId).
						AddOp(txscript.OP_ENDIF).
						Script()),
				},
			}},
		}, []*Envelope{})
	})
	t.Run("ignore_key_path_spends_with_annex", func(t *testing.T) {
		testTx(t, &types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn: []*types.TxIn{{
				Witness: wire.TxWitness{
					utils.Must(NewPushScriptBuilder().
						AddOp(txscript.OP_FALSE).
						AddOp(txscript.OP_IF).
						AddData(protocolId).
						AddOp(txscript.OP_ENDIF).
						Script()),
					[]byte{txscript.TaprootAnnexTag},
				},
			}},
		}, []*Envelope{})
	})
	t.Run("parse_from_tapscript", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddOp(txscript.OP_ENDIF).
				Script()),
			[]*Envelope{{}},
		)
	})
	t.Run("ignore_unparsable_scripts", func(t *testing.T) {
		script := utils.Must(NewPushScriptBuilder().
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddData(protocolId).
			AddOp(txscript.OP_ENDIF).
			Script())

		script = append(script, 0x01)
		testParseWitness(
			t,
			script,
			[]*Envelope{
				{},
			},
		)
	})
	t.Run("no_inscription", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				Script()),
			[]*Envelope{},
		)
	})
	t.Run("duplicate_field", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagNop.Bytes(),
				{},
				TagNop.Bytes(),
				{},
			},
			[]*Envelope{
				{
					DuplicateField: true,
				},
			},
		)
	})
	t.Run("with_content_type", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagBody.Bytes(),
				[]byte("ord"),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("ord"),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("with_content_encoding", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagContentEncoding.Bytes(),
				[]byte("br"),
				TagBody.Bytes(),
				[]byte("ord"),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:         []byte("ord"),
						ContentType:     "text/plain;charset=utf-8",
						ContentEncoding: "br",
					},
				},
			},
		)
	})
	t.Run("with_unknown_tag", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagNop.Bytes(),
				[]byte("bar"),
				TagBody.Bytes(),
				[]byte("ord"),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("ord"),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("no_body", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("no_content_type", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagBody.Bytes(),
				[]byte("foo"),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content: []byte("foo"),
					},
				},
			},
		)
	})
	t.Run("valid_body_in_multiple_pushes", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagBody.Bytes(),
				[]byte("foo"),
				[]byte("bar"),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("foobar"),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("valid_body_in_zero_pushes", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagBody.Bytes(),
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte(""),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("valid_body_in_multiple_empty_pushes", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagBody.Bytes(),
				{},
				{},
				{},
				{},
				{},
				{},
				{},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte(""),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("valid_ignore_trailing", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddData(TagContentType.Bytes()).
				AddData([]byte("text/plain;charset=utf-8")).
				AddData(TagBody.Bytes()).
				AddData([]byte("ord")).
				AddOp(txscript.OP_ENDIF).
				AddOp(txscript.OP_CHECKSIG).
				Script()),
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("ord"),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("valid_ignore_preceding", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_CHECKSIG).
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddData(TagContentType.Bytes()).
				AddData([]byte("text/plain;charset=utf-8")).
				AddData(TagBody.Bytes()).
				AddData([]byte("ord")).
				AddOp(txscript.OP_ENDIF).
				Script()),
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("ord"),
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("multiple_inscriptions_in_a_single_witness", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddData(TagContentType.Bytes()).
				AddData([]byte("text/plain;charset=utf-8")).
				AddData(TagBody.Bytes()).
				AddData([]byte("foo")).
				AddOp(txscript.OP_ENDIF).
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddData(TagContentType.Bytes()).
				AddData([]byte("text/plain;charset=utf-8")).
				AddData(TagBody.Bytes()).
				AddData([]byte("bar")).
				AddOp(txscript.OP_ENDIF).
				Script()),
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("foo"),
						ContentType: "text/plain;charset=utf-8",
					},
				},
				{
					Inscription: Inscription{
						Content:     []byte("bar"),
						ContentType: "text/plain;charset=utf-8",
					},
					Offset: 1,
				},
			},
		)
	})
	t.Run("invalid_utf8_does_not_render_inscription_invalid", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("text/plain;charset=utf-8"),
				TagBody.Bytes(),
				{0b10000000},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte{0b10000000},
						ContentType: "text/plain;charset=utf-8",
					},
				},
			},
		)
	})
	t.Run("no_endif", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				Script()),
			[]*Envelope{},
		)
	})
	t.Run("no_op_false", func(t *testing.T) {
		testParseWitness(
			t,
			utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddOp(txscript.OP_ENDIF).
				Script()),
			[]*Envelope{},
		)
	})
	t.Run("empty_envelope", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{},
			[]*Envelope{},
		)
	})
	t.Run("wrong_protocol_identifier", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				[]byte("foo"),
			},
			[]*Envelope{},
		)
	})
	t.Run("extract_from_second_input", func(t *testing.T) {
		testTx(
			t,
			&types.Transaction{
				Version:  2,
				LockTime: 0,
				TxIn: []*types.TxIn{{}, {
					Witness: wire.TxWitness{
						utils.Must(NewPushScriptBuilder().
							AddOp(txscript.OP_FALSE).
							AddOp(txscript.OP_IF).
							AddData(protocolId).
							AddData(TagContentType.Bytes()).
							AddData([]byte("text/plain;charset=utf-8")).
							AddData(TagBody.Bytes()).
							AddData([]byte("ord")).
							AddOp(txscript.OP_ENDIF).
							Script(),
						),
						{},
					},
				}},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte("ord"),
						ContentType: "text/plain;charset=utf-8",
					},
					InputIndex: 1,
				},
			},
		)
	})
	t.Run("inscribe_png", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagContentType.Bytes(),
				[]byte("image/png"),
				TagBody.Bytes(),
				{0x01, 0x02, 0x03},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Content:     []byte{0x01, 0x02, 0x03},
						ContentType: "image/png",
					},
				},
			},
		)
	})
	t.Run("unknown_odd_fields", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagNop.Bytes(),
				{0x00},
			},
			[]*Envelope{
				{
					Inscription: Inscription{},
				},
			},
		)
	})
	t.Run("unknown_even_fields", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagUnbound.Bytes(),
				{0x00},
			},
			[]*Envelope{
				{
					Inscription:           Inscription{},
					UnrecognizedEvenField: true,
				},
			},
		)
	})
	t.Run("pointer_field_is_recognized", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagPointer.Bytes(),
				{0x01},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Pointer: lo.ToPtr(uint64(1)),
					},
				},
			},
		)
	})
	t.Run("duplicate_pointer_field_makes_inscription_unbound", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagPointer.Bytes(),
				{0x01},
				TagPointer.Bytes(),
				{0x00},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Pointer: lo.ToPtr(uint64(1)),
					},
					DuplicateField:        true,
					UnrecognizedEvenField: true,
				},
			},
		)
	})
	t.Run("incomplete_field", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagNop.Bytes(),
			},
			[]*Envelope{
				{
					Inscription:     Inscription{},
					IncompleteField: true,
				},
			},
		)
	})
	t.Run("metadata_is_parsed_correctly", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagMetadata.Bytes(),
				{},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Metadata: []byte{},
					},
				},
			},
		)
	})
	t.Run("metadata_is_parsed_correctly_from_chunks", func(t *testing.T) {
		testEnvelope(
			t,
			[][]byte{
				protocolId,
				TagMetadata.Bytes(),
				{0x00},
				TagMetadata.Bytes(),
				{0x01},
			},
			[]*Envelope{
				{
					Inscription: Inscription{
						Metadata: []byte{0x00, 0x01},
					},
					DuplicateField: true,
				},
			},
		)
	})
	t.Run("pushnum_opcodes_are_parsed_correctly", func(t *testing.T) {
		pushNumOpCodes := map[byte][]byte{
			txscript.OP_1NEGATE: {0x81},
			txscript.OP_1:       {0x01},
			txscript.OP_2:       {0x02},
			txscript.OP_3:       {0x03},
			txscript.OP_4:       {0x04},
			txscript.OP_5:       {0x05},
			txscript.OP_6:       {0x06},
			txscript.OP_7:       {0x07},
			txscript.OP_8:       {0x08},
			txscript.OP_9:       {0x09},
			txscript.OP_10:      {0x10},
			txscript.OP_11:      {0x11},
			txscript.OP_12:      {0x12},
			txscript.OP_13:      {0x13},
			txscript.OP_14:      {0x14},
			txscript.OP_15:      {0x15},
			txscript.OP_16:      {0x16},
		}
		for opCode, value := range pushNumOpCodes {
			script := utils.Must(NewPushScriptBuilder().
				AddOp(txscript.OP_FALSE).
				AddOp(txscript.OP_IF).
				AddData(protocolId).
				AddData(TagBody.Bytes()).
				AddOp(opCode).
				AddOp(txscript.OP_ENDIF).
				Script())

			testParseWitness(
				t,
				script,
				[]*Envelope{
					{
						Inscription: Inscription{
							Content: value,
						},
						PushNum: true,
					},
				},
			)
		}
	})
	t.Run("stuttering", func(t *testing.T) {
		script := utils.Must(NewPushScriptBuilder().
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddData(protocolId).
			AddOp(txscript.OP_ENDIF).
			Script())
		testParseWitness(
			t,
			script,
			[]*Envelope{
				{
					Inscription: Inscription{},
					Stutter:     true,
				},
			},
		)

		script = utils.Must(NewPushScriptBuilder().
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddData(protocolId).
			AddOp(txscript.OP_ENDIF).
			Script())
		testParseWitness(
			t,
			script,
			[]*Envelope{
				{
					Inscription: Inscription{},
					Stutter:     true,
				},
			},
		)

		script = utils.Must(NewPushScriptBuilder().
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_AND).
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddData(protocolId).
			AddOp(txscript.OP_ENDIF).
			Script())
		testParseWitness(
			t,
			script,
			[]*Envelope{
				{
					Inscription: Inscription{},
					Stutter:     false,
				},
			},
		)
	})
}
