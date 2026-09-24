package infrastructure

import (
	"database/sql"
	"time"
)

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func dateParam(value time.Time) sql.NullTime {
	return sql.NullTime{Time: dateOnly(value), Valid: true}
}
