package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_reminders/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Reminder, error)
	Create(ctx context.Context, ownerID int64, input domain.ReminderInput) (int64, error)
}

type Service struct{ repo repository }

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Reminder, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.ReminderInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}
