package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
	"github.com/voxire/onion/internal/modules/time_off/internal/infrastructure"
)

type repository interface {
	ListTimeOffRequests(ctx context.Context, ownerID int64) ([]domain.TimeOffRequest, error)
	CreateTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error)
	ApproveTimeOffRequest(ctx context.Context, ownerID int64, input domain.TimeOffApprovalInput) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListTimeOffRequests(ctx context.Context, ownerID int64) ([]domain.TimeOffRequest, error) {
	return s.repo.ListTimeOffRequests(ctx, ownerID)
}

func (s *Service) SubmitTimeOffRequest(ctx context.Context, input domain.TimeOffInput) (int64, error) {
	return s.repo.CreateTimeOffRequest(ctx, input)
}

func (s *Service) ApproveTimeOffRequest(ctx context.Context, ownerID int64, input domain.TimeOffApprovalInput) error {
	return s.repo.ApproveTimeOffRequest(ctx, ownerID, input)
}
