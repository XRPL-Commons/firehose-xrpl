package decoder

import (
	"encoding/hex"
	"fmt"
	"strconv"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
	"github.com/xrpl-commons/firehose-xrpl/types"
	"go.uber.org/zap"
)

// Decoder handles XRPL binary format decoding using xrpl-go's binarycodec
type Decoder struct {
	logger *zap.Logger
}

// NewDecoder creates a new XRPL decoder
func NewDecoder(logger *zap.Logger) *Decoder {
	return &Decoder{
		logger: logger,
	}
}

// DecodedTransaction holds the decoded transaction data
type DecodedTransaction struct {
	TransactionType string
	Account         string
	Destination     string
	Fee             uint64
	Sequence        uint32
	Hash            string
	// Raw decoded JSON for additional fields
	RawJSON map[string]interface{}
}

// DecodeTransaction decodes a transaction blob (hex string) to structured data
func (d *Decoder) DecodeTransaction(txBlobHex string) (*DecodedTransaction, error) {
	// Use xrpl-go's binarycodec to decode the transaction
	decoded, err := binarycodec.Decode(txBlobHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction blob: %w", err)
	}

	result := &DecodedTransaction{
		RawJSON: decoded,
	}

	// Extract common fields
	if txType, ok := decoded["TransactionType"].(string); ok {
		result.TransactionType = txType
	}

	if account, ok := decoded["Account"].(string); ok {
		result.Account = account
	}

	if destination, ok := decoded["Destination"].(string); ok {
		result.Destination = destination
	}

	if fee, ok := decoded["Fee"].(string); ok {
		if feeVal, err := strconv.ParseUint(fee, 10, 64); err == nil {
			result.Fee = feeVal
		}
	}

	if seq, ok := decoded["Sequence"]; ok {
		switch s := seq.(type) {
		case float64:
			result.Sequence = uint32(s)
		case int:
			result.Sequence = uint32(s)
		case int64:
			result.Sequence = uint32(s)
		}
	}

	return result, nil
}

// DecodeTransactionFromBytes decodes a transaction from raw bytes
func (d *Decoder) DecodeTransactionFromBytes(txBlob []byte) (*DecodedTransaction, error) {
	hexStr := hex.EncodeToString(txBlob)
	return d.DecodeTransaction(hexStr)
}

// DecodedMetadata holds decoded transaction metadata
type DecodedMetadata struct {
	TransactionResult string
	TransactionIndex  uint32
	AffectedNodes     []AffectedNode
	DeliveredAmount   string
	// Raw decoded JSON
	RawJSON map[string]interface{}
}

// AffectedNode represents a node affected by the transaction
type AffectedNode struct {
	NodeType    string // "CreatedNode", "ModifiedNode", "DeletedNode"
	LedgerEntry string // Entry type like "AccountRoot", "RippleState", etc.
	LedgerIndex string // The ledger index of the affected node
}

