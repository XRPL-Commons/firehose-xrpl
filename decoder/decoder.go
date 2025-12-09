package decoder

import (
	"encoding/hex"
	"fmt"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	xrpltx "github.com/Peersyst/xrpl-go/xrpl/transaction"
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
	"go.uber.org/zap"
)

// Decoder handles XRPL binary format decoding using xrpl-go's binarycodec
type Decoder struct {
	logger *zap.Logger
	mapper *Mapper
}

// NewDecoder creates a new XRPL decoder
func NewDecoder(logger *zap.Logger) *Decoder {
	return &Decoder{
		logger: logger,
		mapper: NewMapper(logger),
	}
}

// DecodeTransactionFromHex decodes a transaction blob (hex string) to a FlatTransaction
func (d *Decoder) DecodeTransactionFromHex(txBlobHex string) (xrpltx.FlatTransaction, error) {
	decoded, err := binarycodec.Decode(txBlobHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction blob: %w", err)
	}

	return xrpltx.FlatTransaction(decoded), nil
}

// DecodeTransactionFromBytes decodes a transaction from raw bytes
func (d *Decoder) DecodeTransactionFromBytes(txBlob []byte) (xrpltx.FlatTransaction, error) {
	hexStr := hex.EncodeToString(txBlob)
	return d.DecodeTransactionFromHex(hexStr)
}

// DecodeMetadataFromHex decodes transaction metadata (hex string)
func (d *Decoder) DecodeMetadataFromHex(metaHex string) (map[string]interface{}, error) {
	decoded, err := binarycodec.Decode(metaHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return decoded, nil
}

// DecodeMetadataFromBytes decodes metadata from raw bytes
func (d *Decoder) DecodeMetadataFromBytes(metaBlob []byte) (map[string]interface{}, error) {
	hexStr := hex.EncodeToString(metaBlob)
	return d.DecodeMetadataFromHex(hexStr)
}

// GetTransactionType extracts the transaction type string from a tx blob
func (d *Decoder) GetTransactionType(txBlob []byte) string {
	decoded, err := d.DecodeTransactionFromBytes(txBlob)
	if err != nil {
		d.logger.Debug("failed to decode transaction for type extraction", zap.Error(err))
		return ""
	}

	if txType, ok := decoded["TransactionType"].(string); ok {
		return txType
	}

	return ""
}

// GetTransactionResult extracts the result code string from metadata
func (d *Decoder) GetTransactionResult(metaBlob []byte) string {
	decoded, err := d.DecodeMetadataFromBytes(metaBlob)
	if err != nil {
		d.logger.Debug("failed to decode metadata for result extraction", zap.Error(err))
		return ""
	}

	if result, ok := decoded["TransactionResult"].(string); ok {
		return result
	}

	return ""
}

// MapTransactionToProto converts a decoded FlatTransaction and metadata to protobuf
// This is the main entry point used by the fetcher
func (d *Decoder) MapTransactionToProto(txBlob, metaBlob []byte, txHash []byte, txIndex uint32) (*pbxrpl.Transaction, error) {
	// Decode the transaction
	flatTx, err := d.DecodeTransactionFromBytes(txBlob)
	if err != nil {
		return nil, fmt.Errorf("decoding transaction: %w", err)
	}

	// Decode the metadata to get the result
	meta, err := d.DecodeMetadataFromBytes(metaBlob)
	if err != nil {
		return nil, fmt.Errorf("decoding metadata: %w", err)
	}

	result := ""
	if txResult, ok := meta["TransactionResult"].(string); ok {
		result = txResult
	}

	// Use the mapper to convert to protobuf
	return d.mapper.MapTransactionToProto(flatTx, txBlob, metaBlob, txHash, txIndex, result)
}
