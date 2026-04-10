package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/timesheets/internal/domain"
	"github.com/voxire/onion/internal/modules/timesheets/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetTimesheets(ctx context.Context, ownerID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	return s.repo.GetTimesheets(ctx, ownerID, weekStart)
}
