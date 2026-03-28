package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetConfig(ctx context.Context, ownerID string) (domain.Config, error) {
	ownerIDInt, err := strconv.Atoi(ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse owner ID", logger.Error(err))

		return domain.Config{}, err
	}

	configs, err := db.NewReadExecutor[domain.Config, sqlc.GetConfigRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetConfig, int32(ownerIDInt))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get config", logger.Error(err))

		return domain.Config{}, err
	}

	if len(configs) == 0 {
		logger.Ctx(ctx).Error("No config found for ownerID", logger.String("ownerID", ownerID))

		return domain.Config{}, err
	}

	return configs[0], nil
}

func (r *Repository) GetThemeByOwnerID(ctx context.Context, ownerID string) (domain.Theme, error) {
	ownerIDInt, err := strconv.Atoi(ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse owner ID", logger.Error(err))

		return domain.Theme{}, err
	}

	theme, fetchErr := db.NewReadExecutor[domain.Theme, sqlc.Theme](db.MainDatabase).Fetch(ctx, r.queries.GetThemeByOwnerID, int32(ownerIDInt))
	if fetchErr != nil {
		if errors.Is(fetchErr, sql.ErrNoRows) {
			return domain.Theme{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get theme by owner ID", logger.Error(fetchErr))

		return domain.Theme{}, fetchErr
	}

	return theme, nil
}

func (r *Repository) CreateConfig(ctx context.Context, input domain.ConfigInput) error {
	ownerIDInt, err := strconv.Atoi(input.OwnerId)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse owner ID", logger.Error(err))

		return err
	}

	metaTagsJSON, marshalErr := json.Marshal(input.MetaTags)
	if marshalErr != nil {
		logger.Ctx(ctx).Error("Failed to marshal meta tags", logger.Error(marshalErr))

		return marshalErr
	}

	err = r.queries.CreateConfig(ctx, sqlc.CreateConfigParams{
		OwnerID:                int32(ownerIDInt),
		SiteName:               sql.NullString{String: input.SiteName, Valid: input.SiteName != ""},
		LogoUrl:                sql.NullString{String: input.LogoUrl, Valid: input.LogoUrl != ""},
		FaviconUrl:             sql.NullString{String: input.FaviconUrl, Valid: input.FaviconUrl != ""},
		MetaTags:               metaTagsJSON,
		ContactEmail:           input.ContactEmail,
		ContactNumber:          input.ContactNumber,
		Instagram:              sql.NullString{String: input.Instagram, Valid: input.Instagram != ""},
		MaintenanceMode:        sql.NullBool{Bool: input.MaintenanceMode, Valid: true},
		MaintenanceNotice:      sql.NullString{String: input.MaintenanceNotice, Valid: input.MaintenanceNotice != ""},
		AddonsEnabled:          sql.NullBool{Bool: input.AddonsEnabled, Valid: true},
		EnableReviews:          sql.NullBool{Bool: input.EnableReviews, Valid: true},
		EnableWishlist:         sql.NullBool{Bool: input.EnableWishlist, Valid: true},
		EnableNewsletterSignup: sql.NullBool{Bool: input.EnableNewsletterSignup, Valid: true},
		AboutContent:           input.AboutContent,
		FooterSubText:          sql.NullString{String: input.FooterSubText, Valid: input.FooterSubText != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create config", logger.Error(err))

		return err
	}

	return nil
}
