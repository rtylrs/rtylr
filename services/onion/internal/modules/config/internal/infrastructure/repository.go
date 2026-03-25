package infrastructure

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
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
		if errors.Is(fetchErr, pgx.ErrNoRows) {
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

	_, err = r.queries.CreateConfig(ctx, sqlc.CreateConfigParams{
		OwnerID:                int32(ownerIDInt),
		SiteName:               pgtype.Text{String: input.SiteName, Valid: true},
		LogoUrl:                pgtype.Text{String: input.LogoUrl, Valid: true},
		FaviconUrl:             pgtype.Text{String: input.FaviconUrl, Valid: true},
		MetaTags:               input.MetaTags,
		ContactEmail:           input.ContactEmail,
		ContactNumber:          input.ContactNumber,
		Instagram:              pgtype.Text{String: input.Instagram, Valid: true},
		MaintenanceMode:        pgtype.Bool{Bool: input.MaintenanceMode, Valid: true},
		MaintenanceNotice:      pgtype.Text{String: input.MaintenanceNotice, Valid: true},
		AddonsEnabled:          pgtype.Bool{Bool: input.AddonsEnabled, Valid: true},
		EnableReviews:          pgtype.Bool{Bool: input.EnableReviews, Valid: true},
		EnableWishlist:         pgtype.Bool{Bool: input.EnableWishlist, Valid: true},
		EnableNewsletterSignup: pgtype.Bool{Bool: input.EnableNewsletterSignup, Valid: true},
		AboutContent:           input.AboutContent,
		FooterSubText:          pgtype.Text{String: input.FooterSubText, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create config", logger.Error(err))

		return err
	}

	return nil
}
