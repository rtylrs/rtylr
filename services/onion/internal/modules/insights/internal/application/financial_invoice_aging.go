package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// InvoiceAgingBucketPayload is one bucket of the A/R aging chart.
type InvoiceAgingBucketPayload struct {
	Bucket           string `json:"bucket"`
	InvoiceCount     int64  `json:"invoice_count"`
	AmountMinorUnits int64  `json:"amount_minor_units"`
}

// GetInvoiceAging returns the A/R aging buckets declared by the contract:
// current (not yet past due), 1-30, 31-60, 61-90 and 90+ days past due.
func (s *Service) GetInvoiceAging(ctx context.Context, businessID int64) ([]InvoiceAgingBucketPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "financial:invoice-aging")
	if err != nil {
		return nil, err
	}
	type result struct {
		Rows []InvoiceAgingBucketPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetInvoiceAging(ctx, businessID)
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]InvoiceAgingBucketPayload, 0, len(rows))
		for _, row := range rows {
			payloads = append(payloads, InvoiceAgingBucketPayload{
				Bucket:           row.Bucket,
				InvoiceCount:     row.InvoiceCount,
				AmountMinorUnits: row.AmountMinorUnits,
			})
		}
		return result{Rows: payloads}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Rows, nil
}
