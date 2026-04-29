package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Bill, error)
	Get(ctx context.Context, ownerID, id int64) (domain.Bill, error)
	Delete(ctx context.Context, ownerID, id int64) error
	Create(ctx context.Context, ownerID int64, input domain.BillInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.BillInput) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Bill, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) GetBill(ctx context.Context, ownerID, id int64) (domain.Bill, error) {
	return s.repo.Get(ctx, ownerID, id)
}

func (s *Service) DeleteBill(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.BillInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.BillInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}
