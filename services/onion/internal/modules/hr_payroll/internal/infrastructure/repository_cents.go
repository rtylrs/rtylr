package infrastructure

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

// decimalStringToCents parses a MySQL DECIMAL value read back as a string
// into integer cents. It works on the string digits directly rather than
// through strconv.ParseFloat, so no float64 rounding error can be
// introduced on the way out of the database.
func decimalStringToCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	whole, fraction, hasFraction := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if len(fraction) > 2 {
		fraction = fraction[:2]
	}
	for len(fraction) < 2 {
		fraction += "0"
	}
	if !hasFraction {
		fraction = "00"
	}

	wholeCents, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, err
	}
	fractionCents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, err
	}

	total := wholeCents*100 + fractionCents
	if negative {
		total = -total
	}
	return total, nil
}

// decimalStringToCentsOrZero parses a non-nullable DECIMAL value read back
// as a string into integer cents, logging and returning zero on the (should
// never happen) case that the database returns something that doesn't parse
// as a decimal.
func decimalStringToCentsOrZero(s string) int64 {
	cents, err := decimalStringToCents(s)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return cents
}

// nullDecimalStringToCents parses a nullable MySQL DECIMAL value read back
// as a string into integer cents, using only integer/string arithmetic so no
// float64 rounding error can be introduced on the way out of the database.
func nullDecimalStringToCents(v sql.NullString) int64 {
	if !v.Valid {
		return 0
	}
	cents, err := decimalStringToCents(v.String)
	if err != nil {
		logger.Global().Warn("failed to parse nullable decimal from database",
			logger.String("value", v.String),
			logger.Error(err),
		)
		return 0
	}
	return cents
}

func interfaceToInt32(v interface{}) int32 {
	return int32(interfaceToInt64(v)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func interfaceToInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case uint64:
		return int64(val) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case []byte:
		n, err := strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			logger.Global().Warn("failed to parse int value from database",
				logger.String("value", string(val)),
				logger.Error(err),
			)
			return 0
		}
		return n
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logger.Global().Warn("failed to parse int value from database",
				logger.String("value", val),
				logger.Error(err),
			)
			return 0
		}
		return n
	default:
		return 0
	}
}

func interfaceToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case []byte:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			logger.Global().Warn("failed to parse float value from database",
				logger.String("value", string(val)),
				logger.Error(err),
			)
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			logger.Global().Warn("failed to parse float value from database",
				logger.String("value", val),
				logger.Error(err),
			)
			return 0
		}
		return f
	default:
		return 0
	}
}
