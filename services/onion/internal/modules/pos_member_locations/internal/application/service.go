package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/infrastructure"
)

type Service struct{ repo *infrastructure.Repository }

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListForUser(ctx context.Context, businessID, userID int64) ([]domain.MemberLocation, error) {
	return s.repo.ListForUser(ctx, businessID, userID)
}

func (s *Service) GetMyLocations(ctx context.Context, businessID, userID int64) ([]infrastructure.MyLocation, error) {
	return s.repo.GetMyLocations(ctx, businessID, userID)
}

func (s *Service) Assign(ctx context.Context, businessID, userID, locationID int64, canManagePOSUsers, canManageTerminals, canViewReports bool) (int64, error) {
	return s.repo.Assign(ctx, businessID, userID, locationID, canManagePOSUsers, canManageTerminals, canViewReports)
}

func (s *Service) Remove(ctx context.Context, businessID, id int64) error {
	return s.repo.Remove(ctx, businessID, id)
}
