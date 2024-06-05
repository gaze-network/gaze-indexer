package nodesale

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
)

type Processor struct {
	repository          *repository.Repository
	nodesaleGenesis     int32
	nodesaleGenesisHash chainhash.Hash
	btcClient           *datasources.BitcoinNodeDatasource
	network             common.Network
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

func extractNodesaleData(witness [][]byte) ([]byte, []byte, bool) {
	tokenizer, rawScript, isTapScript := extractTapScript(witness)
	if !isTapScript {
		return []byte{}, []byte{}, false
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
				return data, rawScript, true
			}
			state = 0
		}
	}
	return []byte{}, []byte{}, false
}

type nodesaleEvent struct {
	transaction  *types.Transaction
	eventMessage *protobuf.NodeSaleEvent
	eventJson    []byte
	txAddress    btcutil.Address
	rawData      []byte
	rawScript    []byte
}

func (p *Processor) parseTransactions(ctx context.Context, transactions []*types.Transaction) ([]nodesaleEvent, error) {
	events := make([]nodesaleEvent, 1)
	for _, t := range transactions {
		for _, txIn := range t.TxIn {
			data, rawScript, isNodesale := extractNodesaleData(txIn.Witness)
			if !isNodesale {
				continue
			}

			event := &protobuf.NodeSaleEvent{}
			err := proto.Unmarshal(data, event)
			if err != nil {
				logger.WarnContext(ctx, "Invalid Protobuf",
					slog.String("block_hash", t.BlockHash.String()),
					slog.Int("txIndex", int(t.Index)))
				continue
			}
			eventJson, err := protojson.Marshal(event)
			if err != nil {
				return []nodesaleEvent{}, fmt.Errorf("Failed to parse protobuf to json : %w", err)
			}

			outIndex := txIn.PreviousOutIndex
			outHash := txIn.PreviousOutTxHash
			result, err := p.btcClient.GetTransactionByHash(ctx, outHash)
			if err != nil {
				return []nodesaleEvent{}, fmt.Errorf("Failed to Get Bitcoin transaction : %w", err)
			}
			pkScript := result.TxOut[outIndex].PkScript
			_, addresses, _, err := txscript.ExtractPkScriptAddrs(pkScript, p.network.ChainParams())
			if err != nil {
				return []nodesaleEvent{}, fmt.Errorf("Failed to Get Bitcoin address : %w", err)
			}
			if len(addresses) != 1 {
				return []nodesaleEvent{}, fmt.Errorf("Multiple addresses detected.")
			}

			events = append(events, nodesaleEvent{
				transaction:  t,
				eventMessage: event,
				eventJson:    eventJson,
				txAddress:    addresses[0],
				rawData:      data,
				rawScript:    rawScript,
			})
		}
	}
	return events, nil
}

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	for _, block := range inputs {
		// parse all event from each transaction including reading tx wallet
		events, err := p.parseTransactions(ctx, block.Transactions)
		if err != nil {
			return fmt.Errorf("Invalid data from bitcoin client : %w", err)
		}

		// open transaction
		tx, err := p.repository.Db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("Failed to create transaction : %w", err)
		}
		defer tx.Rollback(ctx)
		qtx := p.repository.WithTx(tx)

		// write block
		err = qtx.AddBlock(ctx, gen.AddBlockParams{
			BlockHeight: int32(block.Header.Height),
			BlockHash:   block.Header.Hash.String(),
			Module:      p.Name(),
		})
		if err != nil {
			return fmt.Errorf("Failed to add block %d : %w", block.Header.Height, err)
		}
		// for each events
		for _, event := range events {
			eventMessage := event.eventMessage
			switch eventMessage.Action {
			case protobuf.Action_ACTION_DEPLOY:
				err = p.processDeploy(ctx, qtx, block, event)
				if err != nil {
					return fmt.Errorf("Failed to deploy at block %d : %w", block.Header.Height, err)
				}
			case protobuf.Action_ACTION_DELEGATE:
				err = p.processDelegate(ctx, qtx, block, event)
				if err != nil {
					return fmt.Errorf("Failed to delegate at block %d : %w", block.Header.Height, err)
				}
			case protobuf.Action_ACTION_PURCHASE:
				err = p.processPurchase(ctx, qtx, block, event)
				if err != nil {
					return fmt.Errorf("Failed to purchase at block %d : %w", block.Header.Height, err)
				}
			}
		}
		// close transaction
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("Failed to commit transaction : %w", err)
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
	_, err = qtx.RemoveBlockFrom(ctx, int32(from))
	if err != nil {
		return fmt.Errorf("Failed to remove blocks. : %w", err)
	}

	affected, err := qtx.RemoveEventsFromBlock(ctx, int32(from))
	if err != nil {
		return fmt.Errorf("Failed to remove events. : %w", err)
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

func (p *Processor) Shutdown(ctx context.Context) error {
	panic("unimplemented")
}

var _ indexer.Processor[*types.Block] = (*Processor)(nil)
