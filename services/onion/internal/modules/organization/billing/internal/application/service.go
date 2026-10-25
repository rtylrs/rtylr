package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/organization/billing/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	billingScope = "billing"
	billingTTL   = 5 * time.Minute
)

type repository interface {
	GetSummary(ctx context.Context, businessID int64) (domain.BillingSummary, error)
	ListInvoices(ctx context.Context, businessID int64) ([]domain.Invoice, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetSummary(ctx context.Context, businessID int64) (domain.BillingSummary, error) {
	return cacher.Of[domain.BillingSummary]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(billingScope, businessID, "business_summary"),
		func(ctx context.Context) (domain.BillingSummary, error) {
			return s.repo.GetSummary(ctx, businessID)
		},
		cacher.WithExpiry(ctx, billingTTL),
	)
}

func (s *Service) ListInvoices(ctx context.Context, businessID int64) ([]domain.Invoice, error) {
	return cacher.Of[[]domain.Invoice]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(billingScope, businessID, "business_invoices"),
		func(ctx context.Context) ([]domain.Invoice, error) {
			return s.repo.ListInvoices(ctx, businessID)
		},
		cacher.WithExpiry(ctx, billingTTL),
	)
}
