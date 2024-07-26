package nodesale

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var p *Processor

var postgresConf postgres.Config = postgres.Config{
	User:     "postgres",
	Password: "P@ssw0rd",
	DBName:   "gaze_indexer",
}

// TODO: should use mock dg or run db in CI
var qtx datagateway.NodeSaleDataGatewayWithTx

var ctx context.Context

var (
	testBlockHeight int64 = 101
	testTxIndex     int   = 1
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		return
	}

	ctx = context.Background()

	db, err := postgres.NewPool(ctx, postgresConf)
	if err != nil {
		return
	}
	defer db.Close()

	repo := gen.New(db)
	datagateway := repository.NewRepository(db)

	p = &Processor{
		NodeSaleDg: datagateway,
		Network:    common.NetworkMainnet,
	}
	repo.ClearEvents(ctx)

	qtx, err = p.NodeSaleDg.BeginNodeSaleTx(ctx)
	if err != nil {
		return
	}

	m.Run()
	qtx.Commit(ctx)
}

func assembleTestEvent(privateKey *secp256k1.PrivateKey, blockHashHex, txHashHex string, blockHeight int64, txIndex int, message *protobuf.NodeSaleEvent) (NodeSaleEvent, *types.Block) {
	blockHash, _ := chainhash.NewHashFromStr(blockHashHex)
	txHash, _ := chainhash.NewHashFromStr(txHashHex)

	rawData, _ := proto.Marshal(message)

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_FALSE)
	builder.AddOp(txscript.OP_IF)
	builder.AddData(rawData)
	builder.AddOp(txscript.OP_ENDIF)

	messageJson, _ := protojson.Marshal(message)

	if blockHeight == 0 {
		blockHeight = testBlockHeight
		testBlockHeight++
	}
	if txIndex == 0 {
		txIndex = testTxIndex
		testTxIndex++
	}

	event := NodeSaleEvent{
		Transaction: &types.Transaction{
			BlockHeight: int64(blockHeight),
			BlockHash:   *blockHash,
			Index:       uint32(txIndex),
			TxHash:      *txHash,
		},
		RawData:      rawData,
		EventMessage: message,
		EventJson:    messageJson,
		TxPubkey:     privateKey.PubKey(),
	}
	block := &types.Block{
		Header: types.BlockHeader{
			Timestamp: time.Now().UTC(),
		},
	}
	return event, block
}
