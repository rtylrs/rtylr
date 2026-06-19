package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// MarginBucketRow is a single bucket in the margin-% distribution.
type MarginBucketRow struct {
	Label        string
	LowerPct     float64
	UpperPct     float64
	ProductCount int64
}

// margin buckets as percentages — product-level display choice, not a
// tenant-configurable business rule, so hardcoding is fine.
var marginBuckets = []struct {
	Label string
	Lower float64
	Upper float64
}{
	{"<0%", -1000, 0},
	{"0-10%", 0, 10},
	{"10-25%", 10, 25},
	{"25-50%", 25, 50},
	{"50-75%", 50, 75},
	{"75%+", 75, 1000},
}

// GetMarginDistribution returns the per-bucket SKU count from erp_products.
// Products without a sale_price (0) are excluded: a non-priced product
// doesn't have a meaningful margin.
func (r *Repository) GetMarginDistribution(ctx context.Context, businessID int64) ([]MarginBucketRow, error) {
	// cross-schema read from erp_products — margin distribution histogram
	const q = `
		SELECT (sale_price - cost_price) / sale_price * 100 AS margin_pct
		FROM erp_products
		WHERE business_id = ?
		  AND is_active = TRUE
		  AND sale_price > 0
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: margin distribution query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	counts := make([]int64, len(marginBuckets))
	for rows.Next() {
		var v interface{}
		if scanErr := rows.Scan(&v); scanErr != nil {
			logger.Ctx(ctx).Error("insights: margin scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		pct := toFloat(v)
		for i, b := range marginBuckets {
			if pct >= b.Lower && pct < b.Upper {
				counts[i]++
				break
			}
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	out := make([]MarginBucketRow, 0, len(marginBuckets))
	for i, b := range marginBuckets {
		out = append(out, MarginBucketRow{
			Label:        b.Label,
			LowerPct:     b.Lower,
			UpperPct:     b.Upper,
			ProductCount: counts[i],
		})
	}
	return out, nil
}
