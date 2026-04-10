package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/billing/internal/domain"
	"github.com/voxire/onion/internal/modules/billing/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetSummary(ctx context.Context, ownerID int64) (domain.BillingSummary, error) {
	return s.repo.GetSummary(ctx, ownerID)
}

func (s *Service) ListInvoices(ctx context.Context, ownerID int64) ([]domain.Invoice, error) {
	return s.repo.ListInvoices(ctx, ownerID)
}
