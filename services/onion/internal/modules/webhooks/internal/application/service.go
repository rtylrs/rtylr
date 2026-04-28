package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
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

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Webhook, error) {
	return s.repo.List(ctx, ownerID)
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

	return s.repo.Create(ctx, ownerID, url, secret, events)
}

func (s *Service) Delete(ctx context.Context, ownerID, webhookID int64) error {
	if webhookID == 0 {
		return errors.New("webhook id is required")
	}
	return s.repo.Delete(ctx, ownerID, webhookID)
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
