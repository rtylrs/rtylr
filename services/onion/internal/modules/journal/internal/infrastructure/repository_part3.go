package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// CountByStatus returns how many journal entries the owner has for each
// status (e.g. "draft", "posted"). An empty map is returned when the owner
// has no entries.
func (r *Repository) CountByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountJournalEntriesByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count journal entries by status", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.Cnt
	}

	return counts, nil
}

// SumLinesByAccountID returns the total debit and credit amounts posted to a
// given chart-of-accounts ID across all posted journal entries for an owner.
func (r *Repository) SumLinesByAccountID(ctx context.Context, businessID, accountID int64) (debit, credit float64, err error) {
	row, err := r.queries.SumJournalLinesByAccount(ctx, sqlc.SumJournalLinesByAccountParams{
		BusinessID: businessID,
		AccountID:  accountID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum journal lines by account", logger.Error(err))
		return 0, 0, err
	}

	return parseFloatAny(row.TotalDebit), parseFloatAny(row.TotalCredit), nil
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

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse nullable float from database",
			logger.String("value", s.String),
			logger.Error(err),
		)

		return 0
	}

	return v
}

// parseFloatAny converts an aggregate value returned by sqlc as interface{}
// (e.g. COALESCE(SUM(...), 0)) into a float64. Unparseable values are logged
// at Warn level and coerced to 0 rather than discarded silently.
func parseFloatAny(v any) float64 {
	switch val := v.(type) {
	case nil:
		return 0
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return parseFloatString(string(val))
	case string:
		return parseFloatString(val)
	default:
		return parseFloatString(fmt.Sprintf("%v", val))
	}
}

func parseFloatString(s string) float64 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse aggregate float from database",
			logger.String("value", s),
			logger.Error(err),
		)

		return 0
	}

	return v
}
