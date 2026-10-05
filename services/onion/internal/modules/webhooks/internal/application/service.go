package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/eventbus"
)

const (
	webhooksScope = "webhooks"
	webhooksTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Webhook, error)
	Create(ctx context.Context, ownerID, businessID int64, url, secret string, events []string) (int64, error)
	Get(ctx context.Context, businessID, webhookID int64) (domain.Webhook, error)
	Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) error
	Delete(ctx context.Context, businessID, webhookID int64) error
	ListDeliveriesForBusiness(ctx context.Context, businessID int64, filter domain.DeliveryFilter) ([]domain.DeliveryRecord, int64, error)
}

type Service struct {
	repo   repository
	events eventReader
}

func NewService(repo repository, events eventReader) *Service {
	return &Service{repo: repo, events: events}
}

func invalidateWebhooksScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(webhooksScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Webhook, error) {
	return cacher.Of[[]domain.Webhook]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(webhooksScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Webhook, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, webhooksTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, url string, events []string) (int64, error) {
	if url == "" {
		return 0, errors.New("url is required")
	}
	secret, err := generateSecret()
	if err != nil {
		return 0, err
	}
	if events == nil {
		events = []string{}
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, url, secret, events)
	if err != nil {
		return 0, err
	}
	invalidateWebhooksScope(ctx, businessID)
	return id, nil
}

func (s *Service) Delete(ctx context.Context, businessID, webhookID int64) error {
	if webhookID == 0 {
		return errors.New("webhook id is required")
	}
	if err := s.repo.Delete(ctx, businessID, webhookID); err != nil {
		return err
	}
	invalidateWebhooksScope(ctx, businessID)
	return nil
}

func (s *Service) Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) (domain.Webhook, error) {
	if webhookID == 0 {
		return domain.Webhook{}, errors.New("webhook id is required")
	}
	if url == "" {
		return domain.Webhook{}, errors.New("url is required")
	}
	if events == nil {
		events = []string{}
	}
	if err := s.repo.Update(ctx, businessID, webhookID, url, events, isActive); err != nil {
		return domain.Webhook{}, err
	}
	invalidateWebhooksScope(ctx, businessID)

	webhook, err := s.repo.Get(ctx, businessID, webhookID)
	if err != nil {
		return domain.Webhook{}, err
	}
	return webhook, nil
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ListDeliveries returns a page of delivery attempts for a business, most
// recent first, with the total matching the same filter.
//
// Read-only. Seeing a dead delivery is the gap this closes; replaying one is a
// mutation with real effects on somebody's endpoint and is deliberately not
// offered here.
func (s *Service) ListDeliveries(ctx context.Context, businessID int64, filter domain.DeliveryFilter) ([]domain.DeliveryRecord, int64, error) {
	records, total, err := s.repo.ListDeliveriesForBusiness(ctx, businessID, filter)
	if err != nil {
		return nil, 0, err
	}
	for i := range records {
		event, err := s.events.GetEvent(ctx, records[i].EventID, businessID)
		if err != nil {
			if errors.Is(err, eventbus.ErrEventNotFound) {
				continue
			}
			return nil, 0, err
		}
		records[i].EventType = event.Type
	}
	return records, total, nil
}
