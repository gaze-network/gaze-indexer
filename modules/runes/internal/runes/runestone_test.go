package runes

import (
	"fmt"
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
	testDecipherTx := func(name string, tx *types.Transaction, expected *Runestone) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runestone, err := DecipherRunestone(tx)
			assert.NoError(t, err)
			assert.Equal(t, expected, runestone)
		})
	}

	testDecipherInteger := func(name string, integers []uint128.Uint128, expected *Runestone) {
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
		testDecipherTx(name, tx, expected)
	}

	testDecipherPkScript := func(name string, pkScript []byte, expected *Runestone) {
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
		testDecipherTx(name, tx, expected)
	}

	testDecipherPkScript(
		"decipher_returns_none_if_first_opcode_is_malformed",
		utils.Must(txscript.NewScriptBuilder().AddOp(txscript.OP_DATA_4).Script()),
		nil,
	)
	testDecipherTx(
		"deciphering_transaction_with_non_op_return_output_returns_none",
		&types.Transaction{
			Version:  2,
			LockTime: 0,
			TxIn:     []*types.TxIn{},
			TxOut:    []*types.TxOut{},
		},
		nil,
	)
	testDecipherPkScript(
		"deciphering_transaction_with_bare_op_return_returns_none",
		utils.Must(txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).Script()),
		nil,
	)
	testDecipherPkScript(
		"deciphering_transaction_with_non_matching_op_return_returns_none",
		utils.Must(txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddOp(txscript.OP_1).Script()),
		nil,
	)
	testDecipherPkScript(
		"deciphering_valid_runestone_with_invalid_script_postfix_returns_invalid_payload",
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
	testDecipherPkScript(
		"deciphering_runestone_with_truncated_varint_is_cenotaph",
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
	testDecipherPkScript(
		"outputs_with_non_pushdata_opcodes_are_cenotaph_1",
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
	testDecipherPkScript(
		"outputs_with_non_pushdata_opcodes_are_cenotaph_2",
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
	testDecipherPkScript(
		"pushnum_opcodes_in_runestone_produce_cenotaph",
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
	testDecipherPkScript(
		"deciphering_empty_runestone_is_successful",
		utils.Must(txscript.NewScriptBuilder().
			AddOp(txscript.OP_RETURN).
			AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
			Script()),
		&Runestone{},
	)
	testDecipherPkScript(
		"invalid_input_scripts_are_skipped_when_searching_for_runestone",
		utils.Must(txscript.NewScriptBuilder().
			AddOp(txscript.OP_RETURN).
			AddOp(RUNESTONE_PAYLOAD_MAGIC_NUMBER).
			Script()),
		&Runestone{},
	)
	testDecipherTx(
		"invalid_input_scripts_are_skipped_when_searching_for_runestone",
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

	testDecipherInteger(
		"deciphering_non_empty_runestone_is_successful",
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
	testDecipherInteger(
		"decipher_etching",
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
	testDecipherInteger(
		"decipher_etching_with_rune",
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
	testDecipherInteger(
		"terms_flag_without_etching_flag_produces_cenotaph",
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
	// test recognized_fields_without_flag_produces_cenotaph
	{
		testcase := func(integers []uint128.Uint128) {
			testDecipherInteger(
				fmt.Sprintf("recognized_fields_without_flag_produces_cenotaph_%v", integers),
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
	}
	testDecipherInteger(
		"decipher_etching_with_term",
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
	testDecipherInteger(
		"decipher_etching_with_amount",
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
	testDecipherInteger(
		"duplicate_even_tags_produce_cenotaph",
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
	testDecipherInteger(
		"duplicate_odd_tags_are_ignored",
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
	testDecipherInteger(
		"unrecognized_odd_tag_is_ignored",
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
	testDecipherInteger(
		"runestone_with_unrecognized_even_tag_is_cenotaph",
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
	testDecipherInteger(
		"runestone_with_unrecognized_flag_is_cenotaph",
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
}
