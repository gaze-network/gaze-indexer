package nodesale

import (
	"context"
	"flag"
	"os"
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

var qtx datagateway.NodesaleDataGatewayWithTx

var ctx context.Context

var (
	testBlockHeigh int64 = 101
	testTxIndex    int   = 1
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		return
	}

	ctx = context.Background()

	db, _ := postgres.NewPool(ctx, postgresConf)

	repo := gen.New(db)
	datagateway := repository.NewRepository(db)

	p = &Processor{
		datagateway: datagateway,
		network:     common.NetworkMainnet,
	}
	repo.ClearEvents(ctx)

	qtx, _ = p.datagateway.BeginNodesaleTx(ctx)

	res := m.Run()
	qtx.Commit(ctx)
	db.Close()
	os.Exit(res)
}

func assembleTestEvent(privateKey *secp256k1.PrivateKey, blockHashHex, txHashHex string, blockHeight int64, txIndex int, message *protobuf.NodeSaleEvent) (nodesaleEvent, *types.Block) {
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
		blockHeight = testBlockHeigh
		testBlockHeigh++
	}
	if txIndex == 0 {
		txIndex = testTxIndex
		testTxIndex++
	}

	event := nodesaleEvent{
		transaction: &types.Transaction{
			BlockHeight: int64(blockHeight),
			BlockHash:   *blockHash,
			Index:       uint32(txIndex),
			TxHash:      *txHash,
		},
		rawData:      rawData,
		eventMessage: message,
		eventJson:    messageJson,
		txPubkey:     privateKey.PubKey(),
	}
	block := &types.Block{
		Header: types.BlockHeader{
			Timestamp: time.Now().UTC(),
		},
	}
	return event, block
}
