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

func (r *Repository) GetLTVDistribution(ctx context.Context, businessID int64) ([]LTVBucketRow, error) {
	// SQL aggregates paid lifetime spend and returns the fixed six-bucket
	// histogram directly, including zero-spend customers and empty buckets.
	rows, err := r.queries.GetLTVDistribution(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ltv query failed", logger.Error(err))
		return nil, err
	}

	out := make([]LTVBucketRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, LTVBucketRow{
			Label: row.Label, LowerMinorUnits: int64(row.LowerMinorUnits), UpperMinorUnits: int64(row.UpperMinorUnits), CustomerCount: row.CustomerCount,
		})
	}
	return out, nil
}
