package infrastructure

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// Repository aggregates cross-schema insights reads. Static queries use sqlc;
// churn cohort SQL is tagged dynamic-query: squirrel in customers_churn.go.
type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

// decimalToMinorUnits converts a DECIMAL(10,2) string/bytes value read out
// of MySQL into integer minor units (cents). Empty or NULL values return 0.
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

func deltaPct(current, prior int64) float64 {
	if prior == 0 {
		return 0
	}
	return (float64(current-prior) / float64(prior)) * 100
}

func deltaPctFloat(current, prior float64) float64 {
	if prior == 0 {
		return 0
	}
	return ((current - prior) / prior) * 100
}

func mtdRanges(now time.Time) (thisFrom, thisTo, priorFrom, priorTo time.Time) {
	thisFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thisTo = now
	priorFrom = thisFrom.AddDate(0, -1, 0)
	dayOffset := int(now.Sub(thisFrom).Hours() / 24)
	priorTo = priorFrom.AddDate(0, 0, dayOffset)
	return thisFrom, thisTo, priorFrom, priorTo
}

func ifaceToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case float64:
		return int64(val + 0.5)
	case []byte:
		return decimalToMinorUnits(string(val)) / 100
	case string:
		return decimalToMinorUnits(val) / 100
	default:
		return 0
	}
}
