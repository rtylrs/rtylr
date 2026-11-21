package infrastructure

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	"github.com/voxire/onion/internal/modules/pager/internal/infrastructure/persistence/sqlc"
)

// SavePushSubscription stores a subscription, or refreshes the keys of the
// one the same browser already holds for the ticket. Once the ticket has
// limit subscriptions nothing more is stored. The count and the insert are
// not one transaction, so browsers subscribing at the same instant can pass
// the limit by a few, which only bounds abuse a little less tightly.
func (r *Repository) SavePushSubscription(ctx context.Context, sub domain.PushSubscription, at time.Time, limit int) error {
	count, err := r.queries.CountPushSubscriptions(ctx, sqlc.CountPushSubscriptionsParams{
		BusinessID: sub.BusinessID,
		TicketID:   sub.TicketID,
	})
	if err != nil {
		return err
	}
	if count >= int64(limit) {
		return domain.ErrTooManyPushSubscriptions
	}
	hash := sha256.Sum256([]byte(sub.Endpoint))
	return r.queries.SavePushSubscription(ctx, sqlc.SavePushSubscriptionParams{
		BusinessID:   sub.BusinessID,
		TicketID:     sub.TicketID,
		Endpoint:     sub.Endpoint,
		EndpointHash: hash[:],
		P256dh:       sub.P256dh,
		Auth:         sub.Auth,
		CreatedAt:    at,
	})
}

func (r *Repository) PushSubscriptions(ctx context.Context, businessID, ticketID int64) ([]domain.PushSubscription, error) {
	rows, err := r.queries.ListPushSubscriptions(ctx, sqlc.ListPushSubscriptionsParams{
		BusinessID: businessID,
		TicketID:   ticketID,
	})
	if err != nil {
		return nil, err
	}
	subs := make([]domain.PushSubscription, 0, len(rows))
	for _, row := range rows {
		subs = append(subs, domain.PushSubscription{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			TicketID:   row.TicketID,
			Endpoint:   row.Endpoint,
			P256dh:     row.P256dh,
			Auth:       row.Auth,
		})
	}
	return subs, nil
}

func (r *Repository) DeletePushSubscription(ctx context.Context, businessID, id int64) error {
	return r.queries.DeletePushSubscription(ctx, sqlc.DeletePushSubscriptionParams{ID: id, BusinessID: businessID})
}

func (r *Repository) DeleteTicketPushSubscriptions(ctx context.Context, businessID, ticketID int64) error {
	return r.queries.DeleteTicketPushSubscriptions(ctx, sqlc.DeleteTicketPushSubscriptionsParams{
		BusinessID: businessID,
		TicketID:   ticketID,
	})
}

// DeleteStalePushSubscriptions removes a batch of subscriptions, across all
// businesses, created before openSince.
func (r *Repository) DeleteStalePushSubscriptions(ctx context.Context, openSince time.Time) error {
	return r.queries.DeleteStalePushSubscriptions(ctx, openSince)
}
