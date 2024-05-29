package nodesale

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
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

func (p *Processor) processNodesaleEvent(ctx context.Context, querier gen.Querier, transaction *types.Transaction, block *types.BlockHeader, data []byte) error {
	event := &protobuf.NodeSaleEvent{}
	err := proto.Unmarshal(data, event)
	if err != nil {
		return fmt.Errorf("Failed to parse protobuf : %w", err)
	}
	eventJson, err := protojson.Marshal(event)
	if err != nil {
		return fmt.Errorf("Failed to parse protobuf to json : %w", err)
	}

	err = querier.AddEvent(ctx, gen.AddEventParams{
		TxHash:         transaction.TxHash.String(),
		TxIndex:        int32(transaction.Index),
		Action:         int32(event.Action),
		RawMessage:     data,
		ParsedMessage:  eventJson,
		BlockTimestamp: pgtype.Timestamp{Time: block.Timestamp, Valid: true},
		BlockHash:      block.Hash.String(),
		BlockHeight:    int32(block.Height),
	})
	if err != nil {
		return fmt.Errorf("Failed to insert event : %w", err)
	}

	switch event.Action {
	case protobuf.Action_ACTION_DEPLOY:
		deploy := event.Deploy
		// add to nodesales table
		tiers := make([][]byte, len(deploy.Tiers))
		for _, tier := range deploy.Tiers {
			tierJson, err := protojson.Marshal(tier)
			if err != nil {
				return fmt.Errorf("Failed to parse tiers to json : %w", err)
			}
			tiers = append(tiers, tierJson)
		}
		err = querier.AddNodesale(ctx, gen.AddNodesaleParams{
			BlockHeight: int32(block.Height),
			TxIndex:     int32(transaction.Index),
			StartsAt: pgtype.Timestamp{
				Time:  time.Unix(int64(deploy.StartsAt), 0).UTC(),
				Valid: true,
			},
			EndsAt: pgtype.Timestamp{
				Time:  time.Unix(int64(deploy.EndsAt), 0).UTC(),
				Valid: true,
			},
			Tiers:           tiers,
			SellerPublicKey: string(deploy.SellerPublicKey),
			MaxPerAddress:   int32(deploy.MaxPerAddress),
			DeployTxHash:    transaction.TxHash.String(),
		})
		if err != nil {
			return fmt.Errorf("Failed to insert to node_sales : %w", err)
		}
	case protobuf.Action_ACTION_PURCHASE:
		// add to nodes table
	case protobuf.Action_ACTION_DELEGATE:
		// update node table
	}
	return nil
}

func (p *Processor) processTransaction(ctx context.Context, querier gen.Querier, transaction *types.Transaction, block *types.BlockHeader) error {
	for _, txIn := range transaction.TxIn {
		tokenizer, isTapScript := extractTapScript(txIn.Witness)
		if !isTapScript {
			continue
		}
		state := 0
		for tokenizer.Next() {
			switch state {
			case 0:
				if tokenizer.Opcode() == txscript.OP_0 {
					state++
				} else {
					state = 0
				}
			case 1:
				if tokenizer.Opcode() == txscript.OP_IF {
					state++
				} else {
					state = 0
				}
			case 2:
				if tokenizer.Opcode() == txscript.OP_DATA_4 &&
					bytes.Equal(tokenizer.Data(), NODESALE_MAGIC) {
					state++
				} else {
					state = 0
				}
			case 3:
				if tokenizer.Opcode() == txscript.OP_PUSHDATA1 {
					data := tokenizer.Data()
					err := p.processNodesaleEvent(ctx, querier, transaction, block, data)
					if err != nil {
						return fmt.Errorf("Failed to process protobuf event : %w", err)
					}
					return nil
				}
				state = 0
			}
		}
	}
	return nil
}

func (p *Processor) processBlock(ctx context.Context, block *types.Block) error {
	tx, err := p.repository.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create transaction : %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := p.repository.WithTx(tx)

	err = qtx.AddBlock(ctx, gen.AddBlockParams{
		BlockHeight: int32(block.Header.Height),
		BlockHash:   block.Header.Hash.String(),
		Module:      p.Name(),
	})
	if err != nil {
		return fmt.Errorf("Failed to add block %d : %w", block.Header.Height, err)
	}

	for _, transaction := range block.Transactions {
		err = p.processTransaction(ctx, qtx, transaction, &block.Header)
		if err != nil {
			logger.WarnContext(ctx, "failed to process transaction",
				slog.String("block_hash", transaction.BlockHash.String()),
				slog.Int("txIndex", int(transaction.Index)))
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction : %w", err)
	}
	return nil
}

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	for _, block := range inputs {
		if err := p.processBlock(ctx, block); err != nil {
			return fmt.Errorf("Process block %d failed. : %w", block.Header.Height, err)
		}
	}
	return nil
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
