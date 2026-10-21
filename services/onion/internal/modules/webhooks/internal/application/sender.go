package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	// defaultBatch bounds one pass. The loop also stops when the job's context
	// is done, so this is an upper bound rather than a promise: whatever is
	// left stays due and is taken next tick.
	defaultBatch = 25

	// requestTimeout is per delivery. A subscriber that hangs costs this much
	// and no more, and the loop moves to the next tenant.
	requestTimeout = 10 * time.Second

	baseRetrySeconds = 60
	maxRetrySeconds  = 3600

	// abandonedAfter is how long a claimed delivery may sit before it is
	// treated as abandoned. Far longer than requestTimeout, so a slow
	// subscriber is never mistaken for a dead worker.
	abandonedAfter = 15 * time.Minute

	// errorSnippetLimit bounds how much of a failing response is read. The
	// body is a subscriber's, so it is neither trusted nor stored whole.
	errorSnippetLimit = 512
)

// Sender drains the webhook delivery queue.
//
// It lives in this module rather than beside the dispatcher because
// hub_webhooks, its secret and its last_triggered_at belong here. The queue is
// a platform table, read as an annotated cross-schema read.
type Sender struct {
	repo      deliveryRepository
	events    eventReader
	client    *http.Client
	batchSize int32
}

// NewSender builds a sender over the shared pool.
func NewSender(repo deliveryRepository, events eventReader) *Sender {
	return &Sender{
		repo:      repo,
		events:    events,
		client:    newDeliveryClient(requestTimeout),
		batchSize: defaultBatch,
	}
}

