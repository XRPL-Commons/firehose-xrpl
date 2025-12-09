package utils

import (
	"strconv"
	"time"
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

// IsSuccessResult checks if the transaction result indicates success
func IsSuccessResult(result string) bool {
	return result == "tesSUCCESS"
}

// IsClaimedResult checks if the transaction result indicates fee was claimed but tx failed
func IsClaimedResult(result string) bool {
	return len(result) >= 3 && result[:3] == "tec"
}

// GetResultCategory returns the category of a transaction result (tes, tec, tef, tem, ter)
func GetResultCategory(result string) string {
	if len(result) >= 3 {
		return result[:3]
	}
	return ""
}
