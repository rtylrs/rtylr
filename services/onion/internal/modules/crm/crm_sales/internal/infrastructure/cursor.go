package infrastructure

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

// pageCursor is the keyset position after a page's last row: the sort it was
// issued for, that row's sort value and its id.
type pageCursor struct {
	Sort  string `json:"s"`
	Value string `json:"v"`
	ID    int64  `json:"id"`
}

func encodeCursor(sort, value string, id int64) string {
	payload, err := json.Marshal(pageCursor{Sort: sort, Value: value, ID: id})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}

// decodeCursor reads a cursor issued for sort. "" is the first page.
func decodeCursor(token, sort string) (*pageCursor, error) {
	if token == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, domain.ErrInvalidCursor
	}
	var cursor pageCursor
	if err := json.Unmarshal(raw, &cursor); err != nil || cursor.ID <= 0 || cursor.Sort != sort {
		return nil, domain.ErrInvalidCursor
	}
	return &cursor, nil
}

func (c *pageCursor) id() sql.NullInt64 {
	if c == nil {
		return sql.NullInt64{}
	}
	return someInt64(c.ID)
}

func (c *pageCursor) time() (sql.NullTime, error) {
	if c == nil {
		return sql.NullTime{}, nil
	}
	at, err := time.Parse(time.RFC3339Nano, c.Value)
	if err != nil {
		return sql.NullTime{}, domain.ErrInvalidCursor
	}
	return someTime(at.UTC()), nil
}

func (c *pageCursor) date() (sql.NullTime, error) {
	if c == nil {
		return sql.NullTime{}, nil
	}
	day, err := time.Parse(time.DateOnly, c.Value)
	if err != nil {
		return sql.NullTime{}, domain.ErrInvalidCursor
	}
	return someTime(day), nil
}

func (c *pageCursor) text() sql.NullString {
	if c == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: c.Value, Valid: true}
}

func timeValue(at time.Time) string {
	return at.UTC().Format(time.RFC3339Nano)
}

// Sort values for rows without one; they match the COALESCE defaults in the
// list queries so those rows sort last and page correctly.
const (
	noFollowUpDate = "9999-12-31"
	noCloseDate    = "9999-12-31"
)

var neverWorked = time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)
