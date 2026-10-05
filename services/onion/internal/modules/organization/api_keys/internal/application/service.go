package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/securekey"
)

const (
	apiKeysScope = "api_keys"
	apiKeysTTL   = 5 * time.Minute
	// #nosec G101 -- This is a public display prefix for generated API keys, not a secret.
	apiKeyPrefix = "rtylr_api_k1_"
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.APIKey, error)
	Create(ctx context.Context, ownerID, businessID int64, name, keyHash, keyPrefix string, scopes []string) (int64, error)
	Revoke(ctx context.Context, businessID, keyID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func invalidateAPIKeysScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(apiKeysScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.APIKey, error) {
	return cacher.Of[[]domain.APIKey]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(apiKeysScope, businessID, "list"),
		func(ctx context.Context) ([]domain.APIKey, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, apiKeysTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.CreateAPIKeyInput) (int64, string, error) {
	if input.Name == "" {
		return 0, "", errors.New("name is required")
	}
	secret, err := securekey.Generate(securekey.Options{
		Prefix:           apiKeyPrefix,
		DisplayPrefixLen: 16,
	})
	if err != nil {
		return 0, "", err
	}
	scopes := input.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input.Name, secret.Hash, secret.DisplayPrefix, scopes)
	if err != nil {
		return 0, "", err
	}
	invalidateAPIKeysScope(ctx, businessID)
	return id, secret.Plaintext, nil
}

func (s *Service) Revoke(ctx context.Context, businessID, keyID int64) error {
	if keyID == 0 {
		return errors.New("key id is required")
	}
	if err := s.repo.Revoke(ctx, businessID, keyID); err != nil {
		return err
	}
	invalidateAPIKeysScope(ctx, businessID)
	return nil
}

func generateAPIKey() (string, error) {
	secret, err := securekey.Generate(securekey.Options{
		Prefix:           apiKeyPrefix,
		DisplayPrefixLen: 16,
	})
	if err != nil {
		return "", err
	}
	return secret.Plaintext, nil
}

func hashKey(key string) string {
	return securekey.Hash(key)
}

func CreateAPIKeyInputFromRequest(name string, scopes []string) domain.CreateAPIKeyInput {
	return domain.CreateAPIKeyInput{
		Name:   name,
		Scopes: scopes,
	}
}
