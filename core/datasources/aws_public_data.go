// AWS Public Blockchain Datasource
// - https://registry.opendata.aws/aws-public-blockchain
// - https://github.com/aws-solutions-library-samples/guidance-for-digital-assets-on-aws
//
// To setup your own data source, see: https://github.com/aws-solutions-library-samples/guidance-for-digital-assets-on-aws/blob/main/analytics/producer/README.md
package datasources

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/subscription"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/parquetutils"
	"github.com/samber/lo"
	"github.com/xitongsys/parquet-go/reader"
)

const (
	awsPublicDataS3Region = "us-east-2"
	awsPublicDataS3Bucket = "aws-public-blockchain"

	parquetReaderConcurrency = 8
)

var firstBitcoinTimestamp = time.Date(2009, time.January, 3, 18, 15, 5, 0, time.UTC)

// Make sure to implement the BitcoinDatasource interface
var _ Datasource[*types.Block] = (*AWSPublicDataDatasource)(nil)

type AWSPublicDataDatasource struct {
	s3Client          *s3.Client
	s3Bucket          string
	btcclient         *rpcclient.Client
	btcNodeDatasource *BitcoinNodeDatasource
}

func NewAWSPublicData(ctx context.Context, btcclient *rpcclient.Client) (*AWSPublicDataDatasource, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't load aws user config")
	}

	// TODO: support user defined config (self-hosted s3 bucket)
	s3client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.Region = awsPublicDataS3Region
		o.Credentials = aws.AnonymousCredentials{}
	})

	return &AWSPublicDataDatasource{
		s3Client:          s3client,
		s3Bucket:          awsPublicDataS3Bucket,
		btcclient:         btcclient,
		btcNodeDatasource: NewBitcoinNode(btcclient),
	}, nil
}

func (AWSPublicDataDatasource) Name() string {
	return "aws_public_data"
}

func (d *AWSPublicDataDatasource) Fetch(ctx context.Context, from, to int64) ([]*types.Block, error) {
	ch := make(chan []*types.Block)
	subscription, err := d.FetchAsync(ctx, from, to, ch)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer subscription.Unsubscribe()

	blocks := make([]*types.Block, 0)
	for {
		select {
		case b, ok := <-ch:
			if !ok {
				return blocks, nil
			}
			blocks = append(blocks, b...)
		case <-subscription.Done():
			if err := ctx.Err(); err != nil {
				return nil, errors.Wrap(err, "context done")
			}
			return blocks, nil
		case err := <-subscription.Err():
			if err != nil {
				return nil, errors.Wrap(err, "got error while fetch async")
			}
			return blocks, nil
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "context done")
		}
	}
}

