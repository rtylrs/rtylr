package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_terminals/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Terminal, error) {
	return s.repo.List(ctx, businessID)
}

func (s *Service) Create(ctx context.Context, input domain.TerminalInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.TerminalUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Heartbeat(ctx context.Context, terminalID, businessID int64) error {
	return s.repo.Heartbeat(ctx, terminalID, businessID)
}
