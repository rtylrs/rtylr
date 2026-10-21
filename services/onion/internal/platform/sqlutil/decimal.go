package sqlutil

import (
	"fmt"
	"strconv"
)

// DecimalToFloat converts the string or byte representation returned for a
// DECIMAL column into the repository's floating-point representation. Empty,
// unsupported, and malformed values retain the existing zero-value fallback.
func DecimalToFloat(value any) float64 {
	var text string
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	case uint:
		return float64(typed)
	case uint32:
		return float64(typed)
	case uint64:
		return float64(typed)
	case string:
		text = typed
	case []byte:
		text = string(typed)
	default:
		return 0
	}
	if text == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}
	return parsed
}

// ParseDecimal parses the string representation used by non-POS DECIMAL
// columns. Invalid or empty values retain the repository zero-value fallback.
func ParseDecimal(value string) float64 {
	return DecimalToFloat(value)
}

// FloatToDecimal formats a floating-point amount for a DECIMAL(…,2) database
// column using the same two-place representation as the POS repositories.
func FloatToDecimal(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// DecimalToMinorUnits converts a DECIMAL(…,2) column value into integer minor
// units, rounding half up. Empty and malformed values return zero.
func DecimalToMinorUnits(value any) int64 {
	return int64(DecimalToFloat(value)*100 + 0.5)
}
