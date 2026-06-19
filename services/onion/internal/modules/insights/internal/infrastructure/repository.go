package infrastructure

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// Repository is the raw-SQL aggregation repository. It does not use sqlc
// because every query in this module is a cross-schema aggregation that
// would need its own bespoke migration to be parseable by sqlc.
//
// All queries must filter by business_id and must annotate cross-schema reads
// with a `// cross-schema read from {table} — {reason}` comment directly
// above the query, per onion/CLAUDE.md §4.
type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

// decimalToMinorUnits converts a DECIMAL(10,2) string/bytes value read out
// of MySQL into integer minor units (cents). Empty or NULL values return 0.
//
// This helper exists because the insights module never trusts the client
// to do decimal math — the whole API speaks int64 minor units end-to-end.
func decimalToMinorUnits(v interface{}) int64 {
	if v == nil {
		return 0
	}
	var s string
	switch val := v.(type) {
	case string:
		s = val
	case []byte:
		s = string(val)
	case float64:
		// Rare path — driver gave us a float. Round to avoid 0.01 drift.
		return int64(val*100 + 0.5)
	case int64:
		return val * 100
	default:
		return 0
	}
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("insights: failed to parse decimal value",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return int64(f*100 + 0.5)
}

// deltaPct computes a percent delta of `current` vs `prior`, safe when
// the prior period is zero (returns 0 to avoid divide-by-zero and avoid
// surfacing a fake +∞% jump on a brand-new tenant).
func deltaPct(current, prior int64) float64 {
	if prior == 0 {
		return 0
	}
	return (float64(current-prior) / float64(prior)) * 100
}

// deltaPctFloat is deltaPct for float inputs (used for margin %).
func deltaPctFloat(current, prior float64) float64 {
	if prior == 0 {
		return 0
	}
	return ((current - prior) / prior) * 100
}

// mtdRanges returns the (MTD, prior-MTD) date ranges used by dashboard
// cards. MTD is [first-of-month, now]; prior-MTD is the equivalent number
// of days in the previous month.
func mtdRanges(now time.Time) (thisFrom, thisTo, priorFrom, priorTo time.Time) {
	thisFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thisTo = now
	// Prior month same day-of-month; if current month is shorter, clamp.
	priorFrom = thisFrom.AddDate(0, -1, 0)
	dayOffset := int(now.Sub(thisFrom).Hours() / 24)
	priorTo = priorFrom.AddDate(0, 0, dayOffset)
	return thisFrom, thisTo, priorFrom, priorTo
}
