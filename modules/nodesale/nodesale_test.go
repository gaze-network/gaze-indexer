package nodesale

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	repository "github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var p *Processor

var postgresConf postgres.Config = postgres.Config{
	User:     "postgres",
	Password: "P@ssw0rd",
}

var qtx gen.Querier

var ctx context.Context

var tx pgx.Tx

var (
	testBlockHeigh int = 101
	testTxIndex    int = 1
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	ctx = context.Background()

	db, _ := postgres.NewPool(ctx, postgresConf)

	repo := repository.NewRepository(db)

	p = &Processor{
		repository: repo,
		network:    common.NetworkMainnet,
	}
	repo.Queries.ClearEvents(ctx)

	tx, _ = p.repository.Db.Begin(ctx)
	qtx = p.repository.WithTx(tx)

	res := m.Run()
	tx.Commit(ctx)
	os.Exit(res)
}

func assembleTestEvent(privateKey *secp256k1.PrivateKey, blockHashHex, txHashHex string, blockHeight, txIndex int, message *protobuf.NodeSaleEvent) (nodesaleEvent, *types.Block) {
	blockHash, _ := chainhash.NewHashFromStr(blockHashHex)
	txHash, _ := chainhash.NewHashFromStr(txHashHex)

	rawData, _ := proto.Marshal(message)

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_FALSE)
	builder.AddOp(txscript.OP_IF)
	builder.AddData(rawData)
	builder.AddOp(txscript.OP_ENDIF)
	script, _ := builder.Script()
	tapleaf := txscript.NewBaseTapLeaf(script)
	scriptTree := txscript.AssembleTaprootScriptTree(tapleaf)
	rootHash := scriptTree.RootNode.TapHash()
	tapkey := txscript.ComputeTaprootOutputKey(privateKey.PubKey(), rootHash[:])

	addressTaproot, _ := btcutil.NewAddressTaproot(schnorr.SerializePubKey(tapkey), p.network.ChainParams())

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
		txAddress:    addressTaproot,
		rawScript:    script,
	}
	block := &types.Block{
		Header: types.BlockHeader{
			Timestamp: time.Now().UTC(),
		},
	}
	return event, block
}
