package application

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// PushSender delivers Web Push messages. PublicKey is the VAPID key browsers
// subscribe with; Send reports domain.ErrPushGone for a subscription the push
// service no longer knows.
type PushSender interface {
	PublicKey() string
	Send(ctx context.Context, sub domain.PushSubscription, payload []byte, topic string) error
}

// pushTimeout bounds one round of pushes for a ticket, which runs after the
// staff request has been answered.
const pushTimeout = 30 * time.Second

// EnablePush turns on lock-screen notifications. Without it the pager works
// as before and subscriptions are accepted but not stored. run starts the
// pushes for a staff action; nil runs them on their own goroutine.
func (s *Service) EnablePush(sender PushSender, run func(func())) {
	if run == nil {
		run = func(f func()) { go f() }
	}
	s.push, s.runPush = sender, run
}

// PushKey is the VAPID public key, empty while push is off.
func (s *Service) PushKey() string {
	if s.push == nil {
		return ""
	}
	return s.push.PublicKey()
}

// SubscribePush keeps a browser's push subscription for an open ticket at a
// venue whose pager is on.
func (s *Service) SubscribePush(ctx context.Context, token string, sub domain.PushSubscription) error {
	now := s.clock()
	ticket, err := s.ticketByToken(ctx, token)
	if err != nil {
		return err
	}
	status := ticket.StatusAt(now)
	if status != domain.StatusWaiting && status != domain.StatusReady {
		return domain.ErrTicketNotFound
	}
	if err := domain.ValidatePushSubscription(sub.Endpoint, sub.P256dh, sub.Auth); err != nil {
		return err
	}
	if s.push == nil {
		return nil
	}
	// Subscriptions are only taken for open tickets, so one older than a
	// ticket's lifetime belongs to a ticket that is over, collected or not.
	if err := s.repo.DeleteStalePushSubscriptions(ctx, domain.OpenSince(now)); err != nil {
		return err
	}
	sub.BusinessID, sub.TicketID = ticket.BusinessID, ticket.ID
	return s.repo.SavePushSubscription(ctx, sub, now, domain.MaxPushSubscriptionsPerTicket)
}

// pushPayload is what the service worker receives, encrypted for the browser.
type pushPayload struct {
	Kind   string `json:"kind"`
	Number string `json:"number"`
	Venue  string `json:"venue"`
	Token  string `json:"token"`
}

// notify starts the pushes for a ticket that just became ready or was called
// again. The staff action has already succeeded; a failed push is logged.
func (s *Service) notify(ctx context.Context, ticket domain.Ticket, kind string) {
	if s.push == nil {
		return
	}
	ctx = context.WithoutCancel(ctx)
	s.runPush(func() {
		ctx, cancel := context.WithTimeout(ctx, pushTimeout)
		defer cancel()
		if err := s.sendPushes(ctx, ticket, kind); err != nil {
			logger.Ctx(ctx).Warn("pager push failed", logger.Any("ticket_id", ticket.ID), logger.Error(err))
		}
	})
}

func (s *Service) sendPushes(ctx context.Context, ticket domain.Ticket, kind string) error {
	subs, err := s.repo.PushSubscriptions(ctx, ticket.BusinessID, ticket.ID)
	if err != nil || len(subs) == 0 {
		return err
	}
	venue, err := s.venues.VenueByID(ctx, ticket.BusinessID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(pushPayload{Kind: kind, Number: ticket.Number, Venue: venue.Name, Token: ticket.PublicToken})
	if err != nil {
		return err
	}
	// A call again replaces a ready message the phone has not received yet.
	topic := "pager-" + strconv.FormatInt(ticket.ID, 36)
	var failed error
	for _, sub := range subs {
		err := s.push.Send(ctx, sub, payload, topic)
		switch {
		case errors.Is(err, domain.ErrPushGone):
			if err := s.repo.DeletePushSubscription(ctx, ticket.BusinessID, sub.ID); err != nil {
				failed = errors.Join(failed, err)
			}
		case err != nil:
			failed = errors.Join(failed, err)
		}
	}
	return failed
}

// forgetPushes drops a closed ticket's subscriptions. The ticket has already
// moved, so a failure is logged and the sweep removes the rows later.
func (s *Service) forgetPushes(ctx context.Context, ticket domain.Ticket) {
	if err := s.repo.DeleteTicketPushSubscriptions(ctx, ticket.BusinessID, ticket.ID); err != nil {
		logger.Ctx(ctx).Warn("pager push cleanup failed", logger.Any("ticket_id", ticket.ID), logger.Error(err))
	}
}
