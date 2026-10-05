package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// Repository aggregates cross-schema insights reads. Static queries use sqlc;
// churn cohort SQL is tagged dynamic-query: squirrel in customers_churn.go.
type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func (r *Repository) GetInsightsGeneration(ctx context.Context, businessID int64) (time.Time, error) {
	return r.queries.GetInsightsGeneration(ctx, sqlc.GetInsightsGenerationParams{
		BusinessID:   businessID,
		BusinessID_2: businessID,
		BusinessID_3: businessID,
		BusinessID_4: businessID,
		BusinessID_5: businessID,
		BusinessID_6: businessID,
		BusinessID_7: businessID,
		BusinessID_8: businessID,
		BusinessID_9: businessID,
	})
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

func mtdRanges(now time.Time) (thisFrom, thisTo, priorFrom, priorTo time.Time) {
	thisFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thisTo = now
	priorFrom = thisFrom.AddDate(0, -1, 0)
	priorTo = priorFrom.Add(now.Sub(thisFrom))
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

// GetProductMarginRows returns per-product margin rows joining sales volume
// with cost price (#853).
func (r *Repository) GetProductMarginRows(ctx context.Context, businessID int64) ([]domain.ProductMarginRow, error) {
	rows, err := r.queries.GetProductMarginRows(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get product margin rows", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.ProductMarginRow, 0, len(rows))
	for _, row := range rows {
		units := decimalToFloat(row.UnitsSold)
		revenue := row.RevenueMinorUnitsX100
		var cost *int64
		if row.CostMinorUnitsX100 != 0 {
			v := row.CostMinorUnitsX100
			cost = &v
		}
		out = append(out, domain.ProductMarginRow{
			ID:             row.ID,
			Name:           row.Name,
			UnitsSold:      units,
			RevenueMinor:   revenue,
			CostMinorUnits: cost,
		})
	}
	return out, nil
}

// decimalToFloat converts a DECIMAL column value (which the MySQL driver
// returns as []byte) or a numeric value into float64. Unlike
// decimalToMinorUnits it keeps the source scale, so DECIMAL(8,3) quantities
// do not round to two places.
func decimalToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	var s string
	switch n := v.(type) {
	case int64:
		return float64(n)
	case float64:
		return n
	case []byte:
		s = string(n)
	case string:
		s = n
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
	return f
}
