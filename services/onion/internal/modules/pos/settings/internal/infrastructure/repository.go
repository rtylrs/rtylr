package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/pos/settings/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/settings/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db), db: db}
}

func (r *Repository) GetSettings(ctx context.Context, businessID int64) (domain.Settings, error) {
	row, err := r.queries.GetPOSSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Settings{}, domain.ErrSettingsNotFound
		}
		logger.Ctx(ctx).Error("failed to get POS business settings", logger.Error(err))
		return domain.Settings{}, err
	}

	profile := string(row.BusinessProfile)
	theme := string(row.Theme)
	features := featureRecord(domain.DefaultFeatures(domain.NormalizeProfile(profile)))
	if len(row.Features) > 0 {
		if err := json.Unmarshal(row.Features, &features); err != nil {
			logger.Ctx(ctx).Warn("failed to decode POS feature settings",
				logger.Any("business_id", businessID),
				logger.Error(err))
		}
	}

	settings := domain.Settings{
		ID:            row.ID,
		BusinessID:    row.BusinessID,
		Profile:       profile,
		Theme:         theme,
		Features:      domain.FeatureSettings(features),
		ReceiptFooter: nullStr(row.ReceiptFooter),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	return domain.NormalizeSettings(settings), nil
}

func (r *Repository) UpsertSettings(ctx context.Context, settings domain.Settings) error {
	featuresJSON, err := json.Marshal(featureRecord(settings.Features))
	if err != nil {
		return fmt.Errorf("marshal POS settings features: %w", err)
	}

	err = r.queries.UpsertPOSSettings(ctx, sqlc.UpsertPOSSettingsParams{
		BusinessID:      settings.BusinessID,
		BusinessProfile: sqlc.PosBusinessSettingsBusinessProfile(settings.Profile),
		Theme:           sqlc.PosBusinessSettingsTheme(settings.Theme),
		Features:        featuresJSON,
		ReceiptFooter:   sql.NullString{String: settings.ReceiptFooter, Valid: settings.ReceiptFooter != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert POS business settings", logger.Error(err))
		return err
	}
	return nil
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}
