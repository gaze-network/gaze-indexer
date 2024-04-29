package bitcoin

import (
	"cmp"
	"context"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
)

// process is a processing rules for the given blocks before inserting to the database
//
// this function will modify the given data directly.
func (p *Processor) process(ctx context.Context, blocks []*types.Block) ([]*types.Block, error) {
	if len(blocks) == 0 {
		return blocks, nil
	}

	// Sort ASC by block height
	slices.SortFunc(blocks, func(t1, t2 *types.Block) int {
		return cmp.Compare(t1.Header.Height, t2.Header.Height)
	})

	if !p.isContinueFromLatestIndexedBlock(ctx, blocks[0]) {
		return nil, errors.New("given blocks are not continue from the latest indexed block")
	}

	if !p.isBlocksSequential(blocks) {
		return nil, errors.New("given blocks are not in sequence")
	}

	p.removeDuplicateCoinbaseTxInputsOutputs(blocks)

	return blocks, nil
}

// check if the given blocks are continue from the latest indexed block
// to prevent inserting out-of-order blocks or duplicate blocks
func (p *Processor) isBlocksSequential(blocks []*types.Block) bool {
	if len(blocks) == 0 {
		return true
	}

	for i, block := range blocks {
		if i == 0 {
			continue
		}

		if block.Header.Height != blocks[i-1].Header.Height+1 {
			return false
		}
	}

	return true
}

// check if the given blocks are continue from the latest indexed block
// to prevent inserting out-of-order blocks or duplicate blocks
func (p *Processor) isContinueFromLatestIndexedBlock(ctx context.Context, block *types.Block) bool {
	latestBlock, err := p.CurrentBlock(ctx)
	if err != nil {
		return false
	}

	return block.Header.Height == latestBlock.Height+1
}

// there 2 coinbase transaction that are duplicated in the blocks 91842 and 91880.
// if the given block version is v1 and height is `91842` or `91880`,
// then remove transaction inputs/outputs to prevent duplicate txin/txout error when inserting to the database.
//
// Theses duplicated coinbase transactions are having the same transaction input/output and
// utxo from these 2 duplicated coinbase txs can redeem only once. so, it's safe to remove them and can
// use inputs/outputs from the previous block.
//
// Duplicate Coinbase Transactions:
//   - `454279874213763724535987336644243549a273058910332236515429488599` in blocks 91812, 91842
//   - `e3bf3d07d4b0375638d5f1db5255fe07ba2c4cb067cd81b84ee974b6585fb468` in blocks 91722, 91880
//
// This function will modify the given data directly.
func (p *Processor) removeDuplicateCoinbaseTxInputsOutputs(blocks []*types.Block) {
	for _, block := range blocks {
		header := block.Header
		if header.Version == 1 && (header.Height == 91842 || header.Height == 91880) {
			// remove transaction inputs/outputs from coinbase transaction (first transaction)
			block.Transactions[0].TxIn = nil
			block.Transactions[0].TxOut = nil
		}
	}
}
