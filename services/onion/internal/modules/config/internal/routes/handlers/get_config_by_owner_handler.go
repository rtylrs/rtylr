package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	PrimaryColor          string `json:"primaryColor"`
	SecondaryColor        string `json:"secondaryColor"`
	AccentColor           string `json:"accentColor"`
	DarkGreenColor        string `json:"darkGreenColor"`
	BackgroundColor       string `json:"backgroundColor"`
	BgColor               string `json:"bgColor"`
	SurfaceColor          string `json:"surfaceColor"`
	LightColor            string `json:"lightColor"`
	DarkColor             string `json:"darkColor"`
	ItemBackgroundColor   string `json:"itemBackgroundColor"`
	ErrorColor            string `json:"errorColor"`
	SuccessColor          string `json:"successColor"`
	WarningColor          string `json:"warningColor"`
	InfoColor             string `json:"infoColor"`
	TextPrimaryColor      string `json:"textPrimaryColor"`
	TextSecondaryColor    string `json:"textSecondaryColor"`
	SubtextColor          string `json:"subtextColor"`
	LightGrayColor        string `json:"lightGrayColor"`
	ActiveColor           string `json:"activeColor"`
	AdminBackgroundColor  string `json:"adminBackgroundColor"`
	AdminPrimaryColor     string `json:"adminPrimaryColor"`
	VendorBackgroundColor string `json:"vendorBackgroundColor"`
	BorderColor           string `json:"borderColor"`
	FontFamilyPrimary     string `json:"fontFamilyPrimary"`
	FontFamilySecondary   string `json:"fontFamilySecondary"`
	FontSizeBase          string `json:"fontSizeBase"`
	FontSizeHeading       string `json:"fontSizeHeading"`
	MobileBreakpoint      string `json:"mobileBreakpoint"`
}

type ConfigPayload struct {
	OwnerID                string   `json:"ownerId"`
	SiteName               string   `json:"siteName"`
	LogoURL                string   `json:"logoUrl"`
	FaviconURL             string   `json:"faviconUrl"`
	MetaTags               []string `json:"metaTags"`
	ContactEmail           string   `json:"contactEmail"`
	ContactNumber          string   `json:"contactNumber"`
	Instagram              string   `json:"instagram"`
	MaintenanceMode        bool     `json:"maintenanceMode"`
	MaintenanceNotice      string   `json:"maintenanceNotice"`
	AddonsEnabled          bool     `json:"addonsEnabled"`
	EnableReviews          bool     `json:"enableReviews"`
	EnableWishlist         bool     `json:"enableWishlist"`
	EnableNewsletterSignup bool     `json:"enableNewsletterSignup"`
	AboutContent           string   `json:"aboutContent"`
	FooterSubText          string   `json:"footerSubText"`
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
		Instagram:              config.Instagram,
		MaintenanceMode:        config.MaintenanceMode,
		MaintenanceNotice:      config.MaintenanceNotice,
		AddonsEnabled:          config.AddonsEnabled,
		EnableReviews:          config.EnableReviews,
		EnableWishlist:         config.EnableWishlist,
		EnableNewsletterSignup: config.EnableNewsletterSignup,
		AboutContent:           config.AboutContent,
		FooterSubText:          config.FooterSubText,
	}
}

func (h *GetConfigByOwnerHandler) Validate(_ context.Context, req *GetConfigByOwnerRequest) *errorModels.ServerError {
	if req.Owner == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Owner is required",
		}
	}

	return nil
}

func (h *GetConfigByOwnerHandler) Serve(ctx context.Context, req *GetConfigByOwnerRequest) (GetConfigByOwnerResponse, *errorModels.ServerError) {
	config, err := h.ConfigService.GetConfig(ctx, req.Owner)
	if err != nil {
		return GetConfigByOwnerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get config",
		}
	}

	theme, err := h.ConfigService.GetThemeByOwnerID(ctx, req.Owner)
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
