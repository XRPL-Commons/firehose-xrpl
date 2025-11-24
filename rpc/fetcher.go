package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	pbbstream "github.com/streamingfast/bstream/pb/sf/bstream/v1"
	"github.com/xrpl-commons/firehose-xrpl/decoder"
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// XRPL Epoch starts at January 1, 2000 (00:00 UTC)
// This is 946684800 seconds after Unix epoch (January 1, 1970)
const xrplEpochOffset = 946684800

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

	logger *zap.Logger
}

// NewFetcher creates a new XRPL ledger fetcher
func NewFetcher(fetchInterval, latestBlockRetryInterval time.Duration, logger *zap.Logger) *Fetcher {
	return &Fetcher{
		fetchInterval:            fetchInterval,
		latestBlockRetryInterval: latestBlockRetryInterval,
		lastBlockInfo:            NewLastBlockInfo(),
		decoder:                  decoder.NewDecoder(logger),
		logger:                   logger,
	}
}

// Fetch retrieves a ledger by number and converts it to a bstream Block
func (f *Fetcher) Fetch(ctx context.Context, client *Client, requestBlockNum uint64) (b *pbbstream.Block, skipped bool, err error) {
	// 1. Poll until the requested ledger is validated
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

	// 2. Fetch the ledger with all transactions (single call - simpler than Stellar!)
	ledgerResult, err := client.GetLedger(ctx, requestBlockNum)
	if err != nil {
		return nil, false, fmt.Errorf("fetching ledger %d: %w", requestBlockNum, err)
	}

	ledger := ledgerResult.Ledger

	// 3. Build transactions from the ledger data
	transactions := make([]*pbxrpl.Transaction, 0, len(ledger.Transactions))
	for i, tx := range ledger.Transactions {
		// Decode hash
		txHash, err := hex.DecodeString(tx.Hash)
		if err != nil {
			return nil, false, fmt.Errorf("decoding tx hash: %w", err)
		}

		// Decode tx_blob (binary transaction)
		txBlob, err := hex.DecodeString(tx.TxBlob)
		if err != nil {
			return nil, false, fmt.Errorf("decoding tx blob: %w", err)
		}

		// Decode meta (binary metadata)
		metaBlob, err := hex.DecodeString(tx.Meta)
		if err != nil {
			return nil, false, fmt.Errorf("decoding meta blob: %w", err)
		}

		// Extract transaction type and result using decoder
		txType := f.decoder.GetTransactionType(txBlob)
		result := f.decoder.GetTransactionResult(metaBlob)

		// Try to extract additional info from the transaction
		var account string
		var fee uint64
		var sequence uint32

		txInfo, decodeErr := f.decoder.ExtractTransactionInfo(txBlob, metaBlob)
		if decodeErr == nil {
			account = txInfo.Account
			fee = txInfo.Fee
			sequence = txInfo.Sequence
		} else {
			f.logger.Debug("failed to extract full transaction info",
				zap.String("tx_hash", tx.Hash),
				zap.Error(decodeErr))
		}

		transactions = append(transactions, &pbxrpl.Transaction{
			Hash:     txHash,
			Result:   result,
			Index:    uint32(i),
			TxBlob:   txBlob,
			MetaBlob: metaBlob,
			TxType:   txType,
			Account:  account,
			Fee:      fee,
			Sequence: sequence,
		})
	}

	// 4. Build the block header
	ledgerHash, err := hex.DecodeString(ledger.LedgerHash)
	if err != nil {
		return nil, false, fmt.Errorf("decoding ledger hash: %w", err)
	}

	parentHash, err := hex.DecodeString(ledger.ParentHash)
	if err != nil {
		return nil, false, fmt.Errorf("decoding parent hash: %w", err)
	}

	accountHash, err := hex.DecodeString(ledger.AccountHash)
	if err != nil {
		f.logger.Debug("failed to decode account hash", zap.Error(err))
		accountHash = nil
	}

	transactionHash, err := hex.DecodeString(ledger.TransactionHash)
	if err != nil {
		f.logger.Debug("failed to decode transaction hash", zap.Error(err))
		transactionHash = nil
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
			// Note: base_fee, reserve_base, reserve_increment would need server_info call
			// or could be extracted from fee settings in ledger if available
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
		zap.Time("close_time", closeTime))

	return bstreamBlock, false, nil
}

// IsBlockAvailable checks if a block number is available
func (f *Fetcher) IsBlockAvailable(blockNum uint64) bool {
	return blockNum <= f.lastBlockInfo.blockNum
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
