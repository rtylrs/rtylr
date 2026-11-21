package pager

import (
	"context"
	"sync"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

// memoryRepository mirrors sql/queries/pager.sql predicate for predicate, so
// the tests that run without a database exercise the same rules. The mysql
// backend in pager_test.go runs the same tests against the real queries.
type memoryRepository struct {
	mu       sync.Mutex
	tickets  []*domain.Ticket
	settings map[int64]domain.Settings
	pushes   []pushRow
	pushSeq  int64
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{settings: map[int64]domain.Settings{}}
}

func isOpen(t *domain.Ticket) bool {
	return t.Status == domain.StatusWaiting || t.Status == domain.StatusReady
}

func (r *memoryRepository) Settings(_ context.Context, businessID int64) (domain.Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settings[businessID], nil
}

func (r *memoryRepository) SaveSettings(_ context.Context, businessID int64, settings domain.Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings[businessID] = settings
	return nil
}

func (r *memoryRepository) CreateTicket(_ context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tickets {
		if t.BusinessID == ticket.BusinessID && t.Number == ticket.Number && isOpen(t) {
			return domain.Ticket{}, domain.ErrNumberTaken
		}
	}
	ticket.ID = int64(len(r.tickets) + 1)
	ticket.Status = domain.StatusWaiting
	r.tickets = append(r.tickets, &ticket)
	return ticket, nil
}

// where returns the tickets of a business matching keep, oldest first.
func (r *memoryRepository) where(businessID int64, keep func(*domain.Ticket) bool) []*domain.Ticket {
	var out []*domain.Ticket
	for _, t := range r.tickets {
		if t.BusinessID == businessID && keep(t) {
			out = append(out, t)
		}
	}
	return out
}

func (r *memoryRepository) OpenTicketByNumber(_ context.Context, businessID int64, number string, openSince time.Time) (domain.Ticket, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.where(businessID, isOpen) {
		if t.Number == number && !t.CreatedAt.Before(openSince) {
			return *t, true, nil
		}
	}
	return domain.Ticket{}, false, nil
}

func (r *memoryRepository) ExpireStaleTicket(_ context.Context, businessID int64, number string, openSince time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.where(businessID, isOpen) {
		if t.Number == number && t.CreatedAt.Before(openSince) {
			t.Status = domain.StatusExpired
		}
	}
	return nil
}

func (r *memoryRepository) CountOpenCustomerTickets(_ context.Context, businessID int64, openSince time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.where(businessID, func(t *domain.Ticket) bool {
		return isOpen(t) && t.Source == domain.SourceCustomer && !t.CreatedAt.Before(openSince)
	}))), nil
}

func (r *memoryRepository) ListOpenTickets(_ context.Context, businessID int64, openSince time.Time) ([]domain.Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Ticket
	for _, t := range r.where(businessID, isOpen) {
		if !t.CreatedAt.Before(openSince) {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (r *memoryRepository) Ticket(_ context.Context, businessID, id int64) (domain.Ticket, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.where(businessID, func(t *domain.Ticket) bool { return t.ID == id }) {
		return *t, true, nil
	}
	return domain.Ticket{}, false, nil
}

func (r *memoryRepository) TicketByToken(_ context.Context, token string) (domain.Ticket, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tickets {
		if t.PublicToken == token {
			return *t, true, nil
		}
	}
	return domain.Ticket{}, false, nil
}

func (r *memoryRepository) CountWaitingAhead(_ context.Context, businessID, id int64, openSince time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.where(businessID, func(t *domain.Ticket) bool {
		return t.Status == domain.StatusWaiting && !t.CreatedAt.Before(openSince) && t.ID < id
	}))), nil
}

func (r *memoryRepository) LinkPhone(_ context.Context, businessID, id int64) error {
	return r.update(businessID, id, func(*domain.Ticket) bool { return true }, func(t *domain.Ticket) {
		t.PhoneLinked = true
	})
}

func (r *memoryRepository) MarkReady(_ context.Context, businessID, id int64, at, openSince time.Time) error {
	return r.update(businessID, id, openIn(openSince, domain.StatusWaiting), func(t *domain.Ticket) {
		t.Status, t.ReadyAt, t.RemindedAt, t.AcknowledgedAt = domain.StatusReady, at, time.Time{}, time.Time{}
	})
}

func (r *memoryRepository) Remind(_ context.Context, businessID, id int64, at, openSince time.Time) error {
	return r.update(businessID, id, openIn(openSince, domain.StatusReady), func(t *domain.Ticket) {
		t.RemindedAt, t.AcknowledgedAt = at, time.Time{}
	})
}

func (r *memoryRepository) UndoReady(_ context.Context, businessID, id int64, openSince time.Time) error {
	return r.update(businessID, id, openIn(openSince, domain.StatusReady), func(t *domain.Ticket) {
		t.Status, t.ReadyAt, t.RemindedAt, t.AcknowledgedAt = domain.StatusWaiting, time.Time{}, time.Time{}, time.Time{}
	})
}

func (r *memoryRepository) Collect(_ context.Context, businessID, id int64, openSince time.Time) error {
	return r.update(businessID, id, openIn(openSince, domain.StatusWaiting, domain.StatusReady), func(t *domain.Ticket) {
		t.Status = domain.StatusCollected
	})
}

func (r *memoryRepository) Acknowledge(_ context.Context, businessID, id int64, at, openSince time.Time) error {
	ringing := openIn(openSince, domain.StatusReady)
	return r.update(businessID, id, func(t *domain.Ticket) bool { return ringing(t) && t.AcknowledgedAt.IsZero() }, func(t *domain.Ticket) {
		t.AcknowledgedAt = at
	})
}

// openIn matches a ticket that is still open and in one of statuses.
func openIn(openSince time.Time, statuses ...string) func(*domain.Ticket) bool {
	return func(t *domain.Ticket) bool {
		if t.CreatedAt.Before(openSince) {
			return false
		}
		for _, status := range statuses {
			if t.Status == status {
				return true
			}
		}
		return false
	}
}

func (r *memoryRepository) update(businessID, id int64, match func(*domain.Ticket) bool, change func(*domain.Ticket)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.where(businessID, func(t *domain.Ticket) bool { return t.ID == id }) {
		if match(t) {
			change(t)
		}
	}
	return nil
}
