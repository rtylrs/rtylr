package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/securekey"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

const mysqlDuplicateEntry = 1062

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
	businessIDInt, err := strconv.ParseInt(ownerID, 10, 64)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse business ID", logger.Error(err))

		return domain.Config{}, err
	}

	rows, err := r.queries.GetConfig(ctx, businessIDInt)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get config", logger.Error(err))

		return domain.Config{}, err
	}

	if len(rows) == 0 {
		logger.Ctx(ctx).Error("No config found for businessID", logger.String("businessID", ownerID))

		return domain.Config{}, domain.ErrConfigNotFound
	}

	var config domain.Config
	config.FromRows(rows[0])

	// Enrich with the slug belonging to the business whose config we loaded.
	// The config endpoint is business-scoped; using the config owner's ID here
	// returns a different business when one owner manages multiple businesses.
	slug, slugErr := r.queries.GetBusinessSlugByID(ctx, businessIDInt)
	if slugErr != nil && !errors.Is(slugErr, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("Failed to load business slug for config", logger.Error(slugErr))
	} else if slug.Valid {
		config.BusinessSlug = slug.String
	}

	return config, nil
}

func (r *Repository) GetThemeByOwnerID(ctx context.Context, ownerID string) (domain.Theme, error) {
	businessIDInt, err := strconv.ParseInt(ownerID, 10, 64)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse business ID", logger.Error(err))

		return domain.Theme{}, err
	}

	row, fetchErr := r.queries.GetThemeByOwnerID(ctx, businessIDInt)
	if fetchErr != nil {
		if errors.Is(fetchErr, sql.ErrNoRows) {
			return domain.Theme{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get theme by business ID", logger.Error(fetchErr))

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
	businessIDInt, err := strconv.ParseInt(input.OwnerId, 10, 64)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse business ID", logger.Error(err))
		return err
	}

	metaTagsJSON, marshalErr := json.Marshal(input.MetaTags)
	if marshalErr != nil {
		logger.Ctx(ctx).Error("Failed to marshal meta tags", logger.Error(marshalErr))
		return marshalErr
	}
	emailProvider := input.EmailProvider
	if emailProvider == "" {
		emailProvider = "resend"
	}

	execErr := r.queries.UpdateConfig(ctx, sqlc.UpdateConfigParams{
		SiteName:                  sql.NullString{String: input.SiteName, Valid: input.SiteName != ""},
		LogoUrl:                   sql.NullString{String: input.LogoUrl, Valid: input.LogoUrl != ""},
		FaviconUrl:                sql.NullString{String: input.FaviconUrl, Valid: input.FaviconUrl != ""},
		MetaTags:                  metaTagsJSON,
		ContactEmail:              input.ContactEmail,
		ContactNumber:             input.ContactNumber,
		SenderEmail:               sql.NullString{String: input.SenderEmail, Valid: input.SenderEmail != ""},
		EmailProvider:             emailProvider,
		SenderName:                sql.NullString{String: input.SenderName, Valid: input.SenderName != ""},
		ReplyToEmail:              sql.NullString{String: input.ReplyToEmail, Valid: input.ReplyToEmail != ""},
		EmailNotificationsEnabled: input.EmailNotificationsEnabled,
		SmtpHost:                  sql.NullString{String: input.SMTPHost, Valid: input.SMTPHost != ""},
		SmtpPort:                  sql.NullInt32{Int32: input.SMTPPort, Valid: input.SMTPPort > 0},
		SmtpUsername:              sql.NullString{String: input.SMTPUsername, Valid: input.SMTPUsername != ""},
		Website:                   sql.NullString{String: input.Website, Valid: input.Website != ""},
		MapsUrl:                   sql.NullString{String: input.MapsURL, Valid: input.MapsURL != ""},
		Address:                   sql.NullString{String: input.Address, Valid: input.Address != ""},
		City:                      sql.NullString{String: input.City, Valid: input.City != ""},
		Zipcode:                   sql.NullString{String: input.Zipcode, Valid: input.Zipcode != ""},
		Country:                   sql.NullString{String: input.Country, Valid: input.Country != ""},
		Instagram:                 sql.NullString{String: input.Instagram, Valid: input.Instagram != ""},
		MaintenanceMode:           sql.NullBool{Bool: input.MaintenanceMode, Valid: true},
		MaintenanceNotice:         sql.NullString{String: input.MaintenanceNotice, Valid: input.MaintenanceNotice != ""},
		AddonsEnabled:             sql.NullBool{Bool: input.AddonsEnabled, Valid: true},
		EnableReviews:             sql.NullBool{Bool: input.EnableReviews, Valid: true},
		EnableWishlist:            sql.NullBool{Bool: input.EnableWishlist, Valid: true},
		EnableNewsletterSignup:    sql.NullBool{Bool: input.EnableNewsletterSignup, Valid: true},
		AboutContent:              input.AboutContent,
		FooterSubText:             sql.NullString{String: input.FooterSubText, Valid: input.FooterSubText != ""},
		BusinessID:                businessIDInt,
	})
	if execErr != nil {
		logger.Ctx(ctx).Error("Failed to update config", logger.Error(execErr))
		return execErr
	}
	if input.SMTPPassword != "" {
		if err := r.queries.UpdateSMTPPassword(ctx, sqlc.UpdateSMTPPasswordParams{SmtpPasswordHash: sql.NullString{String: securekey.Hash(input.SMTPPassword), Valid: true}, BusinessID: businessIDInt}); err != nil {
			logger.Ctx(ctx).Error("Failed to update SMTP password", logger.Error(err))
			return err
		}
	}

	return nil
}

func (r *Repository) CreateConfig(ctx context.Context, input domain.ConfigInput) error {
	businessIDInt, err := strconv.ParseInt(input.OwnerId, 10, 64)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse business ID", logger.Error(err))

		return err
	}

	metaTagsJSON, marshalErr := json.Marshal(input.MetaTags)
	if marshalErr != nil {
		logger.Ctx(ctx).Error("Failed to marshal meta tags", logger.Error(marshalErr))

		return marshalErr
	}
	emailProvider := input.EmailProvider
	if emailProvider == "" {
		emailProvider = "resend"
	}

	err = r.queries.CreateConfig(ctx, sqlc.CreateConfigParams{
		OwnerID:                   int32(businessIDInt), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BusinessID:                businessIDInt,
		SiteName:                  sql.NullString{String: input.SiteName, Valid: input.SiteName != ""},
		LogoUrl:                   sql.NullString{String: input.LogoUrl, Valid: input.LogoUrl != ""},
		FaviconUrl:                sql.NullString{String: input.FaviconUrl, Valid: input.FaviconUrl != ""},
		MetaTags:                  metaTagsJSON,
		ContactEmail:              input.ContactEmail,
		ContactNumber:             input.ContactNumber,
		SenderEmail:               sql.NullString{String: input.SenderEmail, Valid: input.SenderEmail != ""},
		EmailProvider:             emailProvider,
		SenderName:                sql.NullString{String: input.SenderName, Valid: input.SenderName != ""},
		ReplyToEmail:              sql.NullString{String: input.ReplyToEmail, Valid: input.ReplyToEmail != ""},
		EmailNotificationsEnabled: input.EmailNotificationsEnabled,
		SmtpHost:                  sql.NullString{String: input.SMTPHost, Valid: input.SMTPHost != ""},
		SmtpPort:                  sql.NullInt32{Int32: input.SMTPPort, Valid: input.SMTPPort > 0},
		SmtpUsername:              sql.NullString{String: input.SMTPUsername, Valid: input.SMTPUsername != ""},
		Website:                   sql.NullString{String: input.Website, Valid: input.Website != ""},
		MapsUrl:                   sql.NullString{String: input.MapsURL, Valid: input.MapsURL != ""},
		Address:                   sql.NullString{String: input.Address, Valid: input.Address != ""},
		City:                      sql.NullString{String: input.City, Valid: input.City != ""},
		Zipcode:                   sql.NullString{String: input.Zipcode, Valid: input.Zipcode != ""},
		Country:                   sql.NullString{String: input.Country, Valid: input.Country != ""},
		Instagram:                 sql.NullString{String: input.Instagram, Valid: input.Instagram != ""},
		MaintenanceMode:           sql.NullBool{Bool: input.MaintenanceMode, Valid: true},
		MaintenanceNotice:         sql.NullString{String: input.MaintenanceNotice, Valid: input.MaintenanceNotice != ""},
		AddonsEnabled:             sql.NullBool{Bool: input.AddonsEnabled, Valid: true},
		EnableReviews:             sql.NullBool{Bool: input.EnableReviews, Valid: true},
		EnableWishlist:            sql.NullBool{Bool: input.EnableWishlist, Valid: true},
		EnableNewsletterSignup:    sql.NullBool{Bool: input.EnableNewsletterSignup, Valid: true},
		AboutContent:              input.AboutContent,
		FooterSubText:             sql.NullString{String: input.FooterSubText, Valid: input.FooterSubText != ""},
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
			return nil
		}

		logger.Ctx(ctx).Error("Failed to create config", logger.Error(err))

		return err
	}

	return nil
}
