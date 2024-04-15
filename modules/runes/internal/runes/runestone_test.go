package runes

import (
	"math"
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/leb128"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func encodeLEB128VarIntsToPayload(integers []uint128.Uint128) []byte {
	payload := make([]byte, 0)
	for _, integer := range integers {
		payload = append(payload, leb128.EncodeUint128(integer)...)
	}
	return payload
}

func TestDecipherRunestone(t *testing.T) {
	testDecipherTx := func(t *testing.T, tx *types.Transaction, expected *Runestone) {
		runestone, err := DecipherRunestone(tx)
		assert.NoError(t, err)
		assert.Equal(t, expected, runestone)
	}

	testDecipherInteger := func(t *testing.T, integers []uint128.Uint128, expected *Runestone) {
		payload := encodeLEB128VarIntsToPayload(integers)
		pkScript, err := txscript.NewScriptBuilder().
			AddOp(txscript.OP_RETURN).
			AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
			AddData(payload).
			Script()
		assert.NoError(t, err)
		tx := &types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn:     []*types.TxIn{},
			TxOut: []*types.TxOut{
				{
					PkScript: pkScript,
					Value:    0,
				},
			},
		}
		testDecipherTx(t, tx, expected)
	}

	testDecipherPkScript := func(t *testing.T, pkScript []byte, expected *Runestone) {
		tx := &types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn:     []*types.TxIn{},
			TxOut: []*types.TxOut{
				{
					PkScript: pkScript,
					Value:    0,
				},
			},
		}
		testDecipherTx(t, tx, expected)
	}

	t.Run("decipher_returns_none_if_first_opcode_is_malformed", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().AddOp(txscript.OP_DATA_4).Script()),
			nil,
		)
	})
	t.Run("deciphering_transaction_with_non_op_return_output_returns_none", func(t *testing.T) {
		testDecipherTx(
			t,
			&types.Transaction{
				Version:  2,
				LockTime: 0,
				TxIn:     []*types.TxIn{},
				TxOut:    []*types.TxOut{},
			},
			nil,
		)
	})
	t.Run("deciphering_transaction_with_bare_op_return_returns_none", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).Script()),
			nil,
		)
	})
	t.Run("deciphering_transaction_with_non_matching_op_return_returns_none", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddOp(txscript.OP_1).Script()),
			nil,
		)
	})
	t.Run("deciphering_valid_runestone_with_invalid_script_postfix_returns_invalid_payload", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().
				AddOp(txscript.OP_RETURN).
				AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
				AddOp(txscript.OP_DATA_4).
				Script()),
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagInvalidScript.Mask(),
			},
		)
	})
	t.Run("deciphering_runestone_with_truncated_varint_is_cenotaph", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().
				AddOp(txscript.OP_RETURN).
				AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
				AddData([]byte{128}).
				Script()),
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagVarInt.Mask(),
			},
		)
	})
	t.Run("outputs_with_non_pushdata_opcodes_are_cenotaph_1", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().
				AddOp(txscript.OP_RETURN).
				AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
				AddOp(txscript.OP_VERIFY).
				AddData([]byte{0}).
				AddData(leb128.EncodeUint128(uint128.From64(1))).
				AddData(leb128.EncodeUint128(uint128.From64(1))).
				AddData([]byte{2, 0}).
				Script()),
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagOpCode.Mask(),
			},
		)
	})
	t.Run("outputs_with_non_pushdata_opcodes_are_cenotaph_2", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().
				AddOp(txscript.OP_RETURN).
				AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
				AddData([]byte{0}).
				AddData(leb128.EncodeUint128(uint128.From64(1))).
				AddData(leb128.EncodeUint128(uint128.From64(2))).
				AddData([]byte{3, 0}).
				Script()),
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagOpCode.Mask(),
			},
		)
	})
	t.Run("pushnum_opcodes_in_runestone_produce_cenotaph", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().
				AddOp(txscript.OP_RETURN).
				AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
				AddOp(txscript.OP_1).
				Script()),
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagOpCode.Mask(),
			},
		)
	})
	t.Run("deciphering_empty_runestone_is_successful", func(t *testing.T) {
		testDecipherPkScript(
			t,
			utils.Must(txscript.NewScriptBuilder().
				AddOp(txscript.OP_RETURN).
				AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
				Script()),
			&Runestone{},
		)
	})
	t.Run("invalid_input_scripts_are_skipped_when_searching_for_runestone", func(t *testing.T) {
		testDecipherTx(
			t,
			&types.Transaction{
				Version:  2,
				LockTime: 0,
				TxIn:     []*types.TxIn{},
				TxOut: []*types.TxOut{
					{
						PkScript: utils.Must(txscript.NewScriptBuilder().
							AddOp(txscript.OP_RETURN).
							AddOp(txscript.OP_DATA_9).
							AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
							AddOp(txscript.OP_DATA_4).
							Script()),
					},
					{
						PkScript: utils.Must(txscript.NewScriptBuilder().
							AddOp(txscript.OP_RETURN).
							AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
							AddData(encodeLEB128VarIntsToPayload([]uint128.Uint128{
								TagMint.Uint128(),
								uint128.From64(1),
								TagMint.Uint128(),
								uint128.From64(1),
							})).
							Script()),
					},
				},
			},
			&Runestone{
				Mint: lo.ToPtr(RuneId{1, 1}),
			},
		)
	})
	t.Run("deciphering_non_empty_runestone_is_successful", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{TagBody.Uint128(), uint128.From64(1), uint128.From64(1), uint128.From64(2), uint128.From64(0)},
			&Runestone{
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("decipher_etching", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Etching: &Etching{},
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("decipher_etching_with_rune", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.From64(4),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Etching: &Etching{
					Rune: lo.ToPtr(NewRune(4)),
				},
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("terms_flag_without_etching_flag_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagTerms.Mask().Uint128(),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedFlag.Mask(),
			},
		)
	})
	t.Run("recognized_fields_without_flag_produces_cenotaph", func(t *testing.T) {
		testcase := func(integers []uint128.Uint128) {
			testDecipherInteger(
				t,
				integers,
				&Runestone{
					Cenotaph: true,
					Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
				},
			)
		}

		testcase([]uint128.Uint128{TagPremine.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagRune.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagCap.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagAmount.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagOffsetStart.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagOffsetEnd.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagHeightStart.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagHeightEnd.Uint128(), uint128.Zero})

		testcase([]uint128.Uint128{TagFlags.Uint128(), FlagEtching.Mask().Uint128(), TagCap.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagFlags.Uint128(), FlagEtching.Mask().Uint128(), TagAmount.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagFlags.Uint128(), FlagEtching.Mask().Uint128(), TagOffsetStart.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagFlags.Uint128(), FlagEtching.Mask().Uint128(), TagOffsetEnd.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagFlags.Uint128(), FlagEtching.Mask().Uint128(), TagHeightStart.Uint128(), uint128.Zero})
		testcase([]uint128.Uint128{TagFlags.Uint128(), FlagEtching.Mask().Uint128(), TagHeightEnd.Uint128(), uint128.Zero})
	})
	t.Run("decipher_etching_with_term", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128().Or(FlagTerms.Mask().Uint128()),
				TagOffsetEnd.Uint128(),
				uint128.From64(4),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Etching: &Etching{
					Terms: &Terms{
						OffsetEnd: lo.ToPtr(uint64(4)),
					},
				},
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("decipher_etching_with_amount", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128().Or(FlagTerms.Mask().Uint128()),
				TagAmount.Uint128(),
				uint128.From64(4),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Etching: &Etching{
					Terms: &Terms{
						Amount: lo.ToPtr(uint128.From64(4)),
					},
				},
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("duplicate_even_tags_produce_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.From64(4),
				TagRune.Uint128(),
				uint128.From64(5),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Etching: &Etching{
					Rune: lo.ToPtr(NewRune(4)),
				},
				Flaws: FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("duplicate_odd_tags_are_ignored", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagDivisibility.Uint128(),
				uint128.From64(4),
				TagDivisibility.Uint128(),
				uint128.From64(5),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Etching: &Etching{
					Divisibility: lo.ToPtr(uint8(4)),
				},
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("unrecognized_odd_tag_is_ignored", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagNop.Uint128(),
				uint128.From64(5),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("runestone_with_unrecognized_even_tag_is_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagCenotaph.Uint128(),
				uint128.From64(5),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("runestone_with_unrecognized_flag_is_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagCenotaph.Mask().Uint128(),
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedFlag.Mask(),
			},
		)
	})
}
