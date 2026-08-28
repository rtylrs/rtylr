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
	List(ctx context.Context, businessID int64) ([]domain.RecurringInvoice, error)
	Create(ctx context.Context, businessID int64, input domain.RecurringInvoiceInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.RecurringInvoiceInput) error
	SetActive(ctx context.Context, businessID, id int64, active bool) error
	ProcessDue(ctx context.Context, now time.Time) ([]int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateRecurringInvoicesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(recurringInvoicesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.RecurringInvoice, error) {
	return cacher.Of[[]domain.RecurringInvoice]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(recurringInvoicesScope, businessID, "list"),
		func(ctx context.Context) ([]domain.RecurringInvoice, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, recurringInvoicesTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.RecurringInvoiceInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateRecurringInvoicesScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.RecurringInvoiceInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateRecurringInvoicesScope(ctx, businessID)
	return nil
}

func (s *Service) Pause(ctx context.Context, businessID, id int64) error {
	if err := s.repo.SetActive(ctx, businessID, id, false); err != nil {
		return err
	}
	invalidateRecurringInvoicesScope(ctx, businessID)
	return nil
}

func (s *Service) Resume(ctx context.Context, businessID, id int64) error {
	if err := s.repo.SetActive(ctx, businessID, id, true); err != nil {
		return err
	}
	invalidateRecurringInvoicesScope(ctx, businessID)
	return nil
}

// ProcessDue generates one invoice for every recurring schedule that is due.
// Each repository operation atomically creates the invoice and advances the
// schedule, so retries cannot leave a generated invoice paired with an old
// next_invoice_date.
func (s *Service) ProcessDue(ctx context.Context) error {
	businessIDs, err := s.repo.ProcessDue(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, businessID := range businessIDs {
		invalidateRecurringInvoicesScope(ctx, businessID)
	}
	return nil
}