// DecodeMetadata decodes transaction metadata (hex string)
func (d *Decoder) DecodeMetadata(metaHex string) (*DecodedMetadata, error) {
	// Use xrpl-go's binarycodec to decode ledger/metadata
	decoded, err := binarycodec.Decode(metaHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	result := &DecodedMetadata{
		RawJSON: decoded,
	}

	// Extract transaction result
	if txResult, ok := decoded["TransactionResult"].(string); ok {
		result.TransactionResult = txResult
	}

	// Extract transaction index
	if txIndex, ok := decoded["TransactionIndex"]; ok {
		switch idx := txIndex.(type) {
		case float64:
			result.TransactionIndex = uint32(idx)
		case int:
			result.TransactionIndex = uint32(idx)
		}
	}

	// Extract delivered amount if present
	if deliveredAmt, ok := decoded["delivered_amount"]; ok {
		switch amt := deliveredAmt.(type) {
		case string:
			result.DeliveredAmount = amt
		case map[string]interface{}:
			// IOU amount
			if value, ok := amt["value"].(string); ok {
				result.DeliveredAmount = value
			}
		}
	}

	// Extract affected nodes
	if affectedNodes, ok := decoded["AffectedNodes"].([]interface{}); ok {
		for _, node := range affectedNodes {
			if nodeMap, ok := node.(map[string]interface{}); ok {
				an := AffectedNode{}

				for nodeType, nodeData := range nodeMap {
					an.NodeType = nodeType
					if data, ok := nodeData.(map[string]interface{}); ok {
						if ledgerEntry, ok := data["LedgerEntryType"].(string); ok {
							an.LedgerEntry = ledgerEntry
						}
						if ledgerIndex, ok := data["LedgerIndex"].(string); ok {
							an.LedgerIndex = ledgerIndex
						}
					}
					break // Only one key per node
				}

				result.AffectedNodes = append(result.AffectedNodes, an)
			}
		}
	}

	return result, nil
}

// DecodeMetadataFromBytes decodes metadata from raw bytes
func (d *Decoder) DecodeMetadataFromBytes(metaBlob []byte) (*DecodedMetadata, error) {
	hexStr := hex.EncodeToString(metaBlob)
	return d.DecodeMetadata(hexStr)
}

// GetTransactionType extracts the transaction type from a tx blob
func (d *Decoder) GetTransactionType(txBlob []byte) pbxrpl.TransactionType {
	decoded, err := d.DecodeTransactionFromBytes(txBlob)
	if err != nil {
		d.logger.Debug("failed to decode transaction for type extraction", zap.Error(err))
		return pbxrpl.TransactionType_TX_UNKNOWN
	}

	return types.TxTypeStringToProto(decoded.TransactionType)
}

// GetTransactionResult extracts the result code from metadata
func (d *Decoder) GetTransactionResult(metaBlob []byte) pbxrpl.TransactionResult {
	decoded, err := d.DecodeMetadataFromBytes(metaBlob)
	if err != nil {
		d.logger.Debug("failed to decode metadata for result extraction", zap.Error(err))
		return pbxrpl.TransactionResult_RESULT_UNKNOWN
	}

	return ResultStringToProto(decoded.TransactionResult)
}

// ResultStringToProto converts XRPL result string to protobuf enum
func ResultStringToProto(result string) pbxrpl.TransactionResult {
	switch result {
	case "tesSUCCESS":
		return pbxrpl.TransactionResult_TES_SUCCESS
	case "tecCLAIMED":
		return pbxrpl.TransactionResult_TEC_CLAIMED
	case "tecPATH_PARTIAL":
		return pbxrpl.TransactionResult_TEC_PATH_PARTIAL
	case "tecUNFUNDED_ADD":
		return pbxrpl.TransactionResult_TEC_UNFUNDED_ADD
	case "tecUNFUNDED_OFFER":
		return pbxrpl.TransactionResult_TEC_UNFUNDED_OFFER
	case "tecUNFUNDED_PAYMENT":
		return pbxrpl.TransactionResult_TEC_UNFUNDED_PAYMENT
	case "tecFAILED_PROCESSING":
		return pbxrpl.TransactionResult_TEC_FAILED_PROCESSING
	case "tecDIR_FULL":
		return pbxrpl.TransactionResult_TEC_DIR_FULL
	case "tecINSUF_RESERVE_LINE":
		return pbxrpl.TransactionResult_TEC_INSUF_RESERVE_LINE
	case "tecINSUF_RESERVE_OFFER":
		return pbxrpl.TransactionResult_TEC_INSUF_RESERVE_OFFER
	case "tecNO_DST":
		return pbxrpl.TransactionResult_TEC_NO_DST
	case "tecNO_DST_INSUF_XRP":
		return pbxrpl.TransactionResult_TEC_NO_DST_INSUF_XRP
	case "tecNO_LINE_INSUF_RESERVE":
		return pbxrpl.TransactionResult_TEC_NO_LINE_INSUF_RESERVE
	case "tecNO_LINE_REDUNDANT":
		return pbxrpl.TransactionResult_TEC_NO_LINE_REDUNDANT
	case "tecPATH_DRY":
		return pbxrpl.TransactionResult_TEC_PATH_DRY
	case "tecUNFUNDED":
		return pbxrpl.TransactionResult_TEC_UNFUNDED
	case "tecNO_ALTERNATIVE_KEY":
		return pbxrpl.TransactionResult_TEC_NO_ALTERNATIVE_KEY
	case "tecNO_REGULAR_KEY":
		return pbxrpl.TransactionResult_TEC_NO_REGULAR_KEY
	default:
		// Check prefixes for categorization
		if len(result) >= 3 {
			prefix := result[:3]
			switch prefix {
			case "tec":
				return pbxrpl.TransactionResult_TEC_OTHER
			case "tef":
				return pbxrpl.TransactionResult_TEF_FAILURE
			case "tem":
				return pbxrpl.TransactionResult_TEM_MALFORMED
			case "ter":
				return pbxrpl.TransactionResult_TER_RETRY
			}
		}
		return pbxrpl.TransactionResult_RESULT_UNKNOWN
	}
}

// ExtractTransactionInfo extracts key transaction info for the protobuf
func (d *Decoder) ExtractTransactionInfo(txBlob, metaBlob []byte) (*types.TransactionMeta, error) {
	txDecoded, err := d.DecodeTransactionFromBytes(txBlob)
	if err != nil {
		return nil, fmt.Errorf("decoding transaction: %w", err)
	}

	metaDecoded, err := d.DecodeMetadataFromBytes(metaBlob)
	if err != nil {
		return nil, fmt.Errorf("decoding metadata: %w", err)
	}

	txType := types.TxTypeStringToProto(txDecoded.TransactionType)
	result := ResultStringToProto(metaDecoded.TransactionResult)

	return types.NewTransactionMeta(
		nil, // hash will be set by caller
		txBlob,
		metaBlob,
		txType,
		result,
		txDecoded.Account,
		txDecoded.Fee,
		txDecoded.Sequence,
	), nil
}
