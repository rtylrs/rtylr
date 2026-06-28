package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	OwnerID                string   `json:"owner_id"`
	SiteName               string   `json:"site_name"`
	LogoURL                string   `json:"logo_url"`
	FaviconURL             string   `json:"favicon_url"`
	MetaTags               []string `json:"meta_tags"`
	ContactEmail           string   `json:"contact_email"`
	ContactNumber          string   `json:"contact_number"`
	SenderEmail            string   `json:"sender_email"`
	Website                string   `json:"website"`
	MapsURL                string   `json:"maps_url"`
	Address                string   `json:"address"`
	City                   string   `json:"city"`
	Zipcode                string   `json:"zipcode"`
	Country                string   `json:"country"`
	Instagram              string   `json:"instagram"`
	MaintenanceMode        bool     `json:"maintenance_mode"`
	MaintenanceNotice      string   `json:"maintenance_notice"`
	AddonsEnabled          bool     `json:"addons_enabled"`
	EnableReviews          bool     `json:"enable_reviews"`
	EnableWishlist         bool     `json:"enable_wishlist"`
	EnableNewsletterSignup bool     `json:"enable_newsletter_signup"`
	AboutContent           string   `json:"about_content"`
	FooterSubText          string   `json:"footer_sub_text"`
	Slug                   string   `json:"slug"`
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
		OwnerID:                config.OwnerId,
		SiteName:               config.SiteName,
		LogoURL:                config.LogoUrl,
		FaviconURL:             config.FaviconUrl,
		MetaTags:               config.MetaTags,
		ContactEmail:           config.ContactEmail,
		ContactNumber:          config.ContactNumber,
		SenderEmail:            config.SenderEmail,
		Website:                config.Website,
		MapsURL:                config.MapsURL,
		Address:                config.Address,
		City:                   config.City,
		Zipcode:                config.Zipcode,
		Country:                config.Country,
		Instagram:              config.Instagram,
		MaintenanceMode:        config.MaintenanceMode,
		MaintenanceNotice:      config.MaintenanceNotice,
		AddonsEnabled:          config.AddonsEnabled,
		EnableReviews:          config.EnableReviews,
		EnableWishlist:         config.EnableWishlist,
		EnableNewsletterSignup: config.EnableNewsletterSignup,
		AboutContent:           config.AboutContent,
		FooterSubText:          config.FooterSubText,
		Slug:                   config.BusinessSlug,
	}
}

func (h *GetConfigByOwnerHandler) Validate(_ context.Context, _ *GetConfigByOwnerRequest) *errorModels.ServerError {
	return nil
}

func (h *GetConfigByOwnerHandler) Serve(ctx context.Context, req *GetConfigByOwnerRequest) (GetConfigByOwnerResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil {
		return GetConfigByOwnerResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		if !strings.EqualFold(authUser.Role, "console_admin") || req == nil || req.Owner == "" {
			return GetConfigByOwnerResponse{}, errorModels.PermissionDeniedServerError()
		}
		ownerBusinessID, parseErr := strconv.ParseInt(req.Owner, 10, 64)
		if parseErr != nil || ownerBusinessID <= 0 {
			return GetConfigByOwnerResponse{}, errorModels.PermissionDeniedServerError()
		}
		businessID = ownerBusinessID
	}

	businessIDStr := strconv.FormatInt(businessID, 10)

	config, err := h.ConfigService.GetConfig(ctx, businessIDStr)
	if err != nil {
		return GetConfigByOwnerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get config",
		}
	}

	theme, err := h.ConfigService.GetThemeByOwnerID(ctx, businessIDStr)
	if err != nil {
		return GetConfigByOwnerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get theme",
		}
	}

	return GetConfigByOwnerResponse{
		Config: newConfigPayload(config),
		Theme:  newThemePayload(theme),
	}, nil
}
