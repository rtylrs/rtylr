package application

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	recurringInvoicesScope = "recurring_invoices"
	recurringInvoicesTTL   = 10 * time.Minute
)

type repository interface {
	ListPage(ctx context.Context, businessID int64, limit, offset int32) ([]domain.RecurringInvoice, error)
	Get(ctx context.Context, businessID, id int64) (domain.RecurringInvoice, error)
	Create(ctx context.Context, businessID int64, input domain.RecurringInvoiceInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.RecurringInvoiceInput) error
	SetActive(ctx context.Context, businessID, id int64, active bool) error
	ProcessDue(ctx context.Context, now time.Time) ([]int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateRecurringInvoicesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(recurringInvoicesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID, page, pageSize int64) ([]domain.RecurringInvoice, bool, error) {
	limit, offset := recurringInvoicePagination(page, pageSize)
	rows, err := cacher.Of[[]domain.RecurringInvoice]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(recurringInvoicesScope, businessID, fmt.Sprintf("list|p=%d|ps=%d", page, pageSize))+"|domain=v2",
		func(ctx context.Context) ([]domain.RecurringInvoice, error) {
			return s.repo.ListPage(ctx, businessID, int32(limit+1), int32(offset)) // #nosec G115 -- limit is capped at 500 and offset is clamped to MaxInt32.
		},
		cacher.WithExpiry(ctx, recurringInvoicesTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func (s *Service) Get(ctx context.Context, businessID, id int64) (domain.RecurringInvoice, error) {
	item, err := s.repo.Get(ctx, businessID, id)
	if err != nil {
		return domain.RecurringInvoice{}, err
	}
	return item, nil
}

func recurringInvoicePagination(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	offset := int64(math.MaxInt32)
	if page-1 <= math.MaxInt32/pageSize {
		offset = (page - 1) * pageSize
	}
	return int(pageSize), int(offset)
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
