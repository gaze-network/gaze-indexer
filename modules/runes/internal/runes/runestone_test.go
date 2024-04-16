package runes

import (
	"math"
	"slices"
	"testing"
	"unicode/utf8"

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
		t.Helper()
		runestone, err := DecipherRunestone(tx)
		assert.NoError(t, err)
		assert.Equal(t, expected, runestone)
	}

	testDecipherInteger := func(t *testing.T, integers []uint128.Uint128, expected *Runestone) {
		t.Helper()
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
		t.Helper()
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
	t.Run("runestone_with_edict_id_with_zero_block_and_nonzero_tx_is_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(0),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictRuneId.Mask(),
			},
		)
	})
	t.Run("runestone_with_overflowing_edict_id_delta_is_cenotaph_1", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(0),
				uint128.From64(0),
				uint128.From64(0),
				uint128.From64(math.MaxUint64),
				uint128.From64(0),
				uint128.From64(0),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictRuneId.Mask(),
			},
		)
	})
	t.Run("runestone_with_overflowing_edict_id_delta_is_cenotaph_2", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(0),
				uint128.From64(0),
				uint128.From64(0),
				uint128.From64(math.MaxUint64),
				uint128.From64(0),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictRuneId.Mask(),
			},
		)
	})
	t.Run("runestone_with_output_over_max_is_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(2),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictOutput.Mask(),
			},
		)
	})
	t.Run("tag_with_no_value_is_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				uint128.From64(1),
				TagFlags.Uint128(),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagTruncatedField.Mask(),
			},
		)
	})
	t.Run("trailing_integers_in_body_is_cenotaph", func(t *testing.T) {
		integers := []uint128.Uint128{
			TagBody.Uint128(),
			uint128.From64(1),
			uint128.From64(1),
			uint128.From64(2),
			uint128.From64(0),
		}
		for i := 0; i < 4; i++ {
			if i == 0 {
				testDecipherInteger(t, integers, &Runestone{
					Edicts: []Edict{
						{
							Id:     RuneId{1, 1},
							Amount: uint128.From64(2),
							Output: 0,
						},
					},
				})
			} else {
				testDecipherInteger(t, integers, &Runestone{
					Cenotaph: true,
					Flaws:    FlawFlagTrailingIntegers.Mask(),
				})
			}
			integers = append(integers, uint128.Zero)
		}
	})
	t.Run("decipher_etching_with_divisibility", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
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
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
				Etching: &Etching{
					Rune:         lo.ToPtr(NewRune(4)),
					Divisibility: lo.ToPtr(uint8(5)),
				},
			},
		)
	})
	t.Run("divisibility_above_max_is_ignored", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.From64(4),
				TagDivisibility.Uint128(),
				uint128.From64(uint64(maxDivisibility + 1)),
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
				Etching: &Etching{
					Rune: lo.ToPtr(NewRune(4)),
				},
			},
		)
	})
	t.Run("symbol_above_max_is_ignored", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagSymbol.Uint128(),
				uint128.From64(utf8.MaxRune + 1),
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
				Etching: &Etching{},
			},
		)
	})
	t.Run("decipher_etching_with_symbol", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.From64(4),
				TagSymbol.Uint128(),
				uint128.From64('a'),
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
				Etching: &Etching{
					Rune:   lo.ToPtr(NewRune(4)),
					Symbol: lo.ToPtr('a'),
				},
			},
		)
	})
	t.Run("decipher_etching_with_all_etching_tags", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Or(FlagTerms.Mask()).Or(FlagTurbo.Mask()).Uint128(),
				TagRune.Uint128(),
				uint128.From64(4),
				TagDivisibility.Uint128(),
				uint128.From64(1),
				TagSpacers.Uint128(),
				uint128.From64(5),
				TagSymbol.Uint128(),
				uint128.From64('a'),
				TagOffsetEnd.Uint128(),
				uint128.From64(2),
				TagAmount.Uint128(),
				uint128.From64(3),
				TagPremine.Uint128(),
				uint128.From64(8),
				TagCap.Uint128(),
				uint128.From64(9),
				TagPointer.Uint128(),
				uint128.From64(0),
				TagMint.Uint128(),
				uint128.From64(1),
				TagMint.Uint128(),
				uint128.From64(1),
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
				Etching: &Etching{
					Divisibility: lo.ToPtr(uint8(1)),
					Premine:      lo.ToPtr(uint128.From64(8)),
					Rune:         lo.ToPtr(NewRune(4)),
					Spacers:      lo.ToPtr(uint32(5)),
					Symbol:       lo.ToPtr('a'),
					Terms: &Terms{
						Amount:    lo.ToPtr(uint128.From64(3)),
						Cap:       lo.ToPtr(uint128.From64(9)),
						OffsetEnd: lo.ToPtr(uint64(2)),
					},
					Turbo: true,
				},
				Pointer: lo.ToPtr(uint64(0)),
				Mint:    lo.ToPtr(RuneId{1, 1}),
			},
		)
	})
	t.Run("recognized_even_etching_fields_produce_cenotaph_if_etching_flag_is_not_set", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagRune.Uint128(),
				uint128.From64(4),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("decipher_etching_with_divisibility_and_symbol", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.From64(4),
				TagDivisibility.Uint128(),
				uint128.From64(1),
				TagSymbol.Uint128(),
				uint128.From64('a'),
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
				Etching: &Etching{
					Rune:         lo.ToPtr(NewRune(4)),
					Divisibility: lo.ToPtr(uint8(1)),
					Symbol:       lo.ToPtr('a'),
				},
			},
		)
	})
	t.Run("tag_values_are_not_parsed_as_tags", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagDivisibility.Uint128(),
				TagBody.Uint128(),
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
				Etching: &Etching{
					Divisibility: lo.ToPtr(uint8(0)),
				},
			},
		)
	})
	t.Run("runestone_may_contain_multiple_edicts", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
				uint128.From64(0),
				uint128.From64(3),
				uint128.From64(5),
				uint128.From64(0),
			},
			&Runestone{
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
					{
						Id:     RuneId{1, 4},
						Amount: uint128.From64(5),
						Output: 0,
					},
				},
			},
		)
	})
	t.Run("runestones_with_invalid_rune_id_blocks_are_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
				uint128.Max,
				uint128.From64(1),
				uint128.From64(0),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictRuneId.Mask(),
			},
		)
	})
	t.Run("runestones_with_invalid_rune_id_txs_are_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(2),
				uint128.From64(0),
				uint128.From64(1),
				uint128.Max,
				uint128.From64(0),
				uint128.From64(0),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictRuneId.Mask(),
			},
		)
	})
	t.Run("payload_pushes_are_concatenated", func(t *testing.T) {
		// cannot use txscript.ScriptBuilder because ScriptBuilder.AddData transforms data with low value into small integer opcodes
		pkScript := []byte{
			txscript.OP_RETURN,
			RUNESTONE_PAYLOAD_MAGIC_NUMBER,
		}
		addData := func(data []byte) {
			pkScript = append(pkScript, txscript.OP_DATA_1-1+byte(len(data)))
			pkScript = append(pkScript, data...)
		}
		addData(leb128.EncodeUint128(TagFlags.Uint128()))
		addData(leb128.EncodeUint128(FlagEtching.Mask().Uint128()))
		addData(leb128.EncodeUint128(TagDivisibility.Uint128()))
		addData(leb128.EncodeUint128(uint128.From64(5)))
		addData(leb128.EncodeUint128(TagBody.Uint128()))
		addData(leb128.EncodeUint128(uint128.From64(1)))
		addData(leb128.EncodeUint128(uint128.From64(1)))
		addData(leb128.EncodeUint128(uint128.From64(2)))
		addData(leb128.EncodeUint128(uint128.From64(0)))
		testDecipherPkScript(
			t,
			pkScript,
			&Runestone{
				Edicts: []Edict{
					{
						Id:     RuneId{1, 1},
						Amount: uint128.From64(2),
						Output: 0,
					},
				},
				Etching: &Etching{
					Divisibility: lo.ToPtr(uint8(5)),
				},
			},
		)
	})
	t.Run("runestone_size", func(t *testing.T) {
		testcase := func(edicts []Edict, etching *Etching, expectedSize int) {
			bytes, err := Runestone{
				Edicts:  edicts,
				Etching: etching,
			}.Encipher()
			assert.NoError(t, err)
			assert.Equal(t, expectedSize, len(bytes))
		}

		testcase(nil, nil, 2)
		testcase(
			nil,
			&Etching{
				Divisibility: lo.ToPtr(maxDivisibility),
				Rune:         lo.ToPtr(NewRune(0)),
			},
			9,
		)
		testcase(
			nil,
			&Etching{
				Divisibility: lo.ToPtr(maxDivisibility),
				Premine:      lo.ToPtr(uint128.From64(math.MaxUint64)),
				Rune:         lo.ToPtr(NewRuneFromUint128(uint128.Max)),
				Spacers:      lo.ToPtr(maxSpacers),
				Symbol:       lo.ToPtr(utf8.MaxRune),
				Terms: &Terms{
					Amount:      lo.ToPtr(uint128.From64(math.MaxUint64)),
					Cap:         lo.ToPtr(uint128.From64(math.MaxUint32)),
					HeightStart: lo.ToPtr(uint64(math.MaxUint32)),
					HeightEnd:   lo.ToPtr(uint64(math.MaxUint32)),
					OffsetStart: lo.ToPtr(uint64(math.MaxUint32)),
					OffsetEnd:   lo.ToPtr(uint64(math.MaxUint32)),
				},
				Turbo: true,
			},
			89,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{0, 0},
					Amount: uint128.From64(0),
					Output: 0,
				},
			},
			&Etching{
				Divisibility: lo.ToPtr(maxDivisibility),
				Rune:         lo.ToPtr(NewRuneFromUint128(uint128.Max)),
			},
			32,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{0, 0},
					Amount: uint128.Max,
					Output: 0,
				},
			},
			&Etching{
				Divisibility: lo.ToPtr(maxDivisibility),
				Rune:         lo.ToPtr(NewRuneFromUint128(uint128.Max)),
			},
			50,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(0),
					Output: 0,
				},
			},
			nil,
			14,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.Max,
					Output: 0,
				},
			},
			nil,
			32,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.Max,
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.Max,
					Output: 0,
				},
			},
			nil,
			54,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.Max,
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.Max,
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.Max,
					Output: 0,
				},
			},
			nil,
			76,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
			},
			nil,
			62,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
			},
			nil,
			75,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{BlockHeight: 0, TxIndex: math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{0, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{0, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{0, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
				{
					Id:     RuneId{0, math.MaxUint32},
					Amount: uint128.From64(math.MaxUint64),
					Output: 0,
				},
			},
			nil,
			73,
		)
		testcase(
			[]Edict{
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: utils.Must(uint128.FromString("1000000000000000000")),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: utils.Must(uint128.FromString("1000000000000000000")),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: utils.Must(uint128.FromString("1000000000000000000")),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: utils.Must(uint128.FromString("1000000000000000000")),
					Output: 0,
				},
				{
					Id:     RuneId{1_000_000, math.MaxUint32},
					Amount: utils.Must(uint128.FromString("1000000000000000000")),
					Output: 0,
				},
			},
			nil,
			70,
		)
	})
	t.Run("etching_with_term_greater_than_maximum_is_still_an_etching", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagOffsetEnd.Uint128(),
				uint128.From64(math.MaxUint64).Add64(1),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("encipher", func(t *testing.T) {
		testcase := func(runestone Runestone, expected []uint128.Uint128) {
			pkScript, err := runestone.Encipher()
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

			payload, flaws := runestonePayloadFromTx(tx)
			assert.NoError(t, err)
			assert.Equal(t, Flaws(0), flaws)

			integers, err := decodeLEB128VarIntsFromPayload(payload)
			assert.NoError(t, err)
			assert.Equal(t, expected, integers)

			slices.SortFunc(runestone.Edicts, func(i, j Edict) int {
				return i.Id.Cmp(j.Id)
			})
			decipheredRunestone, err := DecipherRunestone(tx)
			assert.NoError(t, err)
			assert.Equal(t, runestone, *decipheredRunestone)
		}

		testcase(Runestone{}, []uint128.Uint128{})
		testcase(
			Runestone{
				Edicts: []Edict{
					{
						Id:     RuneId{2, 3},
						Amount: uint128.From64(1),
						Output: 0,
					},
					{
						Id:     RuneId{5, 6},
						Amount: uint128.From64(4),
						Output: 1,
					},
				},
				Etching: &Etching{
					Divisibility: lo.ToPtr(uint8(7)),
					Premine:      lo.ToPtr(uint128.From64(8)),
					Rune:         lo.ToPtr(NewRune(9)),
					Spacers:      lo.ToPtr(uint32(10)),
					Symbol:       lo.ToPtr('@'),
					Terms: &Terms{
						Amount:      lo.ToPtr(uint128.From64(14)),
						Cap:         lo.ToPtr(uint128.From64(11)),
						HeightStart: lo.ToPtr(uint64(12)),
						HeightEnd:   lo.ToPtr(uint64(13)),
						OffsetStart: lo.ToPtr(uint64(15)),
						OffsetEnd:   lo.ToPtr(uint64(16)),
					},
					Turbo: true,
				},
				Mint:    lo.ToPtr(RuneId{17, 18}),
				Pointer: lo.ToPtr(uint64(0)),
			},
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Or(FlagTerms.Mask()).Or(FlagTurbo.Mask()).Uint128(),
				TagRune.Uint128(),
				uint128.From64(9),
				TagDivisibility.Uint128(),
				uint128.From64(7),
				TagSpacers.Uint128(),
				uint128.From64(10),
				TagSymbol.Uint128(),
				uint128.From64('@'),
				TagPremine.Uint128(),
				uint128.From64(8),
				TagAmount.Uint128(),
				uint128.From64(14),
				TagCap.Uint128(),
				uint128.From64(11),
				TagHeightStart.Uint128(),
				uint128.From64(12),
				TagHeightEnd.Uint128(),
				uint128.From64(13),
				TagOffsetStart.Uint128(),
				uint128.From64(15),
				TagOffsetEnd.Uint128(),
				uint128.From64(16),
				TagMint.Uint128(),
				uint128.From64(17),
				TagMint.Uint128(),
				uint128.From64(18),
				TagPointer.Uint128(),
				uint128.From64(0),
				TagBody.Uint128(),
				uint128.From64(2),
				uint128.From64(3),
				uint128.From64(1),
				uint128.From64(0),
				uint128.From64(3),
				uint128.From64(6),
				uint128.From64(4),
				uint128.From64(1),
			},
		)
		testcase(
			Runestone{
				Etching: &Etching{},
			},
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
			},
		)
	})
	t.Run("runestone_payload_is_chunked", func(t *testing.T) {
		checkScriptInstructionCount := func(pkScript []byte, expectedCount int) {
			tokenizer := txscript.MakeScriptTokenizer(0, pkScript)
			actualCount := 0
			for tokenizer.Next() {
				actualCount++
			}
			assert.Equal(t, expectedCount, actualCount)
		}

		edicts := make([]Edict, 0)
		for i := 0; i < 129; i++ {
			edicts = append(edicts, Edict{
				Id:     RuneId{},
				Amount: uint128.From64(0),
				Output: 0,
			})
		}
		pkScript, err := Runestone{
			Edicts: edicts,
		}.Encipher()
		assert.NoError(t, err)
		checkScriptInstructionCount(pkScript, 3)

		edicts = make([]Edict, 0)
		for i := 0; i < 130; i++ {
			edicts = append(edicts, Edict{
				Id:     RuneId{},
				Amount: uint128.From64(0),
				Output: 0,
			})
		}
		pkScript, err = Runestone{
			Edicts: edicts,
		}.Encipher()
		assert.NoError(t, err)
		checkScriptInstructionCount(pkScript, 4)
	})
	t.Run("edict_output_greater_than_32_max_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagBody.Uint128(),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(1),
				uint128.From64(math.MaxUint32).Add64(1),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagEdictOutput.Mask(),
			},
		)
	})
	t.Run("partial_mint_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagMint.Uint128(),
				uint128.From64(1),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("invalid_mint_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagMint.Uint128(),
				uint128.From64(0),
				TagMint.Uint128(),
				uint128.From64(1),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("invalid_deadline_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagOffsetEnd.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("invalid_default_output_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagPointer.Uint128(),
				uint128.From64(1),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagPointer.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("invalid_divisibility_does_not_produce_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagDivisibility.Uint128(),
				uint128.Max,
			},
			&Runestone{},
		)
	})
	t.Run("min_and_max_runes_are_not_cenotaphs", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.From64(0),
			},
			&Runestone{
				Etching: &Etching{
					Rune: lo.ToPtr(NewRune(0)),
				},
			},
		)
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Uint128(),
				TagRune.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Etching: &Etching{
					Rune: lo.ToPtr(NewRuneFromUint128(uint128.Max)),
				},
			},
		)
	})
	t.Run("invalid_spacers_does_not_produce_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagSpacers.Uint128(),
				uint128.Max,
			},
			&Runestone{},
		)
	})
	t.Run("invalid_symbol_does_not_produce_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagSymbol.Uint128(),
				uint128.Max,
			},
			&Runestone{},
		)
	})
	t.Run("invalid_term_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagOffsetEnd.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagUnrecognizedEvenTag.Mask(),
			},
		)
	})
	t.Run("invalid_supply_produces_cenotaph", func(t *testing.T) {
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Or(FlagTerms.Mask()).Uint128(),
				TagCap.Uint128(),
				uint128.From64(1),
				TagAmount.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Etching: &Etching{
					Terms: &Terms{
						Amount: lo.ToPtr(uint128.Max),
						Cap:    lo.ToPtr(uint128.From64(1)),
					},
				},
			},
		)
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Or(FlagTerms.Mask()).Uint128(),
				TagCap.Uint128(),
				uint128.From64(2),
				TagAmount.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagSupplyOverflow.Mask(),
			},
		)
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Or(FlagTerms.Mask()).Uint128(),
				TagCap.Uint128(),
				uint128.From64(2),
				TagAmount.Uint128(),
				uint128.Max.Div64(2).Add64(1),
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagSupplyOverflow.Mask(),
			},
		)
		testDecipherInteger(
			t,
			[]uint128.Uint128{
				TagFlags.Uint128(),
				FlagEtching.Mask().Or(FlagTerms.Mask()).Uint128(),
				TagPremine.Uint128(),
				uint128.From64(1),
				TagCap.Uint128(),
				uint128.From64(1),
				TagAmount.Uint128(),
				uint128.Max,
			},
			&Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagSupplyOverflow.Mask(),
			},
		)
	})
	t.Run("all_pushdata_opcodes_are_valid", func(t *testing.T) {
		// PushData opcodes include (per ord's spec):
		// 1. OP_0
		// 2. OP_DATA_1 - OP_DATA_76
		// 3. OP_PUSHDATA1, OP_PUSHDATA2, OP_PUSHDATA4
		for i := 0; i < 79; i++ {
			pkScript := make([]byte, 0)

			pkScript = append(pkScript, txscript.OP_RETURN)
			pkScript = append(pkScript, RUNESTONE_PAYLOAD_MAGIC_NUMBER)
			pkScript = append(pkScript, byte(i))

			if i <= 75 {
				for j := 0; j < i; j++ {
					pkScript = append(pkScript, lo.Ternary(j%2 == 0, byte(1), byte(0)))
				}
				if i%2 == 1 {
					pkScript = append(pkScript, byte(1), byte(1))
				}
			} else if i == 76 {
				pkScript = append(pkScript, byte(0))
			} else if i == 77 {
				pkScript = append(pkScript, byte(0), byte(0))
			} else {
				pkScript = append(pkScript, byte(0), byte(0), byte(0), byte(0))
			}

			testDecipherPkScript(t, pkScript, &Runestone{})
		}
	})
	t.Run("all_non_pushdata_opcodes_are_invalid", func(t *testing.T) {
		for i := 79; i <= math.MaxUint8; i++ {
			pkScript := []byte{
				txscript.OP_RETURN,
				RUNESTONE_PAYLOAD_MAGIC_NUMBER,
				byte(i),
			}
			testDecipherPkScript(t, pkScript, &Runestone{
				Cenotaph: true,
				Flaws:    FlawFlagOpCode.Mask(),
			})
		}
	})
}
