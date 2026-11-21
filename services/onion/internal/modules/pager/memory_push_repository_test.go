package pager

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

// pushRow mirrors a pager_push_subscriptions row, keyed like the table on
// ticket and endpoint.
type pushRow struct {
	sub       domain.PushSubscription
	createdAt time.Time
}

func (r *memoryRepository) SavePushSubscription(_ context.Context, sub domain.PushSubscription, at time.Time, limit int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, row := range r.pushes {
		if row.sub.BusinessID == sub.BusinessID && row.sub.TicketID == sub.TicketID {
			count++
		}
	}
	if count >= limit {
		return domain.ErrTooManyPushSubscriptions
	}
	for i, row := range r.pushes {
		if row.sub.TicketID == sub.TicketID && row.sub.Endpoint == sub.Endpoint {
			r.pushes[i].sub.P256dh, r.pushes[i].sub.Auth = sub.P256dh, sub.Auth
			return nil
		}
	}
	r.pushSeq++
	sub.ID = r.pushSeq
	r.pushes = append(r.pushes, pushRow{sub: sub, createdAt: at})
	return nil
}

func (r *memoryRepository) PushSubscriptions(_ context.Context, businessID, ticketID int64) ([]domain.PushSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.PushSubscription
	for _, row := range r.pushes {
		if row.sub.BusinessID == businessID && row.sub.TicketID == ticketID {
			out = append(out, row.sub)
		}
	}
	return out, nil
}

func (r *memoryRepository) DeletePushSubscription(_ context.Context, businessID, id int64) error {
	r.deletePushes(func(row pushRow) bool { return row.sub.BusinessID == businessID && row.sub.ID == id })
	return nil
}

func (r *memoryRepository) DeleteTicketPushSubscriptions(_ context.Context, businessID, ticketID int64) error {
	r.deletePushes(func(row pushRow) bool { return row.sub.BusinessID == businessID && row.sub.TicketID == ticketID })
	return nil
}

func (r *memoryRepository) DeleteStalePushSubscriptions(_ context.Context, openSince time.Time) error {
	r.deletePushes(func(row pushRow) bool { return row.createdAt.Before(openSince) })
	return nil
}

func (r *memoryRepository) deletePushes(match func(pushRow) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	kept := r.pushes[:0]
	for _, row := range r.pushes {
		if !match(row) {
			kept = append(kept, row)
		}
	}
	r.pushes = kept
}
