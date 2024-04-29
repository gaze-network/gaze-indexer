package bitcoin

import (
	"fmt"
	"testing"

	"github.com/gaze-network/indexer-network/core/types"
	"github.com/stretchr/testify/assert"
)

func TestDuplicateCoinbaseTxHashHandling(t *testing.T) {
	processor := Processor{}
	generator := func() []*types.Block {
		return []*types.Block{
			{
				Header: types.BlockHeader{Height: 91842, Version: 1},
				Transactions: []*types.Transaction{
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
				},
			},
			{
				Header: types.BlockHeader{Height: 91880, Version: 1},
				Transactions: []*types.Transaction{
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
				},
			},
		}
	}

	t.Run("all_duplicated_txs", func(t *testing.T) {
		blocks := generator()
		processor.removeDuplicateCoinbaseTxInputsOutputs(blocks)

		assert.Len(t, blocks, 2, "should not remove any blocks")
		for _, block := range blocks {
			assert.Len(t, block.Transactions, 2, "should not remove any transactions")
			assert.Len(t, block.Transactions[0].TxIn, 0, "should remove tx inputs from coinbase transaction")
			assert.Len(t, block.Transactions[0].TxOut, 0, "should remove tx outputs from coinbase transaction")
		}
	})

	t.Run("not_duplicated_txs", func(t *testing.T) {
		blocks := []*types.Block{
			{
				Header: types.BlockHeader{Height: 91812, Version: 1},
				Transactions: []*types.Transaction{
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
				},
			},
			{
				Header: types.BlockHeader{Height: 91722, Version: 1},
				Transactions: []*types.Transaction{
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
				},
			},
		}
		processor.removeDuplicateCoinbaseTxInputsOutputs(blocks)

		assert.Len(t, blocks, 2, "should not remove any blocks")
		for _, block := range blocks {
			assert.Len(t, block.Transactions, 2, "should not remove any transactions")
			assert.Len(t, block.Transactions[0].TxIn, 4, "should not remove tx inputs from coinbase transaction")
			assert.Len(t, block.Transactions[0].TxOut, 4, "should not remove tx outputs from coinbase transaction")
		}
	})

	t.Run("mixed", func(t *testing.T) {
		blocks := []*types.Block{
			{
				Header: types.BlockHeader{Height: 91812, Version: 1},
				Transactions: []*types.Transaction{
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
					{
						TxIn:  []*types.TxIn{{}, {}, {}, {}},
						TxOut: []*types.TxOut{{}, {}, {}, {}},
					},
				},
			},
		}
		blocks = append(blocks, generator()...)
		blocks = append(blocks, &types.Block{
			Header: types.BlockHeader{Height: 91722, Version: 1},
			Transactions: []*types.Transaction{
				{
					TxIn:  []*types.TxIn{{}, {}, {}, {}},
					TxOut: []*types.TxOut{{}, {}, {}, {}},
				},
				{
					TxIn:  []*types.TxIn{{}, {}, {}, {}},
					TxOut: []*types.TxOut{{}, {}, {}, {}},
				},
			},
		})
		processor.removeDuplicateCoinbaseTxInputsOutputs(blocks)

		assert.Len(t, blocks, 4, "should not remove any blocks")

		// only 2nd and 3rd blocks should be modified
		for i, block := range blocks {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				if i == 1 || i == 2 {
					assert.Len(t, block.Transactions, 2, "should not remove any transactions")
					assert.Len(t, block.Transactions[0].TxIn, 0, "should remove tx inputs from coinbase transaction")
					assert.Len(t, block.Transactions[0].TxOut, 0, "should remove tx outputs from coinbase transaction")
				} else {
					assert.Len(t, block.Transactions, 2, "should not remove any transactions")
					assert.Lenf(t, block.Transactions[0].TxIn, 4, "should not remove tx inputs from coinbase transaction")
					assert.Len(t, block.Transactions[0].TxOut, 4, "should not remove tx outputs from coinbase transaction")
				}
			})
		}
	})
}
