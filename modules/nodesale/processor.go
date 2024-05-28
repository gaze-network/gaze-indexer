package nodesale

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"

	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
)

type Processor struct {
	repository          repository.Repository
	nodesaleGenesis     int32
	nodesaleGenesisHash chainhash.Hash
}

// CurrentBlock implements indexer.Processor.
func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	block, err := p.repository.Queries.GetLastProcessedBlock(ctx)
	if err != nil {
		logger.InfoContext(ctx, "Couldn't get last processed block. Start from Nodesale genesis.",
			slog.Int("currentBlock", int(p.nodesaleGenesis)))
		return types.BlockHeader{
			Hash:   p.nodesaleGenesisHash,
			Height: int64(p.nodesaleGenesis),
		}, nil
	}

	hash, err := chainhash.NewHashFromStr(block.BlockHash)
	if err != nil {
		logger.PanicContext(ctx, "Invalid hash format found in Database.")
	}
	return types.BlockHeader{
		Hash:   *hash,
		Height: int64(block.BlockHeight),
	}, nil
}

// GetIndexedBlock implements indexer.Processor.
func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	block, err := p.repository.Queries.GetBlock(ctx, int32(height))
	if err != nil {
		return types.BlockHeader{}, fmt.Errorf("Block %d not found : %w", height, err)
	}
	hash, err := chainhash.NewHashFromStr(block.BlockHash)
	if err != nil {
		logger.PanicContext(ctx, "Invalid hash format found in Database.")
	}
	return types.BlockHeader{
		Hash:   *hash,
		Height: int64(block.BlockHeight),
	}, nil
}

// Name implements indexer.Processor.
func (p *Processor) Name() string {
	return "nodesale"
}

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	panic("unimplemented")
}

// RevertData implements indexer.Processor.
func (p *Processor) RevertData(ctx context.Context, from int64) error {
	tx, err := p.repository.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create transaction : %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := p.repository.WithTx(tx)
	affected, err := qtx.RemoveEventsFromBlock(ctx, int32(from))
	if err != nil {
		return fmt.Errorf("Failed to remove events. : %w", err)
	}
	_, err = qtx.RemoveBlockFrom(ctx, int32(from))
	if err != nil {
		return fmt.Errorf("Failed to remove blocks. : %w", err)
	}
	_, err = qtx.ClearDelegate(ctx)
	if err != nil {
		return fmt.Errorf("Failed to clear delegate from nodes : %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction : %w", err)
	}
	logger.InfoContext(ctx, "Events removed",
		slog.Int("Total removed", int(affected)))
	return nil
}

// VerifyStates implements indexer.Processor.
func (p *Processor) VerifyStates(ctx context.Context) error {
	panic("unimplemented")
}

var _ indexer.Processor[*types.Block] = (*Processor)(nil)
