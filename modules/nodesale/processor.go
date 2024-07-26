package nodesale

import (
	"bytes"
	"context"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/internal/entity"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
)

type NodeSaleEvent struct {
	Transaction  *types.Transaction
	EventMessage *protobuf.NodeSaleEvent
	EventJson    []byte
	TxPubkey     *btcec.PublicKey
	RawData      []byte
}

func NewProcessor(repository datagateway.NodeSaleDataGateway,
	datasource *datasources.BitcoinNodeDatasource,
	network common.Network,
	cleanupFuncs []func(context.Context) error,
	lastBlockDefault int64,
) *Processor {
	return &Processor{
		NodeSaleDg:       repository,
		BtcClient:        datasource,
		Network:          network,
		cleanupFuncs:     cleanupFuncs,
		lastBlockDefault: lastBlockDefault,
	}
}

func (p *Processor) Shutdown(ctx context.Context) error {
	for _, cleanupFunc := range p.cleanupFuncs {
		err := cleanupFunc(ctx)
		if err != nil {
			return errors.Wrap(err, "cleanup function error")
		}
	}
	return nil
}

type Processor struct {
	NodeSaleDg       datagateway.NodeSaleDataGateway
	BtcClient        *datasources.BitcoinNodeDatasource
	Network          common.Network
	cleanupFuncs     []func(context.Context) error
	lastBlockDefault int64
}

// CurrentBlock implements indexer.Processor.
func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	block, err := p.NodeSaleDg.GetLastProcessedBlock(ctx)
	if err != nil {
		logger.InfoContext(ctx, "Couldn't get last processed block. Start from NODESALE_LAST_BLOCK_DEFAULT.",
			slogx.Int64("currentBlock", p.lastBlockDefault))
		header, err := p.BtcClient.GetBlockHeader(ctx, p.lastBlockDefault)
		if err != nil {
			return types.BlockHeader{}, errors.Wrap(err, "Cannot get default block from bitcoin node")
		}
		return types.BlockHeader{
			Hash:   header.Hash,
			Height: p.lastBlockDefault,
		}, nil
	}

	hash, err := chainhash.NewHashFromStr(block.BlockHash)
	if err != nil {
		logger.PanicContext(ctx, "Invalid hash format found in Database.")
	}
	return types.BlockHeader{
		Hash:   *hash,
		Height: block.BlockHeight,
	}, nil
}

// GetIndexedBlock implements indexer.Processor.
func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	block, err := p.NodeSaleDg.GetBlock(ctx, height)
	if err != nil {
		return types.BlockHeader{}, errors.Wrapf(err, "Block %d not found", height)
	}
	hash, err := chainhash.NewHashFromStr(block.BlockHash)
	if err != nil {
		logger.PanicContext(ctx, "Invalid hash format found in Database.")
	}
	return types.BlockHeader{
		Hash:   *hash,
		Height: block.BlockHeight,
	}, nil
}

// Name implements indexer.Processor.
func (p *Processor) Name() string {
	return "nodesale"
}

func extractNodeSaleData(witness [][]byte) (data []byte, internalPubkey *btcec.PublicKey, isNodeSale bool) {
	tokenizer, controlBlock, isTapScript := extractTapScript(witness)
	if !isTapScript {
		return []byte{}, nil, false
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
			// Any instruction > txscript.OP_16 is not push data. Note: txscript.OP_PUSHDATAX < txscript.OP_16
			if tokenizer.Opcode() <= txscript.OP_16 {
				data := tokenizer.Data()
				return data, controlBlock.InternalKey, true
			}
			state = 0
		}
	}
	return []byte{}, nil, false
}

