package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
	"github.com/voxire/onion/internal/modules/api_keys/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.APIKey, error) {
	rows, err := r.queries.ListAPIKeys(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list API keys", logger.Error(err))
		return nil, err
	}

	var keys []domain.APIKey
	for _, row := range rows {
		var k domain.APIKey
		k.ID = row.ID
		k.OwnerID = row.OwnerID
		k.Name = row.Name
		k.KeyPrefix = row.KeyPrefix
		k.IsActive = row.IsActive
		if row.LastUsedAt.Valid {
			k.LastUsedAt = &row.LastUsedAt.Time
		}
		if row.ExpiresAt.Valid {
			k.ExpiresAt = &row.ExpiresAt.Time
		}
		if row.CreatedAt.Valid {
			k.CreatedAt = row.CreatedAt.Time
		}
		if err := json.Unmarshal(row.Scopes, &k.Scopes); err != nil {
			k.Scopes = []string{}
		}
		keys = append(keys, k)
	}

	return keys, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, name, keyHash, keyPrefix string, scopes []string) (int64, error) {
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return 0, err
	}

	res, err := r.queries.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		OwnerID:   ownerID,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scopes:    scopesJSON,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create API key", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Revoke(ctx context.Context, ownerID, keyID int64) error {
	err := r.queries.RevokeAPIKey(ctx, sqlc.RevokeAPIKeyParams{
		ID:      keyID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to revoke API key", logger.Error(err))
		return err
	}

	return nil
}
