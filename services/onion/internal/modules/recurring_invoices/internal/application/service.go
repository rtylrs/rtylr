package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	recurringInvoicesScope = "recurring_invoices"
	recurringInvoicesTTL   = 10 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.RecurringInvoice, error)
	Create(ctx context.Context, ownerID int64, input domain.RecurringInvoiceInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.RecurringInvoiceInput) error
	SetActive(ctx context.Context, ownerID, id int64, active bool) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateRecurringInvoicesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(recurringInvoicesScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.RecurringInvoice, error) {
	return cacher.Of[[]domain.RecurringInvoice]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(recurringInvoicesScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.RecurringInvoice, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, recurringInvoicesTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.RecurringInvoiceInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateRecurringInvoicesScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.RecurringInvoiceInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	invalidateRecurringInvoicesScope(ctx, ownerID)
	return nil
}

func (s *Service) Pause(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.SetActive(ctx, ownerID, id, false); err != nil {
		return err
	}
	invalidateRecurringInvoicesScope(ctx, ownerID)
	return nil
}

func (s *Service) Resume(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.SetActive(ctx, ownerID, id, true); err != nil {
		return err
	}
	invalidateRecurringInvoicesScope(ctx, ownerID)
	return nil
}
