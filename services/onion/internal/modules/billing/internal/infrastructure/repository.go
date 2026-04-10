package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/billing/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// GetSummary returns a stub billing summary. Integrate with actual billing provider later.
func (r *Repository) GetSummary(_ context.Context, _ int64) (domain.BillingSummary, error) {
	return domain.BillingSummary{
		Plan:            "free",
		Status:          "active",
		CurrentUsage:    0,
		UsageLimit:      1000,
		NextBillingDate: nil,
	}, nil
}

// ListInvoices returns a stub invoices list. Integrate with actual billing provider later.
func (r *Repository) ListInvoices(_ context.Context, _ int64) ([]domain.Invoice, error) {
	return []domain.Invoice{}, nil
}
