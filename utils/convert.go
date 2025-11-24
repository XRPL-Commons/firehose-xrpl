package utils

import (
	"strconv"
	"time"

	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
)

// XRPL Epoch starts at January 1, 2000 (00:00 UTC)
const XRPLEpochOffset int64 = 946684800

// XRPLEpochToTime converts XRPL epoch seconds to Go time.Time
func XRPLEpochToTime(xrplTime uint64) time.Time {
	unixTime := int64(xrplTime) + XRPLEpochOffset
	return time.Unix(unixTime, 0).UTC()
}

// TimeToXRPLEpoch converts Go time.Time to XRPL epoch seconds
func TimeToXRPLEpoch(t time.Time) uint64 {
	return uint64(t.Unix() - XRPLEpochOffset)
}

// DropsToXRP converts drops to XRP (1 XRP = 1,000,000 drops)
func DropsToXRP(drops uint64) float64 {
	return float64(drops) / 1000000.0
}

// XRPToDrops converts XRP to drops
func XRPToDrops(xrp float64) uint64 {
	return uint64(xrp * 1000000.0)
}

// ParseDrops parses a drops string to uint64
func ParseDrops(drops string) (uint64, error) {
	return strconv.ParseUint(drops, 10, 64)
}

// ConvertTransactionResult converts XRPL result string to protobuf enum
func ConvertTransactionResult(result string) pbxrpl.TransactionResult {
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
	default:
		// Check prefix for categorization
		if len(result) >= 3 {
			switch result[:3] {
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

// IsSuccessResult checks if the transaction result indicates success
func IsSuccessResult(result pbxrpl.TransactionResult) bool {
	return result == pbxrpl.TransactionResult_TES_SUCCESS
}

// IsClaimedResult checks if the transaction result indicates fee was claimed but tx failed
func IsClaimedResult(result pbxrpl.TransactionResult) bool {
	return result >= pbxrpl.TransactionResult_TEC_CLAIMED &&
		result <= pbxrpl.TransactionResult_TEC_OTHER
}
