package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	"github.com/voxire/onion/internal/modules/integrations/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Integration, error) {
	rows, err := r.queries.ListIntegrations(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list integrations", logger.Error(err))
		return nil, err
	}

	var integrations []domain.Integration
	for _, row := range rows {
		var i domain.Integration
		i.ID = row.ID
		i.OwnerID = row.OwnerID
		i.Provider = row.Provider
		i.Status = domain.IntegrationStatus(row.Status)
		i.Config = row.Config
		if row.ConnectedAt.Valid {
			i.ConnectedAt = &row.ConnectedAt.Time
		}
		if row.CreatedAt.Valid {
			i.CreatedAt = row.CreatedAt.Time
		}
		integrations = append(integrations, i)
	}

	return integrations, nil
}

func (r *Repository) Connect(ctx context.Context, ownerID int64, provider string, config json.RawMessage) (int64, error) {
	res, err := r.queries.ConnectIntegration(ctx, sqlc.ConnectIntegrationParams{
		OwnerID:  ownerID,
		Provider: provider,
		Config:   config,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to connect integration", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Disconnect(ctx context.Context, ownerID, integrationID int64) error {
	err := r.queries.DisconnectIntegration(ctx, sqlc.DisconnectIntegrationParams{
		ID:      integrationID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to disconnect integration", logger.Error(err))
		return err
	}

	return nil
}
