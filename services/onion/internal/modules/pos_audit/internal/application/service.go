package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Create(ctx context.Context, input domain.AuditLogInput) (int64, error) {
	return s.repo.Create(ctx, input)
}
