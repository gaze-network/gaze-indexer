package nodesale

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
)

type Processor struct {
	repository          repository.Repository
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

func extractNodesaleData(witness [][]byte) ([]byte, bool) {
	tokenizer, isTapScript := extractTapScript(witness)
	if !isTapScript {
		return []byte{}, false
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
				return data, true
			}
			state = 0
		}
	}
	return []byte{}, false
}

type nodesaleEvent struct {
	transaction  *types.Transaction
	eventMessage *protobuf.NodeSaleEvent
	eventJson    []byte
	txAddress    btcutil.Address
	rawData      []byte
}

func (p *Processor) parseTransactions(ctx context.Context, transactions []*types.Transaction) ([]nodesaleEvent, error) {
	events := make([]nodesaleEvent, 1)
	for _, t := range transactions {
		for _, txIn := range t.TxIn {
			data, isNodesale := extractNodesaleData(txIn.Witness)
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
			})
		}
	}
	return events, nil
}

func (p *Processor) pubkeyToAddress(pubkey string) btcutil.Address {
	pubKeyBytes, _ := hex.DecodeString(pubkey)
	pubKey, _ := btcec.ParsePubKey(pubKeyBytes)
	sellerAddr, _ := btcutil.NewAddressTaproot(schnorr.SerializePubKey(pubKey), p.network.ChainParams())
	return sellerAddr
}

func (p *Processor) processDeploy(ctx context.Context, qtx gen.Querier, block *types.Block, event nodesaleEvent) error {
	valid := true
	deploy := event.eventMessage.Deploy
	sellerAddr := p.pubkeyToAddress(string(deploy.SellerPublicKey))
	if !bytes.Equal(
		[]byte(sellerAddr.EncodeAddress()),
		[]byte(event.txAddress.EncodeAddress()),
	) {
		valid = false
	}
	tiers := make([][]byte, len(deploy.Tiers))
	for _, tier := range deploy.Tiers {
		tierJson, err := protojson.Marshal(tier)
		if err != nil {
			return fmt.Errorf("Failed to parse tiers to json : %w", err)
		}
		tiers = append(tiers, tierJson)
	}

	err := qtx.AddEvent(ctx, gen.AddEventParams{
		TxHash:         event.transaction.TxHash.String(),
		TxIndex:        int32(event.transaction.Index),
		Action:         int32(event.eventMessage.Action),
		RawMessage:     event.rawData,
		ParsedMessage:  event.eventJson,
		BlockTimestamp: pgtype.Timestamp{Time: block.Header.Timestamp, Valid: true},
		BlockHash:      block.Header.Hash.String(),
		BlockHeight:    int32(block.Header.Height),
		Valid:          valid,
		WalletAddress:  event.txAddress.EncodeAddress(),
	})
	if err != nil {
		return fmt.Errorf("Failed to insert event : %w", err)
	}
	if valid {
		err = qtx.AddNodesale(ctx, gen.AddNodesaleParams{
			BlockHeight: int32(event.transaction.BlockHeight),
			TxIndex:     int32(event.transaction.Index),
			Name:        deploy.Name,
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
			DeployTxHash:    event.transaction.TxHash.String(),
		})
		if err != nil {
			return fmt.Errorf("Failed to insert nodesale : %w", err)
		}
	}

	return nil
}

func (p *Processor) processDelegate(ctx context.Context, qtx gen.Querier, block *types.Block, event nodesaleEvent) error {
	valid := true
	delegate := event.eventMessage.Delegate
	nodeIds := make([]int32, len(delegate.NodeIDs))
	for _, id := range delegate.NodeIDs {
		nodeIds = append(nodeIds, int32(id))
	}
	nodes, err := qtx.GetNodes(ctx, gen.GetNodesParams{
		SaleBlock:   int32(delegate.DeployID.Block),
		SaleTxIndex: int32(delegate.DeployID.TxIndex),
		NodeIds:     nodeIds,
	})
	if err != nil {
		return fmt.Errorf("Failed to get nodes : %w", err)
	}

	if len(nodeIds) != len(nodes) {
		valid = false
	}

	for _, node := range nodes {
		ownerAddress := p.pubkeyToAddress(node.OwnerPublicKey)
		if !bytes.Equal(
			[]byte(ownerAddress.EncodeAddress()),
			[]byte(event.txAddress.EncodeAddress()),
		) {
			valid = false
		}
	}

	err = qtx.AddEvent(ctx, gen.AddEventParams{
		TxHash:         event.transaction.TxHash.String(),
		TxIndex:        int32(event.transaction.Index),
		Action:         int32(event.eventMessage.Action),
		RawMessage:     event.rawData,
		ParsedMessage:  event.eventJson,
		BlockTimestamp: pgtype.Timestamp{Time: block.Header.Timestamp, Valid: true},
		BlockHash:      block.Header.Hash.String(),
		BlockHeight:    int32(block.Header.Height),
		Valid:          valid,
		WalletAddress:  event.txAddress.EncodeAddress(),
	})
	if err != nil {
		return fmt.Errorf("Failed to insert event : %w", err)
	}

	if valid {
		_, err = qtx.SetDelegates(ctx, gen.SetDelegatesParams{
			SaleBlock:   int32(event.transaction.BlockHeight),
			SaleTxIndex: int32(event.transaction.Index),
			Delegatee:   string(delegate.DelegateePublicKey),
			NodeIds:     nodeIds,
		})
		if err != nil {
			return fmt.Errorf("Failed to set delegate : %w", err)
		}
	}

	return nil
}

func (p *Processor) processPurchase(ctx context.Context, qtx gen.Querier, block *types.Block, event nodesaleEvent) error {
	panic("unimplemented")
}

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	for _, block := range inputs {
		// parse all event from each transaction including reading tx wallet
		events, err := p.parseTransactions(ctx, block.Transactions)
		if err != nil {
			return fmt.Errorf("Invalid data from bitcoin client : %w", err)
		}
		// events[]

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

		/*if err := p.processBlock(ctx, block); err != nil {
			return fmt.Errorf("Process block %d failed. : %w", block.Header.Height, err)
		}*/
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

var _ indexer.Processor[*types.Block] = (*Processor)(nil)
