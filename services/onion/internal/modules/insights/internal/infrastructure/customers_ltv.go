package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

type LTVBucketRow struct {
	Label           string
	LowerMinorUnits int64
	UpperMinorUnits int64
	CustomerCount   int64
}

var ltvBuckets = []struct {
	Label string
	Lower int64
	Upper int64
}{
	{"0-100", 0, 10_000},
	{"100-500", 10_000, 50_000},
	{"500-1k", 50_000, 100_000},
	{"1k-5k", 100_000, 500_000},
	{"5k-10k", 500_000, 1_000_000},
	{"10k+", 1_000_000, 0},
}

func (r *Repository) GetLTVDistribution(ctx context.Context, businessID int64) ([]LTVBucketRow, error) {
	// cross-schema read from crm_customers + pos_orders — lifetime spend histogram for LTV widget
	rows, err := r.queries.GetLTVSpendByCustomer(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ltv query failed", logger.Error(err))
		return nil, err
	}
	counts := make([]int64, len(ltvBuckets))
	for _, row := range rows {
		amount := decimalToMinorUnits(row.Spend)
		for i, b := range ltvBuckets {
			if amount >= b.Lower && (b.Upper == 0 || amount < b.Upper) {
				counts[i]++
				break
			}
		}
	}
	out := make([]LTVBucketRow, 0, len(ltvBuckets))
	for i, b := range ltvBuckets {
		out = append(out, LTVBucketRow{
			Label: b.Label, LowerMinorUnits: b.Lower, UpperMinorUnits: b.Upper, CustomerCount: counts[i],
		})
	}
	return out, nil
}
