package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// InvoiceAgingBucketRow is one A/R aging bucket. Amount is the sum of
// amount_due across all outstanding invoices that fall in the bucket.
type InvoiceAgingBucketRow struct {
	Bucket           string
	InvoiceCount     int64
	AmountMinorUnits int64
}

// GetInvoiceAging returns the four standard A/R aging buckets.
//
// We age each outstanding invoice against the current date relative to
// its due_date. "0-30" means 0-30 days past due (inclusive of 0 — due
// today). Invoices with status 'paid' or 'cancelled'/'void' are excluded.
func (r *Repository) GetInvoiceAging(ctx context.Context, businessID int64) ([]InvoiceAgingBucketRow, error) {
	// cross-schema read from finance_invoices — A/R aging buckets for receivables panel
	const q = `
		SELECT
		  CASE
		    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 0 AND 30 THEN '0-30'
		    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 31 AND 60 THEN '31-60'
		    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 61 AND 90 THEN '61-90'
		    WHEN DATEDIFF(CURDATE(), due_date) > 90 THEN '90+'
		    ELSE 'not_due'
		  END AS bucket,
		  COUNT(*) AS invoice_count,
		  COALESCE(SUM(amount_due), 0) AS amount_due
		FROM finance_invoices
		WHERE business_id = ?
		  AND status NOT IN ('paid','cancelled','void')
		  AND amount_due > 0
		GROUP BY bucket
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: invoice aging query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	found := make(map[string]InvoiceAgingBucketRow)
	for rows.Next() {
		var r2 InvoiceAgingBucketRow
		var amount interface{}
		if scanErr := rows.Scan(&r2.Bucket, &r2.InvoiceCount, &amount); scanErr != nil {
			logger.Ctx(ctx).Error("insights: invoice aging scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		r2.AmountMinorUnits = decimalToMinorUnits(amount)
		found[r2.Bucket] = r2
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	// Always emit all four bucket slots in order so the frontend can render
	// a stable chart even when a bucket is empty.
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
