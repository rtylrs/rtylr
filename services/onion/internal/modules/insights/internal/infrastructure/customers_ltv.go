package infrastructure

import (
	"context"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/logger"
)

// LTVBucketRow is a distribution bucket for the LTV histogram. Upper
// bounds are exclusive except the final open-ended bucket (100k+).
type LTVBucketRow struct {
	Label           string
	LowerMinorUnits int64
	UpperMinorUnits int64 // 0 means "no upper bound"
	CustomerCount   int64
}

// lifetime spend thresholds in minor units (cents, USD for now). These
// are a product-level display bucketing, not a business rule — keeping
// them hardcoded is fine because they do not differ per tenant.
var ltvBuckets = []struct {
	Label string
	Lower int64
	Upper int64 // 0 = unbounded
}{
	{"0-100", 0, 10_000},
	{"100-500", 10_000, 50_000},
	{"500-1k", 50_000, 100_000},
	{"1k-5k", 100_000, 500_000},
	{"5k-10k", 500_000, 1_000_000},
	{"10k+", 1_000_000, 0},
}

// GetLTVDistribution returns per-bucket customer counts based on their
// lifetime spend from paid POS orders. Customers with no orders at all
// fall into the 0-100 bucket.
func (r *Repository) GetLTVDistribution(ctx context.Context, businessID int64) ([]LTVBucketRow, error) {
	// cross-schema read from crm_customers + pos_orders — lifetime spend histogram for LTV widget
	q, args, err := sq.
		Select("c.id", "COALESCE(SUM(o.total), 0) AS spend").
		From("crm_customers c").
		LeftJoin("pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id AND o.payment_status = 'paid'").
		Where(sq.Eq{"c.business_id": businessID}).
		GroupBy("c.id").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build ltv query failed", logger.Error(err))
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("insights: ltv query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	counts := make([]int64, len(ltvBuckets))
	for rows.Next() {
		var id int64
		var spend interface{}
		if scanErr := rows.Scan(&id, &spend); scanErr != nil {
			logger.Ctx(ctx).Error("insights: ltv scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		amount := decimalToMinorUnits(spend)
		for i, b := range ltvBuckets {
			if amount >= b.Lower && (b.Upper == 0 || amount < b.Upper) {
				counts[i]++
				break
			}
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	out := make([]LTVBucketRow, 0, len(ltvBuckets))
	for i, b := range ltvBuckets {
		out = append(out, LTVBucketRow{
			Label:           b.Label,
			LowerMinorUnits: b.Lower,
			UpperMinorUnits: b.Upper,
			CustomerCount:   counts[i],
		})
	}
	return out, nil
}
