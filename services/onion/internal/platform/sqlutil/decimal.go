package sqlutil

import "fmt"

// FloatToDecimal formats a floating-point amount for a DECIMAL(…,2) database
// column using the same two-place representation as the POS repositories.
func FloatToDecimal(value float64) string {
	return fmt.Sprintf("%.2f", value)
}
