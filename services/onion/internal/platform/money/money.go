package money

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
)

var ErrInvalidAmount = errors.New("invalid monetary amount")

// DollarsToCents converts a float64 dollar amount from an external contract
// into integer cents, rounding to the nearest cent.
func DollarsToCents(dollars float64) int64 {
	return int64(math.Round(dollars * 100))
}

// CentsToDollars converts an internally stored integer cents amount to the
// float64 dollar shape used by existing external contracts.
func CentsToDollars(cents int64) float64 {
	return float64(cents) / 100
}

// Amount stores money in the smallest currency unit (normally cents).
// Keeping the value as an integer avoids floating-point rounding drift while
// custom JSON/SQL encodings preserve the existing decimal API shape.
type Amount int64

func FromString(value string) (Amount, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return 0, ErrInvalidAmount
	}
	if strings.ContainsAny(s, "eE") {
		return 0, ErrInvalidAmount
	}

	sign := int64(1)
	if s[0] == '-' || s[0] == '+' {
		if s[0] == '-' {
			sign = -1
		}
		s = s[1:]
	}
	if s == "" {
		return 0, ErrInvalidAmount
	}

	whole, fraction := s, ""
	hasDecimal := false
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		hasDecimal = true
		whole, fraction = s[:dot], s[dot+1:]
		if strings.Contains(fraction, ".") || len(fraction) > 2 {
			return 0, ErrInvalidAmount
		}
		if fraction == "" {
			return 0, ErrInvalidAmount
		}
	}
	if whole == "" {
		whole = "0"
	}
	if !isDecimalDigits(whole) {
		return 0, ErrInvalidAmount
	}
	if !hasDecimal {
		fraction = "00"
	} else if len(fraction) == 1 {
		fraction += "0"
	}
	if !isDecimalDigits(fraction) {
		return 0, ErrInvalidAmount
	}
	wholeValue, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || wholeValue < 0 {
		return 0, ErrInvalidAmount
	}
	fractionValue, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}
	if wholeValue > math.MaxInt64/100 {
		return 0, ErrInvalidAmount
	}
	cents := wholeValue*100 + fractionValue
	// wholeValue is capped at MaxInt64/100 above; if the addition wraps it
	// becomes negative and we reject it here.
	if cents < 0 {
		return 0, ErrInvalidAmount
	}
	if sign < 0 {
		return Amount(-cents), nil
	}
	return Amount(cents), nil
}

func isDecimalDigits(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func FromFloat64(value float64) (Amount, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrInvalidAmount
	}
	return FromString(strconv.FormatFloat(value, 'f', -1, 64))
}

func (a Amount) String() string {
	value := int64(a)
	if value < 0 {
		return "-" + formatCents(-value)
	}
	return formatCents(value)
}

func formatCents(value int64) string {
	return strconv.FormatInt(value/100, 10) + "." + leftPad2(strconv.FormatInt(value%100, 10))
}

func leftPad2(value string) string {
	if len(value) == 1 {
		return "0" + value
	}
	return value
}

func (a Amount) Float64() float64 {
	return float64(a) / 100
}

func (a Amount) IsZero() bool {
	return a == 0
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = 0
		return nil
	}
	var value string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
	} else {
		value = string(data)
	}
	parsed, err := FromString(value)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

func (a *Amount) Scan(value any) error {
	if value == nil {
		*a = 0
		return nil
	}
	switch v := value.(type) {
	case []byte:
		parsed, err := FromString(string(v))
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	case string:
		parsed, err := FromString(v)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	case int64:
		*a = Amount(v * 100)
		return nil
	case float64:
		parsed, err := FromFloat64(v)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	default:
		return ErrInvalidAmount
	}
}

func (a Amount) Value() (driver.Value, error) {
	return a.String(), nil
}
