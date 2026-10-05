package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type InvoiceAgingBucketRow = domain.InvoiceAgingBucketRow

// invoiceAgingBucketOrder is the bucket set declared by the
// insights.financial.v1 InvoiceAgingBucket contract.
var invoiceAgingBucketOrder = []string{"current", "1-30", "31-60", "61-90", "90+"}

// invoiceAgingBucketsFromRows maps raw aging rows to the contract's bucket
// set, always emitting every bucket in contract order and zero-filling the
// ones with no outstanding invoices. Not-yet-due invoices are part of
// "current", so total outstanding includes current receivables.
func invoiceAgingBucketsFromRows(rows []sqlc.GetInvoiceAgingRow) []InvoiceAgingBucketRow {
	found := make(map[string]InvoiceAgingBucketRow, len(rows))
	for _, row := range rows {
		found[row.Bucket] = InvoiceAgingBucketRow{
			Bucket:           row.Bucket,
			InvoiceCount:     row.InvoiceCount,
			AmountMinorUnits: decimalToMinorUnits(row.AmountDue),
		}
	}
	out := make([]InvoiceAgingBucketRow, 0, len(invoiceAgingBucketOrder))
	for _, label := range invoiceAgingBucketOrder {
		if row, ok := found[label]; ok {
			out = append(out, row)
		} else {
			out = append(out, InvoiceAgingBucketRow{Bucket: label})
		}
	}
	return out
}

func (r *Repository) GetInvoiceAging(ctx context.Context, businessID int64) ([]InvoiceAgingBucketRow, error) {
	// cross-schema read from finance_invoices — A/R aging buckets for receivables panel
	rows, err := r.queries.GetInvoiceAging(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: invoice aging query failed", logger.Error(err))
		return nil, err
	}
	return invoiceAgingBucketsFromRows(rows), nil
}
