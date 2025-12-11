package rpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	pbbstream "github.com/streamingfast/bstream/pb/sf/bstream/v1"
	"github.com/xrpl-commons/firehose-xrpl/decoder"
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
	"github.com/xrpl-commons/firehose-xrpl/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// XRPL Epoch starts at January 1, 2000 (00:00 UTC)
// This is 946684800 seconds after Unix epoch (January 1, 1970)
const xrplEpochOffset = 946684800

// Buffer pools for reducing memory allocations
var (
	// bufferPool reuses byte buffers for hex decoding operations
	bufferPool = sync.Pool{
		New: func() interface{} {
			// Allocate a reasonable default size for XRPL transaction data
			// Most transactions are < 2KB, but we allocate 4KB to handle larger ones
			buf := make([]byte, 0, 4096)
			return &buf
		},
	}

	// bytesBufferPool reuses bytes.Buffer for general purpose operations
	bytesBufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// getBuffer retrieves a byte slice from the pool
func getBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

// putBuffer returns a byte slice to the pool after resetting it
func putBuffer(buf *[]byte) {
	// Reset the slice but keep the underlying capacity
	*buf = (*buf)[:0]
	bufferPool.Put(buf)
}

// getBytesBuffer retrieves a bytes.Buffer from the pool
func getBytesBuffer() *bytes.Buffer {
	return bytesBufferPool.Get().(*bytes.Buffer)
}

// putBytesBuffer returns a bytes.Buffer to the pool after resetting it
func putBytesBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bytesBufferPool.Put(buf)
}

// decodeHexWithPool decodes a hex string into a byte slice using pooled buffers
// The caller must copy the result if they need to retain it after the pool buffer is returned
func decodeHexWithPool(hexStr string) ([]byte, error) {
	// Calculate the required size
	expectedLen := hex.DecodedLen(len(hexStr))

	// Get a buffer from the pool
	buf := getBuffer()
	defer putBuffer(buf)

	// Ensure the buffer has enough capacity
	if cap(*buf) < expectedLen {
		*buf = make([]byte, expectedLen)
	} else {
		*buf = (*buf)[:expectedLen]
	}

	// Decode directly into the pooled buffer
	n, err := hex.Decode(*buf, []byte(hexStr))
	if err != nil {
		return nil, err
	}

	// Create a copy to return (the original buffer goes back to the pool)
	result := make([]byte, n)
	copy(result, (*buf)[:n])

	return result, nil
}

// LastBlockInfo tracks the latest fetched block information
type LastBlockInfo struct {
	blockNum uint64
}

// NewLastBlockInfo creates a new LastBlockInfo
func NewLastBlockInfo() *LastBlockInfo {
	return &LastBlockInfo{}
}

// Fetcher handles fetching XRPL ledgers and converting them to Firehose blocks
type Fetcher struct {
	fetchInterval            time.Duration
	latestBlockRetryInterval time.Duration
	lastBlockInfo            *LastBlockInfo
	decoder                  *decoder.Decoder
	workerPoolSize           int

	logger *zap.Logger
}

// NewFetcher creates a new XRPL ledger fetcher
func NewFetcher(fetchInterval, latestBlockRetryInterval time.Duration, logger *zap.Logger) *Fetcher {
	return &Fetcher{
		fetchInterval:            fetchInterval,
		latestBlockRetryInterval: latestBlockRetryInterval,
		lastBlockInfo:            NewLastBlockInfo(),
		decoder:                  decoder.NewDecoder(logger),
		workerPoolSize:           10, // Default worker pool size
		logger:                   logger,
	}
}

// NewFetcherWithWorkerPool creates a new XRPL ledger fetcher with custom worker pool size
func NewFetcherWithWorkerPool(fetchInterval, latestBlockRetryInterval time.Duration, workerPoolSize int, logger *zap.Logger) *Fetcher {
	return &Fetcher{
		fetchInterval:            fetchInterval,
		latestBlockRetryInterval: latestBlockRetryInterval,
		lastBlockInfo:            NewLastBlockInfo(),
		decoder:                  decoder.NewDecoder(logger),
		workerPoolSize:           workerPoolSize,
		logger:                   logger,
	}
}

