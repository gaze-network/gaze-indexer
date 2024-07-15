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

type Processor struct {
	datagateway  datagateway.NodesaleDataGateway
	btcClient    *datasources.BitcoinNodeDatasource
	network      common.Network
	cleanupFuncs []func(context.Context) error
}

// CurrentBlock implements indexer.Processor.
func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	block, err := p.datagateway.GetLastProcessedBlock(ctx)
	if err != nil {
		logger.InfoContext(ctx, "Couldn't get last processed block. Start from NODESALE_LAST_BLOCK_DEFAULT.",
			slogx.Int("currentBlock", NODESALE_LASTBLOCK_DEFAULT))
		header, err := p.btcClient.GetBlockHeader(ctx, NODESALE_LASTBLOCK_DEFAULT)
		if err != nil {
			return types.BlockHeader{}, errors.Wrap(err, "Cannot get default block from bitcoin node")
		}
		return types.BlockHeader{
			Hash:   header.Hash,
			Height: NODESALE_LASTBLOCK_DEFAULT,
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
	block, err := p.datagateway.GetBlock(ctx, height)
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

func extractNodesaleData(witness [][]byte) (data []byte, internalPubkey *btcec.PublicKey, isNodesale bool) {
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
			if tokenizer.Opcode() == txscript.OP_PUSHDATA1 ||
				tokenizer.Opcode() == txscript.OP_PUSHDATA2 ||
				tokenizer.Opcode() == txscript.OP_PUSHDATA4 {
				data := tokenizer.Data()
				return data, controlBlock.InternalKey, true
			}
			state = 0
		}
	}
	return []byte{}, nil, false
}

type nodesaleEvent struct {
	transaction  *types.Transaction
	eventMessage *protobuf.NodeSaleEvent
	eventJson    []byte
	txPubkey     *btcec.PublicKey
	rawData      []byte
}

func (p *Processor) parseTransactions(ctx context.Context, transactions []*types.Transaction) ([]nodesaleEvent, error) {
	var events []nodesaleEvent
	for _, t := range transactions {
		for _, txIn := range t.TxIn {
			data, txPubkey, isNodesale := extractNodesaleData(txIn.Witness)
			if !isNodesale {
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
				return []nodesaleEvent{}, errors.Wrap(err, "Failed to parse protobuf to json")
			}

			events = append(events, nodesaleEvent{
				transaction:  t,
				eventMessage: event,
				eventJson:    eventJson,
				rawData:      data,
				txPubkey:     txPubkey,
			})
		}
	}
	return events, nil
}

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	for _, block := range inputs {
		logger.InfoContext(ctx, "Nodesale processing a block",
			slogx.Int64("block", block.Header.Height),
			slogx.Stringer("hash", block.Header.Hash))
		// parse all event from each transaction including reading tx wallet
		events, err := p.parseTransactions(ctx, block.Transactions)
		if err != nil {
			return errors.Wrap(err, "Invalid data from bitcoin client")
		}
		// open transaction
		qtx, err := p.datagateway.BeginNodesaleTx(ctx)
		if err != nil {
			return errors.Wrap(err, "Failed to create transaction")
		}
		defer func() {
			err = qtx.Rollback(ctx)
		}()

		// write block
		err = qtx.AddBlock(ctx, entity.Block{
			BlockHeight: block.Header.Height,
			BlockHash:   block.Header.Hash.String(),
			Module:      p.Name(),
		})
		if err != nil {
			return errors.Wrapf(err, "Failed to add block %d", block.Header.Height)
		}
		// for each events
		for _, event := range events {
			logger.InfoContext(ctx, "Nodesale processing event",
				slogx.Uint32("txIndex", event.transaction.Index),
				slogx.Int64("blockHeight", block.Header.Height),
				slogx.Stringer("blockhash", block.Header.Hash),
			)
			eventMessage := event.eventMessage
			switch eventMessage.Action {
			case protobuf.Action_ACTION_DEPLOY:
				err = p.processDeploy(ctx, qtx, block, event)
				if err != nil {
					return errors.Wrapf(err, "Failed to deploy at block %d", block.Header.Height)
				}
			case protobuf.Action_ACTION_DELEGATE:
				err = p.processDelegate(ctx, qtx, block, event)
				if err != nil {
					return errors.Wrapf(err, "Failed to delegate at block %d", block.Header.Height)
				}
			case protobuf.Action_ACTION_PURCHASE:
				err = p.processPurchase(ctx, qtx, block, event)
				if err != nil {
					return errors.Wrapf(err, "Failed to purchase at block %d", block.Header.Height)
				}
			}
		}
		// close transaction
		err = qtx.Commit(ctx)
		if err != nil {
			return errors.Wrap(err, "Failed to commit transaction")
		}
		logger.InfoContext(ctx, "Nodesale finished processing block",
			slogx.Int64("block", block.Header.Height),
			slogx.Stringer("hash", block.Header.Hash))
	}
	return nil
}

// RevertData implements indexer.Processor.
func (p *Processor) RevertData(ctx context.Context, from int64) error {
	qtx, err := p.datagateway.BeginNodesaleTx(ctx)
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

func (p *Processor) Shutdown(ctx context.Context) error {
	for _, cleanupFunc := range p.cleanupFuncs {
		err := cleanupFunc(ctx)
		if err != nil {
			return errors.Wrap(err, "cleanup function error")
		}
	}
	return nil
}

var _ indexer.Processor[*types.Block] = (*Processor)(nil)