func (d *AWSPublicDataDatasource) FetchAsync(ctx context.Context, from, to int64, ch chan<- []*types.Block) (*subscription.ClientSubscription[[]*types.Block], error) {
	ctx = logger.WithContext(ctx,
		slogx.String("package", "datasources"),
		slogx.String("datasource", d.Name()),
	)

	from, to, skip, err := d.prepareRange(from, to)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare fetch range")
	}

	subscription := subscription.NewSubscription(ch)
	if skip {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return subscription.Client(), nil
	}

	fromBlockHeader, err := d.GetBlockHeader(ctx, from)
	if err != nil {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return nil, errors.Wrapf(err, "failed to get block header for %v", from)
	}

	fromFiles, err := d.listBlocksFilesByDate(ctx, fromBlockHeader.Timestamp)
	if err != nil {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return nil, errors.Wrap(err, "failed to list files by date")
	}

	// supported only merged blocks files
	fromFiles = lo.Filter(fromFiles, func(key string, _ int) bool {
		return strings.Contains(key, "part-")
	})

	// use bitcoin node instead of s3
	if len(fromFiles) == 0 {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return d.btcNodeDatasource.FetchAsync(ctx, from, to, ch)
	}

	// Parallel fetch blocks from Bitcoin node until complete all block heights
	// or subscription is done.
	go func() {
		timestamp := fromBlockHeader.Timestamp
		for {
			// prevent infinite loop
			if timestamp.After(time.Now()) {
				break
			}

			ctx = logger.WithContext(ctx,
				slogx.Time("file_date", timestamp),
			)

			blocksFiles, err := d.listBlocksFilesByDate(ctx, timestamp)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to list blocks files by date from aws s3", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.WithStack(err)); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
				return
			}

			txsFiles, err := d.listTxsFilesByDate(ctx, timestamp)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to list txs files by date from aws s3", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.WithStack(err)); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
				return
			}

			blocksFiles = lo.Filter(blocksFiles, func(key string, _ int) bool {
				return strings.Contains(key, "part-")
			})
			txsFiles = lo.Filter(txsFiles, func(key string, _ int) bool {
				return strings.Contains(key, "part-")
			})

			// reach the end of supported data
			if len(blocksFiles) == 0 || len(txsFiles) == 0 {
				return
			}

			// prevent unexpected error
			if len(blocksFiles) != 1 {
				logger.ErrorContext(ctx, "Unexpected blocks files count, should be 1", slogx.Int("count", len(blocksFiles)))
				if err := subscription.SendError(ctx, errors.Wrap(errs.InternalError, "unexpected blocks files count")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
				return
			}
			if len(txsFiles) != 1 {
				logger.ErrorContext(ctx, "Unexpected txs files count, should be 1", slogx.Int("count", len(txsFiles)))
				if err := subscription.SendError(ctx, errors.Wrap(errs.InternalError, "unexpected txs files count")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
				return
			}

			// TODO: use concurrent stream (max 2 goroutine) to download files then sequentially read parquet files
			// to improve performance while not consuming too much memory (increase around 500 MB per goroutine)

			blocksData, err := d.downloadFile(ctx, blocksFiles[0])
			if err != nil {
				logger.ErrorContext(ctx, "Failed to download blocks file from AWS S3", slogx.Int("count", len(txsFiles)))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't download blocks file")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}
			txsData, err := d.downloadFile(ctx, txsFiles[0])
			if err != nil {
				logger.ErrorContext(ctx, "Failed to download blocks file from AWS S3", slogx.Int("count", len(txsFiles)))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't download blocks file")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}

			// read parquet files
			// TODO: create read function to reduce duplicate code
			pr, err := reader.NewParquetReader(parquetutils.NewBufferFile(blocksData), new(awsBlock), parquetReaderConcurrency)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to create parquet reader for blocks data", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't create parquet reader for blocks data")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}

			rawBlocks := make([]awsBlock, pr.GetNumRows())
			if err = pr.Read(&rawBlocks); err != nil {
				logger.ErrorContext(ctx, "Failed to read parquet blocks data", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't read parquet blocks data")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}
			pr.ReadStop()
			slices.SortFunc(rawBlocks, func(i, j awsBlock) int {
				return cmp.Compare(i.Number, j.Number)
			})

			// TODO: partial read txs data to reduce memory usage (use txs_count from block data to read only needed txs and stream to subscription)
			_ = txsData

			blocks := make([]*types.Block, 0, len(rawBlocks))
			for _, rawBlock := range rawBlocks {
				blockHeader, err := rawBlock.ToBlockHeader()
				if err != nil {
					logger.ErrorContext(ctx, "Failed to convert aws block to type block header", slogx.Error(err))
					if err := subscription.SendError(ctx, errors.Wrap(err, "can't convert aws block to type block header")); err != nil {
						logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
					}
					return
				}

				blocks = append(blocks, &types.Block{
					Header:       blockHeader,
					Transactions: []*types.Transaction{},
				})
			}

			if err := subscription.Send(ctx, blocks); err != nil {
				if errors.Is(err, errs.Closed) {
					return
				}
				logger.WarnContext(ctx, "Failed to send bitcoin blocks to subscription client",
					slogx.Int64("start", blocks[0].Header.Height),
					slogx.Int64("end", blocks[len(blocks)-1].Header.Height),
					slogx.Error(err),
				)
			}

			// Continue to next day
			timestamp = timestamp.Add(24 * time.Hour)
		}
	}()

	return subscription.Client(), nil
}

func (d *AWSPublicDataDatasource) GetBlockHeader(ctx context.Context, height int64) (types.BlockHeader, error) {
	header, err := d.btcNodeDatasource.GetBlockHeader(ctx, height)
	return header, errors.WithStack(err)
}

func (d *AWSPublicDataDatasource) prepareRange(fromHeight, toHeight int64) (start, end int64, skip bool, err error) {
	start = fromHeight
	end = toHeight

	// get current bitcoin block height
	latestBlockHeight, err := d.btcclient.GetBlockCount()
	if err != nil {
		return -1, -1, false, errors.Wrap(err, "failed to get block count")
	}

	// set start to genesis block height
	if start < 0 {
		start = 0
	}

	// set end to current bitcoin block height if
	// - end is -1
	// - end is greater that current bitcoin block height
	if end < 0 || end > latestBlockHeight {
		end = latestBlockHeight
	}

	// if start is greater than end, skip this round
	if start > end {
		return -1, -1, true, nil
	}

	return start, end, false, nil
}

func (d *AWSPublicDataDatasource) listFiles(ctx context.Context, prefix string) ([]string, error) {
	result, err := d.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.s3Bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "can't list s3 bucket objects for bucket %q and prefix %q", d.s3Bucket, prefix)
	}

	// filter empty keys
	objs := lo.Filter(result.Contents, func(item s3types.Object, _ int) bool { return item.Key != nil })

	return lo.Map(objs, func(item s3types.Object, _ int) string {
		return *item.Key
	}), nil
}

