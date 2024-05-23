// AWS Public Blockchain Datasource
// - https://registry.opendata.aws/aws-public-blockchain
// - https://github.com/aws-solutions-library-samples/guidance-for-digital-assets-on-aws
//
// To setup your own data source, see: https://github.com/aws-solutions-library-samples/guidance-for-digital-assets-on-aws/blob/main/analytics/producer/README.md
package datasources

import (
	"cmp"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/subscription"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/parquetutils"
	"github.com/samber/lo"
	"github.com/xitongsys/parquet-go/reader"
	parquettypes "github.com/xitongsys/parquet-go/types"
	"golang.org/x/sync/errgroup"
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
	btcDatasource Datasource[*types.Block]
	s3Client      *s3.Client
	s3Bucket      string
}

func NewAWSPublicData(btcDatasource Datasource[*types.Block]) *AWSPublicDataDatasource {
	sdkConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logger.Panic("Can't load AWS SDK user config", slogx.Error(err), slog.String("package", "datasources"))
	}

	// TODO: support user defined config (self-hosted s3 bucket)
	s3client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.Region = awsPublicDataS3Region
		o.Credentials = aws.AnonymousCredentials{}
	})

	return &AWSPublicDataDatasource{
		btcDatasource: btcDatasource,
		s3Client:      s3client,
		s3Bucket:      awsPublicDataS3Bucket,
	}
}

