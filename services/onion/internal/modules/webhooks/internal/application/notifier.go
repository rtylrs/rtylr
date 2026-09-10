package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	// deadNotifyWithinHours bounds which failures are worth mentioning. A
	// registration that broke months ago and was never fixed still has dead
	// rows; without this it would generate a notification every day forever.
	deadNotifyWithinHours = int64(24)

	// deadNotifyBatch bounds one pass. One notification per failing
	// registration, so this is a ceiling on how many businesses are told in a
	// single run, not on how many events were lost.
	deadNotifyBatch = int32(100)
)

// Notifier tells the owner of a registration that it has stopped receiving.
//
// #1192 made dead deliveries findable. This makes them noticeable, which is
// the difference between an owner who can debug their endpoint and one who has
// to think of looking.
type Notifier struct {
	repo        *infrastructure.Repository
	db          *sql.DB
	withinHours int64
	batchSize   int32
}

// NewNotifier builds a notifier over the shared pool.
func NewNotifier(db *sql.DB) *Notifier {
	return &Notifier{
		repo:        infrastructure.NewRepository(db),
		db:          db,
		withinHours: deadNotifyWithinHours,
		batchSize:   deadNotifyBatch,
	}
}

// NotifyFailing sends at most one message per failing registration.
//
// A registration that cannot be notified is logged and skipped rather than
// aborting the pass, so one unreachable owner does not stop everyone else
// hearing about their own endpoints.
func (n *Notifier) NotifyFailing(ctx context.Context) error {
	failing, err := n.repo.ListFailingRegistrations(ctx, n.withinHours, n.batchSize)
	if err != nil {
		return fmt.Errorf("listing failing registrations: %w", err)
	}

	for _, registration := range failing {
		if ctx.Err() != nil {
			return nil
		}
		n.notifyOne(ctx, registration)
	}
	return nil
}

// notifyOne tells one registration's owner, once a day at most.
func (n *Notifier) notifyOne(ctx context.Context, registration domain.FailingRegistration) {
	recipient, err := n.repo.GetNotificationRecipient(ctx, registration.WebhookID, registration.BusinessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// The creator has left the business, or their membership is no
			// longer active. Deliberately not falling back to the business
			// owner or anybody else: that would send somebody a message about
			// something they never set up. Logged so it is visible that a
			// failing registration has nobody watching it.
			logger.Ctx(ctx).Warn("webhooks: failing registration has no active owner to notify",
				logger.Any("webhook_id", registration.WebhookID),
				logger.Any("business_id", registration.BusinessID),
				logger.Any("dead_deliveries", registration.DeadCount))
			return
		}
		logger.Ctx(ctx).Error("webhooks: could not resolve a notification recipient",
			logger.Any("webhook_id", registration.WebhookID), logger.Error(err))
		return
	}
	if recipient.Email == "" {
		logger.Ctx(ctx).Warn("webhooks: failing registration owner has no email address",
			logger.Any("webhook_id", registration.WebhookID),
			logger.Any("user_id", recipient.UserID))
		return
	}

	// One message per registration per day. The key is what makes this safe to
	// run hourly over a query that keeps returning the same rows: the second
	// and later enqueues in a day are no-ops.
	dedupeKey := fmt.Sprintf("webhook_dead:%d:%s", registration.WebhookID, time.Now().UTC().Format("2006-01-02"))

	if err := emailqueue.Enqueue(ctx, n.db, emailqueue.Job{
		To:        recipient.Email,
		Subject:   "A webhook endpoint has stopped receiving events",
		HTML:      deadDeliveryHTML(registration),
		Text:      deadDeliveryText(registration),
		DedupeKey: dedupeKey,
	}); err != nil {
		logger.Ctx(ctx).Error("webhooks: could not queue a dead-delivery notification",
			logger.Any("webhook_id", registration.WebhookID), logger.Error(err))
	}
}

// deadDeliveryHTML renders the message. The URL is the tenant's own and is
// escaped because it lands in an HTML body, not because it is untrusted.
func deadDeliveryHTML(registration domain.FailingRegistration) string {
	return fmt.Sprintf(
		"<p>%s to your webhook endpoint <code>%s</code> could not be delivered and have been given up on.</p>"+
			"<p>Events are retried several times before this happens, so the endpoint has been unreachable "+
			"or returning errors for a while. Anything already given up on will not be sent again.</p>"+
			"<p>The delivery attempts, including what your endpoint answered, are available from the "+
			"developer settings for this business.</p>",
		html.EscapeString(deliveryCountPhrase(registration.DeadCount)),
		html.EscapeString(registration.URL),
	)
}

func deadDeliveryText(registration domain.FailingRegistration) string {
	return fmt.Sprintf(
		"%s to your webhook endpoint %s could not be delivered and have been given up on.\n\n"+
			"Events are retried several times before this happens, so the endpoint has been unreachable "+
			"or returning errors for a while. Anything already given up on will not be sent again.\n\n"+
			"The delivery attempts, including what your endpoint answered, are available from the "+
			"developer settings for this business.",
		deliveryCountPhrase(registration.DeadCount),
		registration.URL,
	)
}

// deliveryCountPhrase keeps the message readable at one, which is the most
// common case: an endpoint usually starts failing rather than failing in bulk.
func deliveryCountPhrase(count int64) string {
	if count == 1 {
		return "One event"
	}
	return fmt.Sprintf("%d events", count)
}
