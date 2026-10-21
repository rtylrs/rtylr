package webhooks

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
)

type hubWebhooksProvider struct {
	service *application.Service
}

func NewHubWebhooksProvider(service *application.Service) hubdeveloper.WebhooksProvider {
	return &hubWebhooksProvider{service: service}
}

func (p *hubWebhooksProvider) List(ctx context.Context, businessID int64) ([]hubdeveloper.Webhook, error) {
	webhooks, err := p.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}

	out := make([]hubdeveloper.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		events := w.Events
		if events == nil {
			events = []string{}
		}
		out = append(out, hubdeveloper.Webhook{
			ID:              w.ID,
			URL:             w.URL,
			Events:          events,
			IsActive:        w.IsActive,
			LastTriggeredAt: w.LastTriggeredAt,
			CreatedAt:       w.CreatedAt,
		})
	}
	return out, nil
}

func (p *hubWebhooksProvider) Create(ctx context.Context, ownerID, businessID int64, url string, events []string) (int64, string, error) {
	return p.service.Create(ctx, ownerID, businessID, url, events)
}

func (p *hubWebhooksProvider) RotateSecret(ctx context.Context, businessID, webhookID int64) (string, error) {
	secret, err := p.service.RotateSecret(ctx, businessID, webhookID)
	if err != nil {
		return "", publicError(err)
	}
	return secret, nil
}

func (p *hubWebhooksProvider) Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) (hubdeveloper.Webhook, error) {
	webhook, err := p.service.Update(ctx, businessID, webhookID, url, events, isActive)
	if err != nil {
		return hubdeveloper.Webhook{}, publicError(err)
	}

	eventsOut := webhook.Events
	if eventsOut == nil {
		eventsOut = []string{}
	}

	return hubdeveloper.Webhook{
		ID:              webhook.ID,
		URL:             webhook.URL,
		Events:          eventsOut,
		IsActive:        webhook.IsActive,
		LastTriggeredAt: webhook.LastTriggeredAt,
		CreatedAt:       webhook.CreatedAt,
	}, nil
}

func (p *hubWebhooksProvider) Delete(ctx context.Context, businessID, webhookID int64) error {
	return publicError(p.service.Delete(ctx, businessID, webhookID))
}

func publicError(err error) error {
	if errors.Is(err, domain.ErrWebhookNotFound) {
		return hubdeveloper.ErrWebhookNotFound
	}
	return err
}

// ListDeliveries adapts the module's delivery listing to the platform port.
func (p *hubWebhooksProvider) ListDeliveries(ctx context.Context, businessID int64, filter hubdeveloper.DeliveryFilter) ([]hubdeveloper.Delivery, int64, error) {
	records, total, err := p.service.ListDeliveries(ctx, businessID, domain.DeliveryFilter{
		WebhookID: filter.WebhookID,
		Status:    filter.Status,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]hubdeveloper.Delivery, 0, len(records))
	for _, record := range records {
		out = append(out, hubdeveloper.Delivery{
			// Narrowed from uint64. Both are auto-increment keys, so the
			// range that would overflow is not reachable in practice, and the
			// contract declares them int64.
			ID:             int64(record.ID), // #nosec G115 -- auto-increment key, see above.
			WebhookID:      record.WebhookID,
			EventID:        int64(record.EventID), // #nosec G115 -- auto-increment key, see above.
			EventType:      record.EventType,
			Status:         record.Status,
			Attempts:       record.Attempts,
			MaxAttempts:    record.MaxAttempts,
			ResponseStatus: record.ResponseStatus,
			LastError:      record.LastError,
			NextAttemptAt:  record.NextAttemptAt,
			DeliveredAt:    record.DeliveredAt,
			CreatedAt:      record.CreatedAt,
		})
	}
	return out, total, nil
}
