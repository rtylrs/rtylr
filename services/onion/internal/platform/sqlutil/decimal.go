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

// FloatToDecimal formats a floating-point amount for a DECIMAL(…,2) database
// column using the same two-place representation as the POS repositories.
func FloatToDecimal(value float64) string {
	return fmt.Sprintf("%.2f", value)
}