func (p *Processor) parseTransactions(ctx context.Context, transactions []*types.Transaction) ([]NodeSaleEvent, error) {
	var events []NodeSaleEvent
	for _, t := range transactions {
		for _, txIn := range t.TxIn {
			data, txPubkey, isNodeSale := extractNodeSaleData(txIn.Witness)
			if !isNodeSale {
				continue
			}

			event := &protobuf.NodeSaleEvent{}
			err := proto.Unmarshal(data, event)
			if err != nil {
				logger.WarnContext(ctx, "Invalid Protobuf",
					slogx.String("block_hash", t.BlockHash.String()),
					slogx.Int("txIndex", int(t.Index)))
				continue
			}
			eventJson, err := protojson.Marshal(event)
			if err != nil {
				return []NodeSaleEvent{}, errors.Wrap(err, "Failed to parse protobuf to json")
			}

			events = append(events, NodeSaleEvent{
				Transaction:  t,
				EventMessage: event,
				EventJson:    eventJson,
				RawData:      data,
				TxPubkey:     txPubkey,
			})
		}
	}
	return events, nil
}

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	for _, block := range inputs {
		logger.InfoContext(ctx, "NodeSale processing a block",
			slogx.Int64("block", block.Header.Height),
			slogx.Stringer("hash", block.Header.Hash))
		// parse all event from each transaction including reading tx wallet
		events, err := p.parseTransactions(ctx, block.Transactions)
		if err != nil {
			return errors.Wrap(err, "Invalid data from bitcoin client")
		}
		// open transaction
		qtx, err := p.NodeSaleDg.BeginNodeSaleTx(ctx)
		if err != nil {
			return errors.Wrap(err, "Failed to create transaction")
		}
		defer func() {
			err = qtx.Rollback(ctx)
			if err != nil {
				logger.PanicContext(ctx, "Failed to rollback db")
			}
		}()

		// write block
		err = qtx.CreateBlock(ctx, entity.Block{
			BlockHeight: block.Header.Height,
			BlockHash:   block.Header.Hash.String(),
			Module:      p.Name(),
		})
		if err != nil {
			return errors.Wrapf(err, "Failed to add block %d", block.Header.Height)
		}
		// for each events
		for _, event := range events {
			logger.InfoContext(ctx, "NodeSale processing event",
				slogx.Uint32("txIndex", event.Transaction.Index),
				slogx.Int64("blockHeight", block.Header.Height),
				slogx.Stringer("blockhash", block.Header.Hash),
			)
			eventMessage := event.EventMessage
			switch eventMessage.Action {
			case protobuf.Action_ACTION_DEPLOY:
				err = p.ProcessDeploy(ctx, qtx, block, event)
				if err != nil {
					return errors.Wrapf(err, "Failed to deploy at block %d", block.Header.Height)
				}
			case protobuf.Action_ACTION_DELEGATE:
				err = p.ProcessDelegate(ctx, qtx, block, event)
				if err != nil {
					return errors.Wrapf(err, "Failed to delegate at block %d", block.Header.Height)
				}
			case protobuf.Action_ACTION_PURCHASE:
				err = p.ProcessPurchase(ctx, qtx, block, event)
				if err != nil {
					return errors.Wrapf(err, "Failed to purchase at block %d", block.Header.Height)
				}
			default:
				logger.DebugContext(ctx, "Invalid event ACTION", slogx.Stringer("txHash", (event.Transaction.TxHash)))
			}
		}
		// close transaction
		err = qtx.Commit(ctx)
		if err != nil {
			return errors.Wrap(err, "Failed to commit transaction")
		}
		logger.InfoContext(ctx, "NodeSale finished processing block",
			slogx.Int64("block", block.Header.Height),
			slogx.Stringer("hash", block.Header.Hash))
	}
	return nil
}

// RevertData implements indexer.Processor.
func (p *Processor) RevertData(ctx context.Context, from int64) error {
	qtx, err := p.NodeSaleDg.BeginNodeSaleTx(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to create transaction")
	}
	defer func() { err = qtx.Rollback(ctx) }()
	_, err = qtx.RemoveBlockFrom(ctx, from)
	if err != nil {
		return errors.Wrap(err, "Failed to remove blocks.")
	}

	affected, err := qtx.RemoveEventsFromBlock(ctx, from)
	if err != nil {
		return errors.Wrap(err, "Failed to remove events.")
	}
	_, err = qtx.ClearDelegate(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to clear delegate from nodes")
	}
	err = qtx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to commit transaction")
	}
	logger.InfoContext(ctx, "Events removed",
		slogx.Int64("Total removed", affected))
	return nil
}

// VerifyStates implements indexer.Processor.
func (p *Processor) VerifyStates(ctx context.Context) error {
	panic("unimplemented")
}

var _ indexer.Processor[*types.Block] = (*Processor)(nil)
