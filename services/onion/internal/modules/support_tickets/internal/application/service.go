package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/domain"
	"github.com/voxire/onion/internal/modules/support_tickets/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, businessID int64, status string, page, pageSize int) ([]domain.Ticket, int, error) {
	return s.repo.List(ctx, businessID, status, page, pageSize)
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (*domain.Ticket, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) Create(ctx context.Context, input domain.TicketInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.TicketUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) ListMessages(ctx context.Context, ticketID int64) ([]domain.Message, error) {
	return s.repo.ListMessages(ctx, ticketID)
}

func (s *Service) CreateMessage(ctx context.Context, input domain.MessageInput) (int64, error) {
	return s.repo.CreateMessage(ctx, input)
}
