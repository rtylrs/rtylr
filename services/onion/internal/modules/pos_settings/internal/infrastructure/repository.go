package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetSettings(ctx context.Context, businessID int64) (domain.Settings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			business_id,
			business_profile,
			theme,
			features,
			receipt_footer,
			created_at,
			updated_at
		FROM pos_business_settings
		WHERE business_id = ?
		LIMIT 1
	`, businessID)

	var settings domain.Settings
	var featuresRaw []byte
	var receiptFooter sql.NullString
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	if err := row.Scan(
		&settings.ID,
		&settings.BusinessID,
		&settings.Profile,
		&settings.Theme,
		&featuresRaw,
		&receiptFooter,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Settings{}, domain.ErrSettingsNotFound
		}
		logger.Ctx(ctx).Error("failed to get POS business settings", logger.Error(err))
		return domain.Settings{}, err
	}

	features := domain.DefaultFeatures(domain.NormalizeProfile(settings.Profile))
	if len(featuresRaw) > 0 {
		if err := json.Unmarshal(featuresRaw, &features); err != nil {
			logger.Ctx(ctx).Warn("failed to decode POS feature settings",
				logger.Any("business_id", businessID),
				logger.Error(err))
		}
	}
	settings.Features = features
	if receiptFooter.Valid {
		settings.ReceiptFooter = receiptFooter.String
	}
	settings.CreatedAt = timeOrZero(createdAt)
	settings.UpdatedAt = timeOrZero(updatedAt)
	return domain.NormalizeSettings(settings), nil
}

func (r *Repository) UpsertSettings(ctx context.Context, settings domain.Settings) error {
	featuresJSON, err := json.Marshal(settings.Features)
	if err != nil {
		return fmt.Errorf("marshal POS settings features: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO pos_business_settings (
			business_id,
			business_profile,
			theme,
			features,
			receipt_footer
		) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			business_profile = VALUES(business_profile),
			theme = VALUES(theme),
			features = VALUES(features),
			receipt_footer = VALUES(receipt_footer),
			updated_at = CURRENT_TIMESTAMP
	`, settings.BusinessID, settings.Profile, settings.Theme, string(featuresJSON), settings.ReceiptFooter)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert POS business settings", logger.Error(err))
		return err
	}
	return nil
}

func timeOrZero(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}
