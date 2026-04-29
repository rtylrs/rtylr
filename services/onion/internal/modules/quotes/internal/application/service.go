package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Quote, error)
	Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error
	Delete(ctx context.Context, ownerID, id int64) error
	ConvertToInvoice(ctx context.Context, ownerID, quoteID int64) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Quote, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}

func (s *Service) ConvertToInvoice(ctx context.Context, ownerID, quoteID int64) (int64, error) {
	return s.repo.ConvertToInvoice(ctx, ownerID, quoteID)
}
