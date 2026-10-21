package infrastructure

import (
	"context"

	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type LTVBucketRow = domain.LTVBucketRow

// GetLTVDistribution buckets customers by lifetime paid spend. Customers
// without a paid order count in the lowest bucket.
func (r *Repository) GetLTVDistribution(ctx context.Context, businessID int64) ([]LTVBucketRow, error) {
	buckets, err := r.sources.POS.InsightsBuyerSpendBuckets(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ltv buckets failed", logger.Error(err))
		return nil, err
	}
	customers, err := r.sources.CRM.InsightsCustomerCount(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ltv customer count failed", logger.Error(err))
		return nil, err
	}
	return ltvBuckets(buckets, customers), nil
}

// ltvBuckets adds the customers POS has no paid order for to the lowest
// bucket.
func ltvBuckets(buckets []poscontract.SpendBucket, customers int64) []LTVBucketRow {
	var buyers int64
	for _, bucket := range buckets {
		buyers += bucket.CustomerCount
	}
	out := make([]LTVBucketRow, 0, len(buckets))
	for i, bucket := range buckets {
		count := bucket.CustomerCount
		if i == 0 && customers > buyers {
			count += customers - buyers
		}
		out = append(out, LTVBucketRow{
			Label: bucket.Label, LowerMinorUnits: bucket.LowerMinorUnits, UpperMinorUnits: bucket.UpperMinorUnits, CustomerCount: count,
		})
	}
	return out
}
