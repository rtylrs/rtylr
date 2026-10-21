package infrastructure

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"time"

	erpcontract "github.com/voxire/onion/internal/contracts/erp"
	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// Repository serves Insights' own tables and builds cross-module reports from
// the owning modules' reads in sources.
type Repository struct {
	queries *sqlc.Queries
	sources Sources
}

// GetInsightsGeneration returns the latest change across every source, so a
// write in any owning module moves cached reports to a new key.
func (r *Repository) GetInsightsGeneration(ctx context.Context, businessID int64) (time.Time, error) {
	var latest time.Time
	for _, source := range []lastChange{r.sources.POS, r.sources.Finance, r.sources.CRM, r.sources.ERP, r.sources.HR} {
		changed, err := source.InsightsLastChange(ctx, businessID)
		if err != nil {
			return time.Time{}, err
		}
		if changed.After(latest) {
			latest = changed
		}
	}
	return latest, nil
}

func NewRepository(db *sql.DB, sources Sources) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		sources: sources,
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

// GetProductMarginRows returns per-product margin rows joining all-time sales
// volume with cost price (#853), highest revenue first.
func (r *Repository) GetProductMarginRows(ctx context.Context, businessID int64) ([]domain.ProductMarginRow, error) {
	sales, err := r.sources.POS.InsightsLifetimeProductSales(ctx, businessID)
	if err == nil {
		var facts map[int64]erpcontract.ProductFacts
		if facts, err = r.productFacts(ctx, businessID, sales); err == nil {
			return productMarginRows(sales, facts), nil
		}
	}
	logger.Ctx(ctx).Error("failed to get product margin rows", logger.Any("business_id", businessID), logger.Error(err))
	return nil, err
}

// productMarginRowLimit caps the per-product margin table.
const productMarginRowLimit = 200

func productMarginRows(sales []poscontract.ProductSales, facts map[int64]erpcontract.ProductFacts) []domain.ProductMarginRow {
	out := make([]domain.ProductMarginRow, 0, len(sales))
	for _, sale := range sales {
		product, ok := facts[sale.ProductID]
		if !ok {
			continue
		}
		var cost *int64
		if product.CostMinorUnits != 0 {
			v := product.CostMinorUnits
			cost = &v
		}
		out = append(out, domain.ProductMarginRow{
			ID:             sale.ProductID,
			Name:           product.Name,
			UnitsSold:      sale.Units,
			RevenueMinor:   sale.RevenueMinorUnits,
			CostMinorUnits: cost,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].RevenueMinor != out[j].RevenueMinor {
			return out[i].RevenueMinor > out[j].RevenueMinor
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > productMarginRowLimit {
		out = out[:productMarginRowLimit]
	}
	return out
}