func (d AWSPublicDataDatasource) Name() string {
	return fmt.Sprintf("aws_public_data/%s", d.btcDatasource.Name())
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

	start, end, skip, err := d.prepareRange(ctx, from, to)
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

	startFiles, err := d.listBlocksFilesByDate(ctx, start.Timestamp)
	if err != nil {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return nil, errors.Wrap(err, "failed to list files by date")
	}

	// supported only merged blocks files
	startFiles = lo.Filter(startFiles, func(key string, _ int) bool {
		return strings.Contains(key, "part-")
	})

	// use other datasource instead of s3 if there's no supported data
	if len(startFiles) == 0 {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		s, err := d.btcDatasource.FetchAsync(ctx, start.Height, end.Height, ch)
		return s, errors.WithStack(err)
	}

	go func() {
		defer func() {
			// add a bit delay to prevent shutdown before client receive all blocks
			time.Sleep(100 * time.Millisecond)

			subscription.Unsubscribe()
		}()
		// loop through each day until reach the end of supported data or within end block date
		for ts := start.Timestamp; ts.Before(end.Timestamp.Round(24*time.Hour)) && ts.Before(time.Now()); ts = ts.Add(24 * time.Hour) {
			ctx := logger.WithContext(ctx,
				slogx.Time("date", ts),
				slogx.Int64("date_unix", ts.Unix()),
			)

			logger.DebugContext(ctx, "Fetching data from AWS S3", slogx.Int64("start", start.Height), slogx.Int64("end", end.Height))

			allBlocksFiles, err := d.listBlocksFilesByDate(ctx, ts)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to list blocks files by date from aws s3", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.WithStack(err)); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
				return
			}

			allTxsFiles, err := d.listTxsFilesByDate(ctx, ts)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to list txs files by date from aws s3", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.WithStack(err)); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
				return
			}

			blocksFiles := lo.Filter(allBlocksFiles, func(key string, _ int) bool {
				return strings.Contains(key, "part-")
			})
			txsFiles := lo.Filter(allTxsFiles, func(key string, _ int) bool {
				return strings.Contains(key, "part-")
			})

			logger.DebugContext(ctx, "Found files in AWS S3 bucket",
				slogx.Int("files_blocks", len(allBlocksFiles)),
				slogx.Int("files_blocks_merged", len(blocksFiles)),
				slogx.Int("files_txs_all", len(allTxsFiles)),
				slogx.Int("files_txs_merged", len(txsFiles)),
			)

			// Reach the end of supported data,
			// stop fetching data from AWS S3
			if len(blocksFiles) == 0 || len(txsFiles) == 0 {
				logger.DebugContext(ctx, "No blocks files found, stop fetching data from AWS S3")
				return
			}

			// prevent unexpected error
			{
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
			}

			// TODO: use concurrent stream (max 2 goroutine) to download files then sequentially read parquet files
			// to improve performance while not consuming too much memory (increase around 500 MB per goroutine)

			// TODO: create iobuffer that's implement io.WriterAt and parquetsource.ParquetFile interface
			var (
				// TODO: create []byte pool to reduce alloc
				blocksBuffer = manager.NewWriteAtBuffer([]byte{})
				txsBuffer    = manager.NewWriteAtBuffer([]byte{})
			)
			startDownload := time.Now()
			if err := d.downloadFile(ctx, blocksFiles[0], blocksBuffer); err != nil {
				logger.ErrorContext(ctx, "Failed to download blocks file from AWS S3", slogx.Int("count", len(txsFiles)))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't download blocks file")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}
			if err := d.downloadFile(ctx, txsFiles[0], txsBuffer); err != nil {
				logger.ErrorContext(ctx, "Failed to download blocks file from AWS S3", slogx.Int("count", len(txsFiles)))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't download blocks file")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}
			logger.DebugContext(ctx, "Downloaded files from AWS S3",
				slogx.Duration("duration", time.Since(startDownload)),
				slogx.Int("sizes_blocks", len(blocksBuffer.Bytes())),
				slogx.Int("sizes_txs", len(txsBuffer.Bytes())),
			)

			// read parquet files
			// TODO: create read function to reduce duplicate code
			startRead := time.Now()
			blocksReader, err := reader.NewParquetReader(parquetutils.NewBufferFile(blocksBuffer.Bytes()), new(awsBlock), parquetReaderConcurrency)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to create parquet reader for blocks data", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't create parquet reader for blocks data")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}

			// we can read all blocks data at once because it's small
			rawAllBlocks := make([]awsBlock, blocksReader.GetNumRows())
			if err = blocksReader.Read(&rawAllBlocks); err != nil {
				logger.ErrorContext(ctx, "Failed to read parquet blocks data", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't read parquet blocks data")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}
			blocksReader.ReadStop()

			// TODO: partial read txs data to reduce memory usage
			// (use txs_count from block data to read only needed txs and stream to subscription)
			txsReader, err := reader.NewParquetReader(parquetutils.NewBufferFile(txsBuffer.Bytes()), new(awsTransaction), parquetReaderConcurrency)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to create parquet reader for txs data", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't create parquet reader for txs data")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}

			// TODO: read only needed txs data
			rawAllTxs := make([]awsTransaction, txsReader.GetNumRows())
			if err = txsReader.Read(&rawAllTxs); err != nil {
				logger.ErrorContext(ctx, "Failed to read parquet txs data", slogx.Error(err))
				if err := subscription.SendError(ctx, errors.Wrap(err, "can't read parquet txs data")); err != nil {
					logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
				}
			}
			txsReader.ReadStop()
			groupRawTxs := lo.GroupBy(rawAllTxs, func(tx awsTransaction) int64 {
				return tx.BlockNumber
			})

			// filter blocks data by height range
			rawFilteredBlocks := lo.Filter(rawAllBlocks, func(block awsBlock, _ int) bool {
				return block.Number >= start.Height && block.Number <= end.Height
			})
			slices.SortFunc(rawFilteredBlocks, func(i, j awsBlock) int {
				return cmp.Compare(i.Number, j.Number)
			})

			logger.DebugContext(ctx, "Read parquet files",
				slogx.Duration("duration", time.Since(startRead)),
				slogx.Int("total_blocks", len(rawAllBlocks)),
				slogx.Int("filtered_blocks", len(rawFilteredBlocks)),
				slogx.Int("total_txs", len(rawAllTxs)),
				slogx.Int("total_txs_grouped", len(groupRawTxs)),
			)

			blocks := make([]*types.Block, 0, len(rawFilteredBlocks))
			for _, rawBlock := range rawFilteredBlocks {
				blockHeader, err := rawBlock.ToBlockHeader()
				if err != nil {
					logger.ErrorContext(ctx, "Failed to convert aws block to type block header", slogx.Error(err))
					if err := subscription.SendError(ctx, errors.Wrap(err, "can't convert aws block to type block header")); err != nil {
						logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
					}
					return
				}

				txs := make([]*types.Transaction, 0, len(groupRawTxs[blockHeader.Height]))
				for _, rawTx := range groupRawTxs[rawBlock.Number] {
					tx, err := rawTx.ToTransaction(rawBlock)
					if err != nil {
						logger.ErrorContext(ctx, "Failed to convert aws transaction to type transaction", slogx.Error(err))
						if err := subscription.SendError(ctx, errors.Wrap(err, "can't convert aws transaction to type transaction")); err != nil {
							logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
						}
						return
					}
					txs = append(txs, tx)
				}
				slices.SortFunc(txs, func(i, j *types.Transaction) int {
					return cmp.Compare(i.Index, j.Index)
				})

				logger.DebugContext(ctx, "Append block to blocks slice", slogx.Int64("height", blockHeader.Height), slogx.Int("txs", len(txs)))
				blocks = append(blocks, &types.Block{
					Header:       blockHeader,
					Transactions: txs,
				})
			}

			logger.DebugContext(ctx, "Send blocks to subscription client", slogx.Int("count", len(blocks)))
			if err := subscription.Send(ctx, blocks); err != nil {
				if errors.Is(err, errs.Closed) {
					logger.DebugContext(ctx, "Subscription client closed, can't send", slogx.Error(err))
					return
				}
				logger.WarnContext(ctx, "Failed to send bitcoin blocks to subscription client",
					slogx.Int64("start", blocks[0].Header.Height),
					slogx.Int64("end", blocks[len(blocks)-1].Header.Height),
					slogx.Error(err),
				)
			}
			logger.DebugContext(ctx, "Blocks sent to subscription client", slogx.Int("count", len(blocks)))
		}
	}()

	return subscription.Client(), nil
}