// Fetch retrieves a ledger by number and converts it to a bstream Block
func (f *Fetcher) Fetch(ctx context.Context, client *Client, requestBlockNum uint64) (b *pbbstream.Block, skipped bool, err error) {
	// Add context with block number for better logging
	ctx = context.WithValue(ctx, "block_num", requestBlockNum)
	f.logger.Debug("starting fetch for block", zap.Uint64("block_num", requestBlockNum))
	// 1. Poll until the requested ledger is validated
	blockStartTime := time.Now()
	sleepDuration := time.Duration(0)
	for f.lastBlockInfo.blockNum < requestBlockNum {
		time.Sleep(sleepDuration)

		latestLedger, err := client.GetLatestLedger(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("fetching latest ledger: %w", err)
		}

		f.lastBlockInfo.blockNum = latestLedger.LedgerIndex
		f.logger.Info("got latest validated ledger",
			zap.Uint64("latest_ledger", f.lastBlockInfo.blockNum),
			zap.Uint64("requested_ledger", requestBlockNum))

		if f.lastBlockInfo.blockNum >= requestBlockNum {
			break
		}
		sleepDuration = f.latestBlockRetryInterval
	}

	// 2. Fetch the ledger with all transactions
	ledgerResult, err := client.GetLedger(ctx, requestBlockNum)
	if err != nil {
		return nil, false, fmt.Errorf("fetching ledger %d: %w", requestBlockNum, err)
	}
	ledger := ledgerResult.Ledger

	// 3. Build transactions from the ledger data using parallel processing
	transactions := make([]*pbxrpl.Transaction, len(ledger.Transactions))
	var wg sync.WaitGroup
	errChan := make(chan error, len(ledger.Transactions))

	// Use worker pool pattern for parallel processing
	workerCount := f.workerPoolSize
	if len(ledger.Transactions) < workerCount {
		workerCount = len(ledger.Transactions)
	}
	if workerCount == 0 {
		workerCount = 1 // Ensure at least one worker
	}

	txChan := make(chan struct {
		index int
		tx    types.LedgerTransaction
	}, len(ledger.Transactions))

	// Start worker pool
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for txData := range txChan {
				i, tx := txData.index, txData.tx

				// Decode hash using pooled buffers
				txHash, err := decodeHexWithPool(tx.Hash)
				if err != nil {
					errChan <- fmt.Errorf("decoding tx hash at index %d: %w", i, err)
					continue
				}

				// Decode tx_blob (binary transaction) using pooled buffers
				txBlob, err := decodeHexWithPool(tx.TxBlob)
				if err != nil {
					errChan <- fmt.Errorf("decoding tx blob at index %d: %w", i, err)
					continue
				}

				// Decode meta (binary metadata) using pooled buffers
				metaBlob, err := decodeHexWithPool(tx.Meta)
				if err != nil {
					errChan <- fmt.Errorf("decoding meta blob at index %d: %w", i, err)
					continue
				}

				// Use decoder to map transaction to protobuf (includes all fields and tx_details)
				protoTx, err := f.decoder.MapTransactionToProto(txBlob, metaBlob, txHash, uint32(i))
				if err != nil {
					f.logger.Warn("failed to map transaction to protobuf, skipping",
						zap.Int("tx_index", i),
						zap.String("tx_hash", tx.Hash),
						zap.Error(err))
					continue
				}

				transactions[i] = protoTx
			}
		}()
	}

	// Feed transactions to workers
	for i, tx := range ledger.Transactions {
		txChan <- struct {
			index int
			tx    types.LedgerTransaction
		}{
			index: i,
			tx:    tx,
		}
	}
	close(txChan)

	// Wait for all workers to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	if len(errChan) > 0 {
		return nil, false, <-errChan
	}

	// Filter out nil transactions (failed mappings)
	var validTransactions []*pbxrpl.Transaction
	for _, tx := range transactions {
		if tx != nil {
			validTransactions = append(validTransactions, tx)
		}
	}
	transactions = validTransactions

	// 4. Build the block header with parallel hex decoding using pooled buffers
	var ledgerHash, parentHash, accountHash, transactionHash []byte
	var decodeWg sync.WaitGroup
	var decodeErr error
	var decodeErrMutex sync.Mutex

	decodeWg.Add(4)

	go func() {
		defer decodeWg.Done()
		var err error
		ledgerHash, err = decodeHexWithPool(ledger.LedgerHash)
		if err != nil {
			decodeErrMutex.Lock()
			if decodeErr == nil {
				decodeErr = fmt.Errorf("decoding ledger hash: %w", err)
			}
			decodeErrMutex.Unlock()
		}
	}()

	go func() {
		defer decodeWg.Done()
		var err error
		parentHash, err = decodeHexWithPool(ledger.ParentHash)
		if err != nil {
			decodeErrMutex.Lock()
			if decodeErr == nil {
				decodeErr = fmt.Errorf("decoding parent hash: %w", err)
			}
			decodeErrMutex.Unlock()
		}
	}()

	go func() {
		defer decodeWg.Done()
		var err error
		accountHash, err = decodeHexWithPool(ledger.AccountHash)
		if err != nil {
			f.logger.Debug("failed to decode account hash", zap.Error(err))
			accountHash = nil
		}
	}()

	go func() {
		defer decodeWg.Done()
		var err error
		transactionHash, err = decodeHexWithPool(ledger.TransactionHash)
		if err != nil {
			f.logger.Debug("failed to decode transaction hash", zap.Error(err))
			transactionHash = nil
		}
	}()

	decodeWg.Wait()

	if decodeErr != nil {
		return nil, false, decodeErr
	}

	// Parse total coins (drops)
	totalDrops, err := strconv.ParseInt(ledger.TotalCoins, 10, 64)
	if err != nil {
		f.logger.Warn("failed to parse total_coins", zap.String("total_coins", ledger.TotalCoins), zap.Error(err))
		totalDrops = 0
	}

	// Convert XRPL epoch time to Unix time
	closeTime := xrplEpochToTime(ledger.CloseTime)

	// 5. Build the XRPL Block protobuf
	xrplBlock := &pbxrpl.Block{
		Number: ledger.LedgerIndex,
		Hash:   ledgerHash,
		Header: &pbxrpl.Header{
			ParentHash:          parentHash,
			TotalDrops:          totalDrops,
			AccountHash:         accountHash,
			TransactionHash:     transactionHash,
			CloseTimeResolution: ledger.CloseTimeResolution,
			CloseFlags:          ledger.CloseFlags,
		},
		Version:      1,
		Transactions: transactions,
		CloseTime:    timestamppb.New(closeTime),
	}

	// 6. Convert to bstream block
	bstreamBlock, err := convertBlock(xrplBlock)
	if err != nil {
		return nil, false, fmt.Errorf("converting block: %w", err)
	}

	f.logger.Info("fetched ledger",
		zap.Uint64("ledger_index", ledger.LedgerIndex),
		zap.Int("tx_count", len(transactions)),
		zap.Time("close_time", closeTime),
		zap.Duration("processing_time", time.Since(blockStartTime)))

	return bstreamBlock, false, nil
}

