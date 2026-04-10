package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
	"github.com/voxire/onion/internal/modules/api_keys/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.APIKey, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.CreateAPIKeyInput) (int64, string, error) {
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

	id, err := s.repo.Create(ctx, ownerID, input.Name, hash, prefix, scopes)
	if err != nil {
		return 0, "", err
	}

	return id, rawKey, nil
}

func (s *Service) Revoke(ctx context.Context, ownerID, keyID int64) error {
	if keyID == 0 {
		return errors.New("key id is required")
	}
	return s.repo.Revoke(ctx, ownerID, keyID)
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
