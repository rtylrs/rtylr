package convert

import (
	"database/sql"
	"encoding/json"
	"strconv"
)

func ToString(t sql.NullString) string {
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

func ToBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}

	return false
}

func ToInt32(i sql.NullInt32) int32 {
	if i.Valid {
		return i.Int32
	}

	return 0
}

func ToInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}

	return 0
}

func ToStringSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}

	return values
}
func BoolToString(b bool) string {
	if b {
		return "1"
	}

	return "0"
}

func ToFloat64(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}

	return 0.0
}