func (d *AWSPublicDataDatasource) GetBlockHeader(ctx context.Context, height int64) (types.BlockHeader, error) {
	header, err := d.btcDatasource.GetBlockHeader(ctx, height)
	return header, errors.WithStack(err)
}

func (d *AWSPublicDataDatasource) GetCurrentBlockHeight(ctx context.Context) (int64, error) {
	height, err := d.btcDatasource.GetCurrentBlockHeight(ctx)
	return height, errors.WithStack(err)
}

func (d *AWSPublicDataDatasource) prepareRange(ctx context.Context, fromHeight, toHeight int64) (startHeader, endHeader types.BlockHeader, skip bool, err error) {
	start := fromHeight
	end := toHeight

	// get current bitcoin block height
	latestBlockHeight, err := d.btcDatasource.GetCurrentBlockHeight(ctx)
	if err != nil {
		return types.BlockHeader{}, types.BlockHeader{}, false, errors.Wrap(err, "failed to get block count")
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
		return types.BlockHeader{}, types.BlockHeader{}, true, nil
	}

	if err != nil {
		return types.BlockHeader{}, types.BlockHeader{}, false, errors.Wrapf(err, "block %v", end)
	}

	group, groupctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		startHeader, err = d.GetBlockHeader(groupctx, start)
		return errors.Wrapf(err, "block %v", start)
	})
	group.Go(func() error {
		endHeader, err = d.GetBlockHeader(ctx, end)
		return errors.Wrapf(err, "block %v", end)
	})
	if err := group.Wait(); err != nil {
		return types.BlockHeader{}, types.BlockHeader{}, false, errors.Wrap(err, "failed to get block header")
	}

	return startHeader, endHeader, false, nil
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
	prefix := "v1.0/btc/blocks/date=" + date.UTC().Format(time.DateOnly)
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
	prefix := "v1.0/btc/transactions/date=" + date.UTC().Format(time.DateOnly)
	files, err := d.listFiles(ctx, prefix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list txs files by date")
	}
	return files, nil
}

