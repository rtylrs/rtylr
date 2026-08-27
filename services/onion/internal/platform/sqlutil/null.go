package sqlutil

import "database/sql"

// NullString converts an optional string value to the nullable form expected
// by database/sql. Empty strings represent absent values throughout the
// repository layer.
func NullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

// NullInt64 converts the repository's zero-value optional IDs to
// database/sql's nullable form.
func NullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}
