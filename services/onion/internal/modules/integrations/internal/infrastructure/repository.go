package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Integration, error) {
	query := `
		SELECT id, owner_id, provider, status, config, connected_at, created_at
		FROM hub_integrations
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list integrations", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var integrations []domain.Integration
	for rows.Next() {
		var i domain.Integration
		var status string
		var configJSON []byte
		if err := rows.Scan(&i.ID, &i.OwnerID, &i.Provider, &status, &configJSON, &i.ConnectedAt, &i.CreatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan integration row", logger.Error(err))
			return nil, err
		}
		i.Status = domain.IntegrationStatus(status)
		i.Config = json.RawMessage(configJSON)
		integrations = append(integrations, i)
	}

	return integrations, rows.Err()
}

func (r *Repository) Connect(ctx context.Context, ownerID int64, provider string, config json.RawMessage) (int64, error) {
	query := `
		INSERT INTO hub_integrations (owner_id, provider, status, config, connected_at)
		VALUES (?, ?, 'connected', ?, NOW())`

	res, err := r.pool.ExecContext(ctx, query, ownerID, provider, config)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to connect integration", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Disconnect(ctx context.Context, ownerID, integrationID int64) error {
	query := `DELETE FROM hub_integrations WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query, integrationID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to disconnect integration", logger.Error(err))
		return err
	}

	return nil
}
