package webhooks

import (
	"context"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
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

func (p *hubWebhooksProvider) Create(ctx context.Context, ownerID, businessID int64, url string, events []string) (int64, error) {
	return p.service.Create(ctx, ownerID, businessID, url, events)
}

func (p *hubWebhooksProvider) Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) (hubdeveloper.Webhook, error) {
	webhook, err := p.service.Update(ctx, businessID, webhookID, url, events, isActive)
	if err != nil {
		return hubdeveloper.Webhook{}, err
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
	return p.service.Delete(ctx, businessID, webhookID)
}
