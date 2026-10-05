package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/api_keys/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/application"
	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hub api keys manage", App: authorization.AppHub, Permission: "hub:api_keys:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewDeveloperAPI(&directAPIKeysProvider{service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}, nil).ListApiKeys(ctx, nil)
			return response, err
		}},
	})
}

type directAPIKeysProvider struct {
	service *application.Service
}

func (p *directAPIKeysProvider) List(ctx context.Context, businessID int64) ([]hubdeveloper.APIKey, error) {
	keys, err := p.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]hubdeveloper.APIKey, 0, len(keys))
	for _, key := range keys {
		out = append(out, hubdeveloper.APIKey{
			ID:         key.ID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scopes:     key.Scopes,
			IsActive:   key.IsActive,
			LastUsedAt: key.LastUsedAt,
			ExpiresAt:  key.ExpiresAt,
			CreatedAt:  key.CreatedAt,
		})
	}
	return out, nil
}

func (p *directAPIKeysProvider) Create(ctx context.Context, userID, businessID int64, name string, scopes []string) (int64, string, error) {
	return p.service.Create(ctx, userID, businessID, domain.CreateAPIKeyInput{Name: name, Scopes: scopes})
}

func (p *directAPIKeysProvider) Revoke(ctx context.Context, businessID, keyID int64) error {
	return p.service.Revoke(ctx, businessID, keyID)
}
