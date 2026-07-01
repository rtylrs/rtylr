package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

type MarginBucketRow struct {
	Label        string
	LowerPct     float64
	UpperPct     float64
	ProductCount int64
}

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

func (r *Repository) GetMarginDistribution(ctx context.Context, businessID int64) ([]MarginBucketRow, error) {
	// cross-schema read from erp_products — margin distribution histogram
	rows, err := r.queries.GetProductMargins(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: margin distribution query failed", logger.Error(err))
		return nil, err
	}
	counts := make([]int64, len(marginBuckets))
	for _, margin := range rows {
		pct := toFloat(margin)
		for i, b := range marginBuckets {
			if pct >= b.Lower && pct < b.Upper {
				counts[i]++
				break
			}
		}
	}
	out := make([]MarginBucketRow, 0, len(marginBuckets))
	for i, b := range marginBuckets {
		out = append(out, MarginBucketRow{Label: b.Label, LowerPct: b.Lower, UpperPct: b.Upper, ProductCount: counts[i]})
	}
	return out, nil
}
