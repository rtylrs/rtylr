package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_users/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListByBusiness(ctx context.Context, businessID int64) ([]domain.POSUser, error) {
	return s.repo.ListByBusiness(ctx, businessID)
}

func (s *Service) ListByLocation(ctx context.Context, businessID, locationID int64) ([]domain.POSUser, error) {
	return s.repo.ListByLocation(ctx, businessID, locationID)
}

func (s *Service) Create(ctx context.Context, businessID, createdBy int64, input domain.CreatePOSUserInput) (int64, error) {
	return s.repo.Create(ctx, businessID, createdBy, input)
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.UpdatePOSUserInput) error {
	return s.repo.Update(ctx, businessID, id, input)
}

func (s *Service) Deactivate(ctx context.Context, businessID, id int64) error {
	return s.repo.Deactivate(ctx, businessID, id)
}

func (s *Service) VerifyPIN(ctx context.Context, businessID int64, pin string) (*domain.POSUser, error) {
	return s.repo.VerifyPIN(ctx, businessID, pin)
}
