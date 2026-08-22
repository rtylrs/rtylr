package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/api_keys/internal/domain"
	"github.com/voxire/onion/internal/modules/api_keys/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// APIKeyExistsInBusiness reports whether the row exists within the business. The write queries
// in this module are :exec and cannot report a zero-row match, so writes
// pre-check existence to avoid silent success on wrong/foreign ids.
func (r *Repository) APIKeyExistsInBusiness(ctx context.Context, businessID, id int64) bool {
	var exists bool
	if err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM hub_api_keys WHERE id = ? AND business_id = ?)",
		id, businessID,
	).Scan(&exists); err != nil {
		return false
	}
	return exists
}
func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.APIKey, error) {
	rows, err := r.queries.ListAPIKeys(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list API keys", logger.Error(err))
		return nil, err
	}

	keys := make([]domain.APIKey, 0, len(rows))
	for _, row := range rows {
		var k domain.APIKey
		k.ID = row.ID
		k.OwnerID = row.OwnerID
		k.BusinessID = row.BusinessID
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

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, name, keyHash, keyPrefix string, scopes []string) (int64, error) {
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return 0, err
	}

	res, err := r.queries.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Name:       name,
		KeyHash:    keyHash,
		KeyPrefix:  keyPrefix,
		Scopes:     scopesJSON,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create API key", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Revoke(ctx context.Context, businessID, keyID int64) error {
	if !r.APIKeyExistsInBusiness(ctx, businessID, keyID) {
		return domain.ErrAPIKeyNotFound
	}
	err := r.queries.RevokeAPIKey(ctx, sqlc.RevokeAPIKeyParams{
		ID:         keyID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to revoke API key", logger.Error(err))
		return err
	}

	return nil
}
