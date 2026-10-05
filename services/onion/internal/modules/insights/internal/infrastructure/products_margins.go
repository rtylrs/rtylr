package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"

	"github.com/voxire/onion/internal/platform/logger"
)

type MarginBucketRow = domain.MarginBucketRow

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
	// cross-schema read from erp_products — aggregate the fixed histogram in SQL.
	counts, err := r.queries.GetProductMarginDistribution(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: margin distribution query failed", logger.Error(err))
		return nil, err
	}
	bucketCounts := []int64{
		counts.BelowZeroCount,
		counts.ZeroToTenCount,
		counts.TenToTwentyFiveCount,
		counts.TwentyFiveToFiftyCount,
		counts.FiftyToSeventyFiveCount,
		counts.SeventyFivePlusCount,
	}
	out := make([]MarginBucketRow, 0, len(marginBuckets))
	for i, b := range marginBuckets {
		out = append(out, MarginBucketRow{Label: b.Label, LowerPct: b.Lower, UpperPct: b.Upper, ProductCount: bucketCounts[i]})
	}
	return out, nil
}
