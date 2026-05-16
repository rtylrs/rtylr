package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	webhooksScope = "webhooks"
	webhooksTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Webhook, error)
	Create(ctx context.Context, ownerID int64, url, secret string, events []string) (int64, error)
	Delete(ctx context.Context, ownerID, webhookID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateWebhooksScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(webhooksScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Webhook, error) {
	return cacher.Of[[]domain.Webhook]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(webhooksScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Webhook, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, webhooksTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, url string, events []string) (int64, error) {
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
	id, err := s.repo.Create(ctx, ownerID, url, secret, events)
	if err != nil {
		return 0, err
	}
	invalidateWebhooksScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Delete(ctx context.Context, ownerID, webhookID int64) error {
	if webhookID == 0 {
		return errors.New("webhook id is required")
	}
	if err := s.repo.Delete(ctx, ownerID, webhookID); err != nil {
		return err
	}
	invalidateWebhooksScope(ctx, ownerID)
	return nil
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
