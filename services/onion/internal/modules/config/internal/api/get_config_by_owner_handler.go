package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ConfigByOwnerRoutes() []string {
	return []string{
		"/get-configs",
	}
}

type GetConfigByOwnerHandler struct {
	ConfigService *application.Service
}

type GetConfigByOwnerRequest struct {
	Owner string `json:"owner"`
}

type ThemePayload struct {
	ID                    int64  `json:"id"`
	PrimaryColor          string `json:"primary_color"`
	SecondaryColor        string `json:"secondary_color"`
	AccentColor           string `json:"accent_color"`
	DarkGreenColor        string `json:"dark_green_color"`
	BackgroundColor       string `json:"background_color"`
	BgColor               string `json:"bg_color"`
	SurfaceColor          string `json:"surface_color"`
	LightColor            string `json:"light_color"`
	DarkColor             string `json:"dark_color"`
	ItemBackgroundColor   string `json:"item_background_color"`
	ErrorColor            string `json:"error_color"`
	SuccessColor          string `json:"success_color"`
	WarningColor          string `json:"warning_color"`
	InfoColor             string `json:"info_color"`
	TextPrimaryColor      string `json:"text_primary_color"`
	TextSecondaryColor    string `json:"text_secondary_color"`
	SubtextColor          string `json:"subtext_color"`
	LightGrayColor        string `json:"light_gray_color"`
	ActiveColor           string `json:"active_color"`
	AdminBackgroundColor  string `json:"admin_background_color"`
	AdminPrimaryColor     string `json:"admin_primary_color"`
	VendorBackgroundColor string `json:"vendor_background_color"`
	BorderColor           string `json:"border_color"`
	FontFamilyPrimary     string `json:"font_family_primary"`
	FontFamilySecondary   string `json:"font_family_secondary"`
	FontSizeBase          string `json:"font_size_base"`
	FontSizeHeading       string `json:"font_size_heading"`
	MobileBreakpoint      string `json:"mobile_breakpoint"`
}

type ConfigPayload struct {
	OwnerID                   string   `json:"owner_id"`
	SiteName                  string   `json:"site_name"`
	LogoURL                   string   `json:"logo_url"`
	FaviconURL                string   `json:"favicon_url"`
	MetaTags                  []string `json:"meta_tags"`
	ContactEmail              string   `json:"contact_email"`
	ContactNumber             string   `json:"contact_number"`
	SenderEmail               string   `json:"sender_email"`
	EmailProvider             string   `json:"email_provider"`
	SenderName                string   `json:"sender_name"`
	ReplyToEmail              string   `json:"reply_to_email"`
	EmailNotificationsEnabled bool     `json:"email_notifications_enabled"`
	SMTPHost                  string   `json:"smtp_host"`
	SMTPPort                  int32    `json:"smtp_port"`
	SMTPUsername              string   `json:"smtp_username"`
	SMTPPasswordConfigured    bool     `json:"smtp_password_configured"`
	Website                   string   `json:"website"`
	MapsURL                   string   `json:"maps_url"`
	Address                   string   `json:"address"`
	City                      string   `json:"city"`
	Zipcode                   string   `json:"zipcode"`
	Country                   string   `json:"country"`
	Instagram                 string   `json:"instagram"`
	MaintenanceMode           bool     `json:"maintenance_mode"`
	MaintenanceNotice         string   `json:"maintenance_notice"`
	AddonsEnabled             bool     `json:"addons_enabled"`
	EnableReviews             bool     `json:"enable_reviews"`
	EnableWishlist            bool     `json:"enable_wishlist"`
	EnableNewsletterSignup    bool     `json:"enable_newsletter_signup"`
	AboutContent              string   `json:"about_content"`
	FooterSubText             string   `json:"footer_sub_text"`
	Slug                      string   `json:"slug"`
}

type GetConfigByOwnerResponse struct {
	Config *ConfigPayload `json:"config"`
	Theme  *ThemePayload  `json:"theme"`
}

func newThemePayload(theme domain.Theme) *ThemePayload {
	return &ThemePayload{
		ID:                    theme.ID,
		PrimaryColor:          theme.PrimaryColor,
		SecondaryColor:        theme.SecondaryColor,
		AccentColor:           theme.AccentColor,
		DarkGreenColor:        theme.DarkGreenColor,
		BackgroundColor:       theme.BackgroundColor,
		BgColor:               theme.BgColor,
		SurfaceColor:          theme.SurfaceColor,
		LightColor:            theme.LightColor,
		DarkColor:             theme.DarkColor,
		ItemBackgroundColor:   theme.ItemBackgroundColor,
		ErrorColor:            theme.ErrorColor,
		SuccessColor:          theme.SuccessColor,
		WarningColor:          theme.WarningColor,
		InfoColor:             theme.InfoColor,
		TextPrimaryColor:      theme.TextPrimaryColor,
		TextSecondaryColor:    theme.TextSecondaryColor,
		SubtextColor:          theme.SubtextColor,
		LightGrayColor:        theme.LightGrayColor,
		ActiveColor:           theme.ActiveColor,
		AdminBackgroundColor:  theme.AdminBackgroundColor,
		AdminPrimaryColor:     theme.AdminPrimaryColor,
		VendorBackgroundColor: theme.VendorBackgroundColor,
		BorderColor:           theme.BorderColor,
		FontFamilyPrimary:     theme.FontFamilyPrimary,
		FontFamilySecondary:   theme.FontFamilySecondary,
		FontSizeBase:          theme.FontSizeBase,
		FontSizeHeading:       theme.FontSizeHeading,
		MobileBreakpoint:      theme.MobileBreakpoint,
	}
}

