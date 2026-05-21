package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/billing/internal/domain"
	"github.com/voxire/onion/internal/modules/billing/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	billingScope = "billing"
	billingTTL   = 5 * time.Minute
)

type repository interface {
	GetSummary(ctx context.Context, ownerID int64) (domain.BillingSummary, error)
	ListInvoices(ctx context.Context, ownerID int64) ([]domain.Invoice, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) GetSummary(ctx context.Context, ownerID int64) (domain.BillingSummary, error) {
	return cacher.Of[domain.BillingSummary]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(billingScope, ownerID, "summary"),
		func(ctx context.Context) (domain.BillingSummary, error) {
			return s.repo.GetSummary(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, billingTTL),
	)
}

func (s *Service) ListInvoices(ctx context.Context, ownerID int64) ([]domain.Invoice, error) {
	return cacher.Of[[]domain.Invoice]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(billingScope, ownerID, "invoices"),
		func(ctx context.Context) ([]domain.Invoice, error) {
			return s.repo.ListInvoices(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, billingTTL),
	)
}
