package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/tasks/internal/domain"
	"github.com/voxire/onion/internal/modules/tasks/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Task, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, input domain.TaskInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.TaskUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}

func (s *Service) ListLabels(ctx context.Context, ownerID int64) ([]domain.Label, error) {
	return s.repo.ListLabels(ctx, ownerID)
}

func (s *Service) CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error) {
	return s.repo.CreateLabel(ctx, input)
}
