package utils

import (
	"fmt"
	"math"
	"strconv"
)

// Uint32ToStr converts uint32 to string
func Uint32ToStr(val uint32) string {
	return strconv.FormatUint(uint64(val), 10)
}

// StrToUint32 converts string to uint32
func StrToUint32(val string) (uint32, error) {
	num, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to uint32: %w", err)
	}
	return uint32(num), nil
}

// IntToUint32 converts int to uint32
func IntToUint32(val int) (uint32, error) {
	if val < 0 {
		return 0, fmt.Errorf("negative value %d cannot be converted to uint32", val)
	}
	if val > math.MaxUint32 {
		return 0, fmt.Errorf("value %d exceeds maximum uint32 value", val)
	}
	return uint32(val), nil
}
