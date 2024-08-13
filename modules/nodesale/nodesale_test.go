package nodesale

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	testBlockHeight uint64 = 101
	testTxIndex     uint32 = 1
)

func assembleTestEvent(privateKey *secp256k1.PrivateKey, blockHashHex, txHashHex string, blockHeight uint64, txIndex uint32, message *protobuf.NodeSaleEvent) (NodeSaleEvent, *types.Block) {
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