func newConfigPayload(config domain.Config) *ConfigPayload {
	return &ConfigPayload{
		OwnerID:                   config.OwnerId,
		SiteName:                  config.SiteName,
		LogoURL:                   config.LogoUrl,
		FaviconURL:                config.FaviconUrl,
		MetaTags:                  config.MetaTags,
		ContactEmail:              config.ContactEmail,
		ContactNumber:             config.ContactNumber,
		SenderEmail:               config.SenderEmail,
		EmailProvider:             config.EmailProvider,
		SenderName:                config.SenderName,
		ReplyToEmail:              config.ReplyToEmail,
		EmailNotificationsEnabled: config.EmailNotificationsEnabled,
		SMTPHost:                  config.SMTPHost,
		SMTPPort:                  config.SMTPPort,
		SMTPUsername:              config.SMTPUsername,
		SMTPPasswordConfigured:    config.SMTPPasswordConfigured,
		Website:                   config.Website,
		MapsURL:                   config.MapsURL,
		Address:                   config.Address,
		City:                      config.City,
		Zipcode:                   config.Zipcode,
		Country:                   config.Country,
		Instagram:                 config.Instagram,
		MaintenanceMode:           config.MaintenanceMode,
		MaintenanceNotice:         config.MaintenanceNotice,
		AddonsEnabled:             config.AddonsEnabled,
		EnableReviews:             config.EnableReviews,
		EnableWishlist:            config.EnableWishlist,
		EnableNewsletterSignup:    config.EnableNewsletterSignup,
		AboutContent:              config.AboutContent,
		FooterSubText:             config.FooterSubText,
		Slug:                      config.BusinessSlug,
	}
}

func (h *GetConfigByOwnerHandler) Validate(_ context.Context, _ *GetConfigByOwnerRequest) error {
	return nil
}

func parseRequestedBusinessID(req *GetConfigByOwnerRequest) (int64, bool, error) {
	if req == nil || strings.TrimSpace(req.Owner) == "" {
		return 0, false, nil
	}

	businessID, err := strconv.ParseInt(strings.TrimSpace(req.Owner), 10, 64)
	if err != nil || businessID <= 0 {
		return 0, false, tenant.ErrNoBusinessScope
	}

	return businessID, true, nil
}

func resolveOwnedConfigBusinessID(ctx context.Context, ownerID, requestedBusinessID int64) (int64, error) {
	return tenant.ResolveOwnedConfigBusinessID(ctx, ownerID, requestedBusinessID)
}

func resolveConfigBusinessID(ctx context.Context, authUser *authPkg.AuthUser, req *GetConfigByOwnerRequest) (int64, error) {
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr == nil {
		return businessID, nil
	}

	requestedBusinessID, hasRequestedBusinessID, parseErr := parseRequestedBusinessID(req)
	if parseErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}

	if strings.EqualFold(authUser.Role, "console_admin") {
		if !hasRequestedBusinessID {
			return 0, onekiterrors.PermissionDenied()
		}
		return requestedBusinessID, nil
	}

	if !strings.EqualFold(authUser.Role, "owner") || authUser.UserID <= 0 {
		return 0, onekiterrors.PermissionDenied()
	}

	resolvedBusinessID, err := resolveOwnedConfigBusinessID(ctx, authUser.UserID, requestedBusinessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, tenant.ErrNoBusinessScope) {
			return 0, onekiterrors.PermissionDenied()
		}
		return 0, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve business")
	}

	return resolvedBusinessID, nil
}

func (h *GetConfigByOwnerHandler) Serve(ctx context.Context, req *GetConfigByOwnerRequest) (GetConfigByOwnerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	authUser := device.User
	if authUser == nil {
		return GetConfigByOwnerResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, resolveErr := resolveConfigBusinessID(ctx, authUser, req)
	if resolveErr != nil {
		return GetConfigByOwnerResponse{}, resolveErr
	}

	businessIDStr := strconv.FormatInt(businessID, 10)

	config, err := h.ConfigService.GetConfig(ctx, businessIDStr)
	if err != nil {
		return GetConfigByOwnerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get config")
	}

	theme, err := h.ConfigService.GetThemeByOwnerID(ctx, businessIDStr)
	if err != nil {
		return GetConfigByOwnerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get theme")
	}

	return GetConfigByOwnerResponse{
		Config: newConfigPayload(config),
		Theme:  newThemePayload(theme),
	}, nil
}
