package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// CountByStatus returns the number of quotes in each status for the business.
func (r *Repository) CountByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountQuotesByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count quotes by status", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		counts[string(row.Status.FinanceQuotesStatus)] = row.Cnt
	}

	return counts, nil
}

func nullableStringInput(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStrValue(s sql.NullString) sql.NullString { return s }

func stringFromAny(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func floatFromAny(v any) float64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0
		}
		return f
	}
	return 0
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func quoteFromSqlc(row sqlc.FinanceQuote) domain.Quote {
	q := domain.Quote{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		CustomerName:  row.CustomerName,
		CustomerEmail: nullStr(row.CustomerEmail),
		Status:        string(row.Status.FinanceQuotesStatus),
		Items:         string(row.Items),
		Subtotal:      parseFloatNull(row.Subtotal),
		Tax:           parseFloatNull(row.Tax),
		Discount:      parseFloatNull(row.Discount),
		Total:         parseFloatNull(row.Total),
		Notes:         nullStr(row.Notes),
		CreatedAt:     nullTime(row.CreatedAt),
		UpdatedAt:     nullTime(row.UpdatedAt),
	}
	if row.ValidUntil.Valid {
		q.ValidUntil = row.ValidUntil.Time
	}
	return q
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func nullTimeStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal from database", logger.String("value", s.String), logger.Error(err))
		return 0
	}
	return v
}

func rawJSON(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("[]")
	}
	return json.RawMessage(s)
}
