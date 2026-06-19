package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
	"github.com/voxire/onion/internal/modules/api_keys/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	apiKeysScope = "api_keys"
	apiKeysTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.APIKey, error)
	Create(ctx context.Context, ownerID, businessID int64, name, keyHash, keyPrefix string, scopes []string) (int64, error)
	Revoke(ctx context.Context, businessID, keyID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateAPIKeysScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(apiKeysScope, businessID))
	}
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
	rawKey, err := generateAPIKey()
	if err != nil {
		return 0, "", err
	}
	prefix := rawKey[:8]
	hash := hashKey(rawKey)
	scopes := input.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input.Name, hash, prefix, scopes)
	if err != nil {
		return 0, "", err
	}
	invalidateAPIKeysScope(ctx, businessID)
	return id, rawKey, nil
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
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func CreateAPIKeyInputFromRequest(name string, scopes []string) domain.CreateAPIKeyInput {
	return domain.CreateAPIKeyInput{
		Name:   name,
		Scopes: scopes,
	}
}
