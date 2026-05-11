package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/shifts/internal/infrastructure"
)

type repository interface {
	ListShifts(ctx context.Context, ownerID, staffID int64, date string) ([]domain.Shift, error)
	CreateShift(ctx context.Context, input domain.ShiftInput) (int64, error)
	UpdateShift(ctx context.Context, ownerID int64, input domain.ShiftUpdateInput) error
	DeleteShift(ctx context.Context, ownerID, shiftID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListShifts(ctx context.Context, ownerID, staffID int64, date string) ([]domain.Shift, error) {
	return s.repo.ListShifts(ctx, ownerID, staffID, date)
}

func (s *Service) CreateShift(ctx context.Context, input domain.ShiftInput) (int64, error) {
	return s.repo.CreateShift(ctx, input)
}

func (s *Service) UpdateShift(ctx context.Context, ownerID int64, input domain.ShiftUpdateInput) error {
	return s.repo.UpdateShift(ctx, ownerID, input)
}

func (s *Service) DeleteShift(ctx context.Context, ownerID, shiftID int64) error {
	return s.repo.DeleteShift(ctx, ownerID, shiftID)
}