func (d *AWSPublicDataDatasource) downloadFile(ctx context.Context, key string, w io.WriterAt) error {
	downloader := manager.NewDownloader(d.s3Client, func(d *manager.Downloader) {
		d.Concurrency = 16
		d.PartSize = 10 * 1024 * 1024
	})

	numBytes, err := downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(d.s3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to download file for bucket %q and key %q", d.s3Bucket, key)
	}

	if numBytes < 1 {
		return errors.Wrap(errs.NotFound, "got empty file")
	}

	return nil
}

// TODO: remove unused fields to reduce memory usage
type (
	awsBlock struct {
		Hash              string  `parquet:"name=hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Version           int64   `parquet:"name=version, type=INT64, repetitiontype=OPTIONAL"`
		MedianTime        string  `parquet:"name=mediantime, type=INT96, repetitiontype=OPTIONAL"`
		Nonce             int64   `parquet:"name=nonce, type=INT64, repetitiontype=OPTIONAL"`
		Bits              string  `parquet:"name=bits, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"` // Hex string format
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
		Hash           string         `parquet:"name=hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Version        int64          `parquet:"name=version, type=INT64, repetitiontype=OPTIONAL"`
		Size           int64          `parquet:"name=size, type=INT64, repetitiontype=OPTIONAL"`
		BlockHash      string         `parquet:"name=block_hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		BlockNumber    int64          `parquet:"name=block_number, type=INT64, repetitiontype=OPTIONAL"`
		Index          int64          `parquet:"name=index, type=INT64, repetitiontype=OPTIONAL"`
		VirtualSize    int64          `parquet:"name=virtual_size, type=INT64, repetitiontype=OPTIONAL"`
		LockTime       int64          `parquet:"name=lock_time, type=INT64, repetitiontype=OPTIONAL"`
		InputCount     int64          `parquet:"name=input_count, type=INT64, repetitiontype=OPTIONAL"`
		OutputCount    int64          `parquet:"name=output_count, type=INT64, repetitiontype=OPTIONAL"`
		IsCoinbase     bool           `parquet:"name=is_coinbase, type=BOOLEAN, repetitiontype=OPTIONAL"`
		OutputValue    float64        `parquet:"name=output_value, type=DOUBLE, repetitiontype=OPTIONAL"`
		Outputs        []*awsTxOutput `parquet:"name=outputs, type=LIST, repetitiontype=OPTIONAL, valuetype=STRUCT"`
		BlockTimestamp string         `parquet:"name=block_timestamp, type=INT96, repetitiontype=OPTIONAL"`
		Date           string         `parquet:"name=date, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		LastModified   string         `parquet:"name=last_modified, type=INT96, repetitiontype=OPTIONAL"`
		Fee            float64        `parquet:"name=fee, type=DOUBLE, repetitiontype=OPTIONAL"`
		InputValue     float64        `parquet:"name=input_value, type=DOUBLE, repetitiontype=OPTIONAL"`
		Inputs         []*awsTxInput  `parquet:"name=inputs, type=LIST, repetitiontype=OPTIONAL, valuetype=STRUCT"`
	}
	awsTxInput struct {
		Address              string    `parquet:"name=address, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Index                int64     `parquet:"name=index, type=INT64, repetitiontype=OPTIONAL"`
		RequiredSignatures   int64     `parquet:"name=required_signatures, type=INT64, repetitiontype=OPTIONAL"`
		ScriptAsm            string    `parquet:"name=script_asm, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		ScriptHex            string    `parquet:"name=script_hex, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Sequence             int64     `parquet:"name=sequence, type=INT64, repetitiontype=OPTIONAL"`
		SpentOutputIndex     int64     `parquet:"name=spent_output_index, type=INT64, repetitiontype=OPTIONAL"`
		SpentTransactionHash string    `parquet:"name=spent_transaction_hash, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		TxInWitness          []*string `parquet:"name=txinwitness, type=LIST, repetitiontype=OPTIONAL, valuetype=BYTE_ARRAY, convertedtype=UTF8"`
		Type                 string    `parquet:"name=type, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Value                float64   `parquet:"name=value, type=DOUBLE, repetitiontype=OPTIONAL"`
	}
	awsTxOutput struct {
		Address             string  `parquet:"name=address, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Index               int64   `parquet:"name=index, type=INT64, repetitiontype=OPTIONAL"`
		Required_signatures int64   `parquet:"name=required_signatures, type=INT64, repetitiontype=OPTIONAL"`
		Script_asm          string  `parquet:"name=script_asm, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Script_hex          string  `parquet:"name=script_hex, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Type                string  `parquet:"name=type, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=OPTIONAL"`
		Value               float64 `parquet:"name=value, type=DOUBLE, repetitiontype=OPTIONAL"`
	}
)