func (d *AWSPublicDataDatasource) listBlocksFilesByDate(ctx context.Context, date time.Time) ([]string, error) {
	if date.Before(firstBitcoinTimestamp) {
		return nil, errors.Wrapf(errs.InvalidArgument, "date %v is before first bitcoin timestamp %v", date, firstBitcoinTimestamp)
	}
	prefix := "v1.0/btc/blocks/date=" + date.Format(time.DateOnly)
	files, err := d.listFiles(ctx, prefix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list blocks files by date")
	}
	return files, nil
}

func (d *AWSPublicDataDatasource) listTxsFilesByDate(ctx context.Context, date time.Time) ([]string, error) {
	if date.Before(firstBitcoinTimestamp) {
		return nil, errors.Wrapf(errs.InvalidArgument, "date %v is before first bitcoin timestamp %v", date, firstBitcoinTimestamp)
	}
	prefix := "v1.0/btc/transactions/date=" + date.Format(time.DateOnly)
	files, err := d.listFiles(ctx, prefix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list txs files by date")
	}
	return files, nil
}

func (d *AWSPublicDataDatasource) downloadFile(ctx context.Context, key string) ([]byte, error) {
	downloader := manager.NewDownloader(d.s3Client, func(d *manager.Downloader) {
		d.Concurrency = 16
		d.PartSize = 10 * 1024 * 1024
	})

	buffer := manager.NewWriteAtBuffer([]byte{})
	numBytes, err := downloader.Download(ctx, buffer, &s3.GetObjectInput{
		Bucket: aws.String(d.s3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download file for bucket %q and key %q", d.s3Bucket, key)
	}

	if numBytes < 1 {
		return nil, errors.Wrap(errs.NotFound, "got empty file")
	}

	return buffer.Bytes(), nil
}

// TODO: remove unused fields to reduce memory usage
type (
	awsBlock struct {
		Hash              string  `parquet:"name=hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Version           int64   `parquet:"name=version, type=INT64, repetitiontype=OPTIONAL"`
		MedianTime        string  `parquet:"name=mediantime, type=INT96, repetitiontype=OPTIONAL"`
		Nonce             int64   `parquet:"name=nonce, type=INT64, repetitiontype=OPTIONAL"`
		Bits              string  `parquet:"name=bits, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Difficulty        float64 `parquet:"name=difficulty, type=DOUBLE, repetitiontype=OPTIONAL"`
		Chainwork         string  `parquet:"name=chainwork, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		PreviousBlockHash string  `parquet:"name=previousblockhash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Size              int64   `parquet:"name=size, type=INT64, repetitiontype=OPTIONAL"`
		Weight            int64   `parquet:"name=weight, type=INT64, repetitiontype=OPTIONAL"`
		CoinbaseParam     string  `parquet:"name=coinbase_param, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Number            int64   `parquet:"name=number, type=INT64, repetitiontype=OPTIONAL"`
		TransactionCount  int64   `parquet:"name=transaction_count, type=INT64, repetitiontype=OPTIONAL"`
		MerkleRoot        string  `parquet:"name=merkle_root, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		StrippedSize      int64   `parquet:"name=stripped_size, type=INT64, repetitiontype=OPTIONAL"`
		Timestamp         string  `parquet:"name=timestamp, type=INT96, repetitiontype=OPTIONAL"`
		Date              string  `parquet:"name=date, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		LastModified      string  `parquet:"name=last_modified, type=INT96, repetitiontype=OPTIONAL"`
	}
	awsTransaction struct {
		Hash         string  `parquet:"name=hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Version      int64   `parquet:"name=version, type=INT64, repetitiontype=OPTIONAL"`
		Size         int64   `parquet:"name=size, type=INT64, repetitiontype=OPTIONAL"`
		BlockHash    string  `parquet:"name=block_hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		BlockNumber  int64   `parquet:"name=block_number, type=INT64, repetitiontype=OPTIONAL"`
		Index        int64   `parquet:"name=index, type=INT64, repetitiontype=OPTIONAL"`
		Virtual_size int64   `parquet:"name=virtual_size, type=INT64, repetitiontype=OPTIONAL"`
		Lock_time    int64   `parquet:"name=lock_time, type=INT64, repetitiontype=OPTIONAL"`
		Input_count  int64   `parquet:"name=input_count, type=INT64, repetitiontype=OPTIONAL"`
		Output_count int64   `parquet:"name=output_count, type=INT64, repetitiontype=OPTIONAL"`
		Is_coinbase  bool    `parquet:"name=is_coinbase, type=BOOLEAN, repetitiontype=OPTIONAL"`
		Output_value float64 `parquet:"name=output_value, type=DOUBLE, repetitiontype=OPTIONAL"`
		Outputs      []*struct {
			Address             string  `parquet:"name=address, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Index               int64   `parquet:"name=index, type=INT64, repetitiontype=OPTIONAL"`
			Required_signatures int64   `parquet:"name=required_signatures, type=INT64, repetitiontype=OPTIONAL"`
			Script_asm          string  `parquet:"name=script_asm, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Script_hex          string  `parquet:"name=script_hex, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Type                string  `parquet:"name=type, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Value               float64 `parquet:"name=value, type=DOUBLE, repetitiontype=OPTIONAL"`
		} `parquet:"name=outputs, type=LIST, repetitiontype=OPTIONAL, valuetype=STRUCT"`
		Block_timestamp string  `parquet:"name=block_timestamp, type=INT96, repetitiontype=OPTIONAL"`
		Date            string  `parquet:"name=date, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Last_modified   string  `parquet:"name=last_modified, type=INT96, repetitiontype=OPTIONAL"`
		Fee             float64 `parquet:"name=fee, type=DOUBLE, repetitiontype=OPTIONAL"`
		Input_value     float64 `parquet:"name=input_value, type=DOUBLE, repetitiontype=OPTIONAL"`
		Inputs          []*struct {
			Address                string    `parquet:"name=address, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Index                  int64     `parquet:"name=index, type=INT64, repetitiontype=OPTIONAL"`
			Required_signatures    int64     `parquet:"name=required_signatures, type=INT64, repetitiontype=OPTIONAL"`
			Script_asm             string    `parquet:"name=script_asm, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Script_hex             string    `parquet:"name=script_hex, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Sequence               int64     `parquet:"name=sequence, type=INT64, repetitiontype=OPTIONAL"`
			Spent_output_index     int64     `parquet:"name=spent_output_index, type=INT64, repetitiontype=OPTIONAL"`
			Spent_transaction_hash string    `parquet:"name=spent_transaction_hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Txinwitness            []*string `parquet:"name=txinwitness, type=LIST, repetitiontype=OPTIONAL, valuetype=BYTE_ARRAY, convertedtype=UTF8"`
			Type                   string    `parquet:"name=type, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
			Value                  float64   `parquet:"name=value, type=DOUBLE, repetitiontype=OPTIONAL"`
		} `parquet:"name=inputs, type=LIST, repetitiontype=OPTIONAL, valuetype=STRUCT"`
	}
)

func (a awsBlock) ToBlockHeader() (types.BlockHeader, error) {
	hash, err := chainhash.NewHashFromStr(a.Hash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert hash")
	}
	prevBlockHash, err := chainhash.NewHashFromStr(a.PreviousBlockHash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert previous block hash")
	}
	merkleRoot, err := chainhash.NewHashFromStr(a.MerkleRoot)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert merkle root")
	}
	return types.BlockHeader{
		Hash:       *hash,
		Height:     a.Number,
		Version:    int32(a.Version),
		PrevBlock:  *prevBlockHash,
		MerkleRoot: *merkleRoot,
		Timestamp:  time.Time{}, // TODO:  parse timestamp
		Bits:       0,           // TODO: parse bits
		Nonce:      uint32(a.Nonce),
	}, nil
}

func (a awsTransaction) ToTransaction() (types.Transaction, error) {
	// TODO: implement this
	return types.Transaction{
		BlockHeight: 0,
		BlockHash:   [32]byte{},
		Index:       0,
		TxHash:      [32]byte{},
		Version:     0,
		LockTime:    0,
		TxIn:        []*types.TxIn{},
		TxOut:       []*types.TxOut{},
	}, nil
}
