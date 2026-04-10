package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.APIKey, error) {
	query := `
		SELECT id, owner_id, name, key_prefix, scopes, last_used_at, expires_at, is_active, created_at
		FROM hub_api_keys
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list API keys", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		var scopesJSON []byte
		if err := rows.Scan(&k.ID, &k.OwnerID, &k.Name, &k.KeyPrefix, &scopesJSON, &k.LastUsedAt, &k.ExpiresAt, &k.IsActive, &k.CreatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan API key row", logger.Error(err))
			return nil, err
		}
		if err := json.Unmarshal(scopesJSON, &k.Scopes); err != nil {
			k.Scopes = []string{}
		}
		keys = append(keys, k)
	}

	return keys, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, name, keyHash, keyPrefix string, scopes []string) (int64, error) {
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO hub_api_keys (owner_id, name, key_hash, key_prefix, scopes)
		VALUES (?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query, ownerID, name, keyHash, keyPrefix, scopesJSON)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create API key", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Revoke(ctx context.Context, ownerID, keyID int64) error {
	query := `
		UPDATE hub_api_keys
		SET is_active = FALSE
		WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query, keyID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to revoke API key", logger.Error(err))
		return err
	}

	return nil
}