func (a awsBlock) ToBlockHeader() (types.BlockHeader, error) {
	hash, err := chainhash.NewHashFromStr(a.Hash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert block hash")
	}
	prevBlockHash, err := chainhash.NewHashFromStr(a.PreviousBlockHash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert previous block hash")
	}
	merkleRoot, err := chainhash.NewHashFromStr(a.MerkleRoot)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert merkle root")
	}

	bits, err := strconv.ParseUint(a.Bits, 16, 32)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "can't convert bits from hex str to uint32")
	}

	return types.BlockHeader{
		Hash:       *hash,
		Height:     a.Number,
		Version:    int32(a.Version),
		PrevBlock:  *prevBlockHash,
		MerkleRoot: *merkleRoot,
		Timestamp:  parquettypes.INT96ToTime(a.Timestamp),
		Bits:       uint32(bits),
		Nonce:      uint32(a.Nonce),
	}, nil
}

func (a awsTransaction) ToTransaction(block awsBlock) (*types.Transaction, error) {
	blockhash, err := chainhash.NewHashFromStr(block.Hash)
	if err != nil {
		return nil, errors.Wrap(err, "can't convert block hash")
	}
	msgtx, err := a.MsgTx(block)
	if err != nil {
		return nil, errors.Wrap(err, "can't convert aws tx to wire.msgtx")
	}
	return types.ParseMsgTx(msgtx, a.BlockNumber, *blockhash, uint32(a.Index)), nil
}

func (a awsTransaction) MsgTx(block awsBlock) (*wire.MsgTx, error) {
	txIn := make([]*wire.TxIn, 0, len(a.Inputs))
	txOut := make([]*wire.TxOut, 0, len(a.Outputs))

	// coinbase tx from AWS S3 has no inputs, so we need to add it manually
	if a.IsCoinbase {
		scriptsig, err := hex.DecodeString(block.CoinbaseParam)
		if err != nil {
			return nil, errors.Wrap(err, "can't decode script hex")
		}

		txIn = append(txIn, &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  common.ZeroHash,
				Index: math.MaxUint32,
			},
			SignatureScript: scriptsig,
			Witness:         btcutils.CoinbaseWitness,
			Sequence:        0,
		})
	}

	for _, in := range a.Inputs {
		scriptsig, err := hex.DecodeString(in.ScriptHex)
		if err != nil {
			return nil, errors.Wrap(err, "can't decode script hex")
		}

		witness, err := btcutils.WitnessFromHex(lo.Map(in.TxInWitness, func(src *string, _ int) string {
			if src == nil {
				return ""
			}
			return *src
		}))
		if err != nil {
			return nil, errors.Wrap(err, "can't convert witness")
		}

		prevOutHash, err := chainhash.NewHashFromStr(in.SpentTransactionHash)
		if err != nil {
			return nil, errors.Wrap(err, "can't convert prevout hash")
		}

		txIn = append(txIn, &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  *prevOutHash,
				Index: uint32(in.SpentOutputIndex),
			},
			SignatureScript: scriptsig,
			Witness:         witness,
			Sequence:        uint32(in.Sequence),
		})
	}

	for _, out := range a.Outputs {
		scriptpubkey, err := hex.DecodeString(out.Script_hex)
		if err != nil {
			return nil, errors.Wrap(err, "can't decode script hex")
		}
		txOut = append(txOut, &wire.TxOut{
			Value:    btcutils.BitcoinToSatoshi(out.Value),
			PkScript: scriptpubkey,
		})
	}

	return &wire.MsgTx{
		Version:  int32(a.Version),
		TxIn:     txIn,
		TxOut:    txOut,
		LockTime: uint32(a.LockTime),
	}, nil
}