// Add performance monitoring variables
var (
	blocksProcessed       int
	transactionsProcessed int
	startTime             = time.Now()
)

// GetPerformanceMetrics returns performance statistics

// IsBlockAvailable checks if a block number is available
func (f *Fetcher) IsBlockAvailable(blockNum uint64) bool {
	return blockNum <= f.lastBlockInfo.blockNum
}

// FetchBatch retrieves multiple ledgers in parallel and converts them to bstream Blocks
func (f *Fetcher) FetchBatch(ctx context.Context, client *Client, requestBlockNums []uint64) ([]*pbbstream.Block, error) {
	if len(requestBlockNums) == 0 {
		return nil, nil
	}

	// Create a context for the batch operation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use a worker pool for parallel block fetching
	blocks := make([]*pbbstream.Block, len(requestBlockNums))
	var wg sync.WaitGroup
	errChan := make(chan error, len(requestBlockNums))

	// Limit concurrent block fetches to avoid overwhelming the RPC endpoint
	concurrencyLimit := 5
	if len(requestBlockNums) < concurrencyLimit {
		concurrencyLimit = len(requestBlockNums)
	}

	semaphore := make(chan struct{}, concurrencyLimit)

	for i, blockNum := range requestBlockNums {
		wg.Add(1)
		go func(idx int, num uint64) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Fetch individual block
			block, _, err := f.Fetch(ctx, client, num)
			if err != nil {
				errChan <- fmt.Errorf("failed to fetch block %d: %w", num, err)
				return
			}

			blocks[idx] = block
		}(i, blockNum)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return blocks, nil
}

// xrplEpochToTime converts XRPL epoch seconds to Go time.Time
// XRPL epoch starts at 2000-01-01 00:00:00 UTC
func xrplEpochToTime(xrplTime uint64) time.Time {
	unixTime := int64(xrplTime) + xrplEpochOffset
	return time.Unix(unixTime, 0).UTC()
}

// convertBlock converts an XRPL Block to a bstream Block
func convertBlock(xrplBlk *pbxrpl.Block) (*pbbstream.Block, error) {
	anyBlock, err := anypb.New(xrplBlk)
	if err != nil {
		return nil, fmt.Errorf("unable to create anypb: %w", err)
	}

	// Use hex encoding for block IDs (standard for XRPL)
	blockHash := hex.EncodeToString(xrplBlk.Hash)
	parentHash := hex.EncodeToString(xrplBlk.Header.ParentHash)

	return &pbbstream.Block{
		Number:    xrplBlk.Number,
		Id:        blockHash,
		ParentId:  parentHash,
		Timestamp: xrplBlk.CloseTime,
		LibNum:    xrplBlk.Number - 1, // Every validated ledger in XRPL is final
		ParentNum: xrplBlk.Number - 1,
		Payload:   anyBlock,
	}, nil
}
