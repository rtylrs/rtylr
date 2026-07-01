package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

type InvoiceAgingBucketRow struct {
	Bucket           string
	InvoiceCount     int64
	AmountMinorUnits int64
}

func (r *Repository) GetInvoiceAging(ctx context.Context, businessID int64) ([]InvoiceAgingBucketRow, error) {
	// cross-schema read from finance_invoices — A/R aging buckets for receivables panel
	rows, err := r.queries.GetInvoiceAging(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: invoice aging query failed", logger.Error(err))
		return nil, err
	}
	found := make(map[string]InvoiceAgingBucketRow)
	for _, r2 := range rows {
		if r2.Bucket == "not_due" {
			continue
		}
		found[r2.Bucket] = InvoiceAgingBucketRow{
			Bucket: r2.Bucket, InvoiceCount: r2.InvoiceCount,
			AmountMinorUnits: decimalToMinorUnits(r2.AmountDue),
		}
	}
	order := []string{"0-30", "31-60", "61-90", "90+"}
	out := make([]InvoiceAgingBucketRow, 0, len(order))
	for _, label := range order {
		if row, ok := found[label]; ok {
			out = append(out, row)
		} else {
			out = append(out, InvoiceAgingBucketRow{Bucket: label})
		}
	}
	return out, nil
}
