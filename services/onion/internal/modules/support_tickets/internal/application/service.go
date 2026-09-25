package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/domain"
	"github.com/voxire/onion/internal/modules/support_tickets/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	supportTicketsScope = "support_tickets"
	supportTicketsTTL   = 30 * time.Second
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateSupportTicketsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(supportTicketsScope, businessID))
}

type ticketsListResult struct {
	Items []domain.Ticket
	Total int
}

func (s *Service) List(ctx context.Context, businessID int64, status string, page, pageSize int) ([]domain.Ticket, int, error) {
	key := cacheKeys.Scoped(supportTicketsScope, businessID, fmt.Sprintf("list|s=%s|p=%d|ps=%d", status, page, pageSize))
	res, err := cacher.Of[ticketsListResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (ticketsListResult, error) {
			items, total, err := s.repo.List(ctx, businessID, status, page, pageSize)
			return ticketsListResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, supportTicketsTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (*domain.Ticket, error) {
	key := cacheKeys.Scoped(supportTicketsScope, businessID, fmt.Sprintf("id=%d", id))
	return cacher.Of[*domain.Ticket]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.Ticket, error) {
			return s.repo.GetByID(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, supportTicketsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.TicketInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateSupportTicketsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.TicketUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateSupportTicketsScope(ctx, input.BusinessID)
	return nil
}

// Messages are read through the parent ticket's business scope: the query joins
// support_tickets and filters by business_id, so a caller can never read another
// tenant's ticket messages. The cache key family is keyed by ticketID and wiped
// on every message create.
func (s *Service) ListMessages(ctx context.Context, businessID, ticketID int64) ([]domain.Message, error) {
	key := cacheKeys.Scoped("support_ticket_messages", ticketID, "list")
	return cacher.Of[[]domain.Message]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Message, error) {
			return s.repo.ListMessages(ctx, businessID, ticketID)
		},
		cacher.WithExpiry(ctx, supportTicketsTTL),
	)
}

func (s *Service) CreateMessage(ctx context.Context, input domain.MessageInput) (int64, error) {
	id, err := s.repo.CreateMessage(ctx, input)
	if err != nil {
		return 0, err
	}
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix("support_ticket_messages", input.TicketID))
	return id, nil
}
