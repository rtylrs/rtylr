package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

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

	rows, err := r.queries.GetConfig(ctx, int32(ownerIDInt))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get config", logger.Error(err))

		return domain.Config{}, err
	}

	if len(rows) == 0 {
		logger.Ctx(ctx).Error("No config found for ownerID", logger.String("ownerID", ownerID))

		return domain.Config{}, domain.ErrConfigNotFound
	}

	var config domain.Config
	config.FromRows(rows[0])

	// Enrich with business slug from the businesses table. owner_id on
	// businesses maps to the same numeric user id used by configs, so
	// we can resolve it cheaply with a single targeted query without
	// reaching into the business module.
	var slug sql.NullString
	if err := r.pool.QueryRowContext(ctx,
		`SELECT slug FROM businesses WHERE owner_id = ? LIMIT 1`, ownerIDInt,
	).Scan(&slug); err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("Failed to load business slug for config", logger.Error(err))
	} else if slug.Valid {
		config.BusinessSlug = slug.String
	}

	return config, nil
}

func (r *Repository) GetThemeByOwnerID(ctx context.Context, ownerID string) (domain.Theme, error) {
	ownerIDInt, err := strconv.Atoi(ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse owner ID", logger.Error(err))

		return domain.Theme{}, err
	}

	row, fetchErr := r.queries.GetThemeByOwnerID(ctx, int32(ownerIDInt))
	if fetchErr != nil {
		if errors.Is(fetchErr, sql.ErrNoRows) {
			return domain.Theme{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get theme by owner ID", logger.Error(fetchErr))

		return domain.Theme{}, fetchErr
	}

	theme := domain.Theme{
		ID:                    row.ID,
		PrimaryColor:          convert.ToString(row.PrimaryColor),
		SecondaryColor:        convert.ToString(row.SecondaryColor),
		AccentColor:           convert.ToString(row.AccentColor),
		DarkGreenColor:        convert.ToString(row.DarkGreenColor),
		BackgroundColor:       convert.ToString(row.BackgroundColor),
		BgColor:               convert.ToString(row.BgColor),
		SurfaceColor:          convert.ToString(row.SurfaceColor),
		LightColor:            convert.ToString(row.LightColor),
		DarkColor:             convert.ToString(row.DarkColor),
		ItemBackgroundColor:   convert.ToString(row.ItemBackgroundColor),
		ErrorColor:            convert.ToString(row.ErrorColor),
		SuccessColor:          convert.ToString(row.SuccessColor),
		WarningColor:          convert.ToString(row.WarningColor),
		InfoColor:             convert.ToString(row.InfoColor),
		TextPrimaryColor:      convert.ToString(row.TextPrimaryColor),
		TextSecondaryColor:    convert.ToString(row.TextSecondaryColor),
		SubtextColor:          convert.ToString(row.SubtextColor),
		LightGrayColor:        convert.ToString(row.LightGrayColor),
		ActiveColor:           convert.ToString(row.ActiveColor),
		AdminBackgroundColor:  convert.ToString(row.AdminBackgroundColor),
		AdminPrimaryColor:     convert.ToString(row.AdminPrimaryColor),
		VendorBackgroundColor: convert.ToString(row.VendorBackgroundColor),
		BorderColor:           convert.ToString(row.BorderColor),
		FontFamilyPrimary:     convert.ToString(row.FontFamilyPrimary),
		FontFamilySecondary:   convert.ToString(row.FontFamilySecondary),
		FontSizeBase:          convert.ToString(row.FontSizeBase),
		FontSizeHeading:       convert.ToString(row.FontSizeHeading),
		MobileBreakpoint:      convert.ToString(row.MobileBreakpoint),
	}

	return theme, nil
}

func (r *Repository) UpdateConfig(ctx context.Context, input domain.ConfigInput) error {
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

	_, execErr := r.pool.ExecContext(ctx, `
		UPDATE config SET
			site_name = ?,
			logo_url = ?,
			favicon_url = ?,
			meta_tags = ?,
			contact_email = ?,
			contact_number = ?,
			instagram = ?,
			maintenance_mode = ?,
			maintenance_notice = ?,
			addons_enabled = ?,
			enable_reviews = ?,
			enable_wishlist = ?,
			enable_newsletter_signup = ?,
			about_content = ?,
			footer_sub_text = ?
		WHERE owner_id = ?
	`,
		sql.NullString{String: input.SiteName, Valid: input.SiteName != ""},
		sql.NullString{String: input.LogoUrl, Valid: input.LogoUrl != ""},
		sql.NullString{String: input.FaviconUrl, Valid: input.FaviconUrl != ""},
		metaTagsJSON,
		input.ContactEmail,
		input.ContactNumber,
		sql.NullString{String: input.Instagram, Valid: input.Instagram != ""},
		input.MaintenanceMode,
		sql.NullString{String: input.MaintenanceNotice, Valid: input.MaintenanceNotice != ""},
		input.AddonsEnabled,
		input.EnableReviews,
		input.EnableWishlist,
		input.EnableNewsletterSignup,
		input.AboutContent,
		sql.NullString{String: input.FooterSubText, Valid: input.FooterSubText != ""},
		ownerIDInt,
	)
	if execErr != nil {
		logger.Ctx(ctx).Error("Failed to update config", logger.Error(execErr))
		return execErr
	}

	return nil
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
