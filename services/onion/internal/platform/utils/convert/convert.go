package convert

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}

	return ""
}

func Int32ToString(i int32) string {
	return strconv.Itoa(int(i))
}

func IntToString(i int) string {
	return strconv.Itoa(i)
}

func ToBool(b pgtype.Bool) bool {
	if b.Valid {
		return b.Bool
	}

	return false
}

func ToInt32(i pgtype.Int4) int32 {
	if i.Valid {
		return i.Int32
	}

	return 0
}

func ToStringSlice(s []string) []string {
	return s
}

func ToTimestamp(t pgtype.Timestamptz) pgtype.Timestamptz {
	return t
}

func BoolToString(b bool) string {
	if b {
		return "1"
	}

	return "0"
}

func ToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0.0
	}

	if priceFloat8, err := n.Float64Value(); err == nil && priceFloat8.Valid {
		return priceFloat8.Float64
	}

	return 0.0
}

func ToNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{Valid: true}
	if err := n.Scan(f); err == nil {
		n.Valid = true
	}

	return n
}