// ProcessDue sends one batch of due deliveries.
//
// Each delivery is claimed and sent on its own, and a failure is recorded
// against that row rather than aborting the pass. One tenant's dead endpoint
// therefore costs one timeout and holds up nobody else.
func (s *Sender) ProcessDue(ctx context.Context) error {
	s.recoverAbandoned(ctx)

	deliveries, err := s.repo.ListDueDeliveries(ctx, s.batchSize)
	if err != nil {
		return fmt.Errorf("listing due webhook deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		// The job has a timeout and each send can take up to requestTimeout,
		// so a large batch of slow endpoints would otherwise run past it and
		// be cancelled mid-send. Stopping here leaves the rest due.
		if ctx.Err() != nil {
			return nil
		}

		if claimErr := s.repo.ClaimDelivery(ctx, delivery.ID); claimErr != nil {
			if errors.Is(claimErr, domain.ErrDeliveryNotClaimed) {
				continue
			}
			logger.Ctx(ctx).Error("webhooks: failed to claim delivery",
				logger.Any("delivery_id", delivery.ID), logger.Error(claimErr))
			continue
		}

		s.deliver(ctx, delivery)
	}
	return nil
}

// recoverAbandoned requeues deliveries left in 'delivering' by a worker that
// died between claiming a row and recording its outcome. Without it those rows
// are never looked at again, because the due query only reads 'pending' and
// 'failed', and the subscriber never hears about that event.
//
// Runs at the start of every pass rather than as a separate job: it is a
// single indexed UPDATE that matches nothing in normal operation, and tying it
// to the same cadence means there is no second schedule to get wrong.
//
// A failure to recover does not stop the pass.
func (s *Sender) recoverAbandoned(ctx context.Context) {
	recovered, err := s.repo.RecoverAbandoned(ctx, int64(abandonedAfter.Seconds()))
	if err != nil {
		logger.Ctx(ctx).Error("webhooks: could not recover abandoned deliveries", logger.Error(err))
		return
	}
	// Warn rather than info: a non-zero count means a worker died mid-send, and
	// each one may reach its subscriber twice.
	if recovered > 0 {
		logger.Ctx(ctx).Warn("webhooks: requeued abandoned deliveries",
			logger.Any("deliveries", recovered))
	}
}

// deliver sends one delivery and records the outcome.
func (s *Sender) deliver(ctx context.Context, delivery domain.Delivery) {
	status, err := s.attempt(ctx, delivery)
	if err == nil {
		if markErr := s.repo.MarkDelivered(ctx, delivery.ID, status); markErr != nil {
			logger.Ctx(ctx).Error("webhooks: delivered but could not be marked",
				logger.Any("delivery_id", delivery.ID), logger.Error(markErr))
		}
		if triggerErr := s.repo.RecordTriggered(ctx, delivery.WebhookID, delivery.BusinessID); triggerErr != nil {
			logger.Ctx(ctx).Error("webhooks: could not record last_triggered_at",
				logger.Any("webhook_id", delivery.WebhookID), logger.Error(triggerErr))
		}
		return
	}

	logger.Ctx(ctx).Error("webhooks: delivery failed",
		logger.Any("delivery_id", delivery.ID),
		logger.Any("webhook_id", delivery.WebhookID),
		logger.Error(err))

	var permanent errUndeliverable
	if errors.As(err, &permanent) {
		if markErr := s.repo.MarkDead(ctx, delivery.ID, err.Error()); markErr != nil {
			logger.Ctx(ctx).Error("webhooks: could not dead-letter delivery",
				logger.Any("delivery_id", delivery.ID), logger.Error(markErr))
		}
		return
	}

	if markErr := s.repo.MarkFailed(ctx, delivery.ID, retryDelaySeconds(delivery.Attempts), status, err.Error()); markErr != nil {
		logger.Ctx(ctx).Error("webhooks: could not reschedule delivery",
			logger.Any("delivery_id", delivery.ID), logger.Error(markErr))
	}
}

// retryDelaySeconds doubles from the base with each attempt, capped.
func retryDelaySeconds(attempts int) int64 {
	delay := int64(baseRetrySeconds)
	for i := 1; i < attempts && delay < maxRetrySeconds; i++ {
		delay *= 2
	}
	if delay > maxRetrySeconds {
		return maxRetrySeconds
	}
	return delay
}

// attempt performs one HTTP send, returning the response status when there
// was one.
func (s *Sender) attempt(ctx context.Context, delivery domain.Delivery) (int, error) {
	registration, err := s.repo.GetRegistrationForDelivery(ctx, delivery.WebhookID, delivery.BusinessID)
	if err != nil {
		if errors.Is(err, domain.ErrDeliveryRegistrationNotFound) {
			return 0, undeliverable("registration %d is not in business %d", delivery.WebhookID, delivery.BusinessID)
		}
		return 0, fmt.Errorf("loading registration: %w", err)
	}
	// Switched off between fan-out and now. Terminal: waiting does not turn a
	// disabled registration back on, and delivering to one would ignore the
	// tenant's decision.
	if !registration.IsActive {
		return 0, undeliverable("registration %d is not active", registration.ID)
	}
	if registration.Secret == "" {
		return 0, undeliverable("registration %d has no signing secret", registration.ID)
	}
	if err := validateDestination(registration.URL); err != nil {
		return 0, err
	}

	fact, err := s.events.GetEvent(ctx, delivery.EventID, delivery.BusinessID)
	if err != nil {
		if errors.Is(err, eventbus.ErrEventNotFound) {
			return 0, undeliverable("event %d is not in business %d", delivery.EventID, delivery.BusinessID)
		}
		return 0, fmt.Errorf("loading event: %w", err)
	}
	event := domain.Event{
		ID:          fact.ID,
		Type:        fact.Type,
		SubjectType: fact.SubjectType,
		SubjectID:   fact.SubjectID,
		Payload:     eventPayloadBytes(fact.Payload),
		OccurredAt:  fact.OccurredAt,
	}

	body, err := encodeEnvelope(event)
	if err != nil {
		return 0, undeliverable("event %d could not be encoded: %v", event.ID, err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, registration.URL, bytes.NewReader(body))
	if err != nil {
		return 0, undeliverable("building request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "rtylr-webhooks/1")
	request.Header.Set(EventTypeHeader, event.Type)
	request.Header.Set(EventIDHeader, strconv.FormatUint(event.ID, 10))
	request.Header.Set(DeliveryIDHeader, strconv.FormatUint(delivery.ID, 10))
	request.Header.Set(SignatureHeader, sign(registration.Secret, time.Now().Unix(), body))

	response, err := s.client.Do(request)
	if err != nil {
		// A refused internal address surfaces here, wrapped by the transport.
		// It is permanent: the registration points somewhere it may not.
		if isInternalAddressRefusal(err) {
			return 0, undeliverable("destination is not reachable from this platform: %v", err)
		}
		return 0, fmt.Errorf("sending: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	snippet, _ := io.ReadAll(io.LimitReader(response.Body, errorSnippetLimit))

	if response.StatusCode < 200 || response.StatusCode > 299 {
		// Every non-2xx is retried until attempts run out, rather than
		// classifying 4xx as permanent. A 404 during a subscriber's deploy is
		// as common as a genuinely wrong URL, and dead-lettering the first
		// would lose events for a transient fault.
		return response.StatusCode, fmt.Errorf("endpoint answered %d: %s", response.StatusCode, string(snippet))
	}
	return response.StatusCode, nil
}

// isInternalAddressRefusal recognises the transport's own refusal, so it is
// classified as permanent rather than retried like a network blip.
func isInternalAddressRefusal(err error) bool {
	return err != nil && containsAny(err.Error(),
		"refusing to connect to internal address",
		"refusing unresolved address",
		"refusing unparseable address",
		"redirects are not followed")
}

func containsAny(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if len(needle) > 0 && bytes.Contains([]byte(haystack), []byte(needle)) {
			return true
		}
	}
	return false
}

// envelope is the delivered document. Ids are strings because an int64 crosses
// the wire as a string throughout this product, and a subscriber parsing one
// as a JSON number would lose precision at the top of the range.
type envelope struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Subject    envelopeSubject `json:"subject"`
	OccurredAt string          `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

type envelopeSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func encodeEnvelope(event domain.Event) ([]byte, error) {
	data := json.RawMessage(event.Payload)
	if len(data) == 0 {
		data = json.RawMessage("{}")
	}
	return json.Marshal(envelope{
		ID:         strconv.FormatUint(event.ID, 10),
		Type:       event.Type,
		Subject:    envelopeSubject{Type: event.SubjectType, ID: strconv.FormatInt(event.SubjectID, 10)},
		OccurredAt: event.OccurredAt.UTC().Format(time.RFC3339),
		Data:       data,
	})
}
