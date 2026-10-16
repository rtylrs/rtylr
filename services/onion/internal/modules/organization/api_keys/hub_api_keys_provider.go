package apikeys

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/application"
	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/domain"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
)

type hubAPIKeysProvider struct {
	service *application.Service
}

func NewHubAPIKeysProvider(service *application.Service) hubdeveloper.APIKeysProvider {
	return &hubAPIKeysProvider{service: service}
}

func (p *hubAPIKeysProvider) List(ctx context.Context, businessID int64) ([]hubdeveloper.APIKey, error) {
	keys, err := p.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}

	out := make([]hubdeveloper.APIKey, 0, len(keys))
	for _, k := range keys {
		scopes := k.Scopes
		if scopes == nil {
			scopes = []string{}
		}
		out = append(out, hubdeveloper.APIKey{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			Scopes:     scopes,
			IsActive:   k.IsActive,
			LastUsedAt: k.LastUsedAt,
			ExpiresAt:  k.ExpiresAt,
			CreatedAt:  k.CreatedAt,
		})
	}
	return out, nil
}

func (p *hubAPIKeysProvider) Create(ctx context.Context, userID, businessID int64, name string, scopes []string) (int64, string, error) {
	id, key, err := p.service.Create(ctx, userID, businessID, application.CreateAPIKeyInputFromRequest(name, scopes))
	if errors.Is(err, domain.ErrInvalid) {
		return 0, "", fmt.Errorf("%w: %s", hubdeveloper.ErrInvalidAPIKey, strings.TrimPrefix(err.Error(), domain.ErrInvalid.Error()+": "))
	}
	return id, key, err
}

func (p *hubAPIKeysProvider) Revoke(ctx context.Context, businessID, keyID int64) error {
	err := p.service.Revoke(ctx, businessID, keyID)
	if errors.Is(err, domain.ErrNotFound) {
		return hubdeveloper.ErrAPIKeyNotFound
	}
	return err
}
