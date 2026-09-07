package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func GetRecruitCustomizationRoutes() []string { return []string{"/recruit/customization"} }
func PutRecruitCustomizationRoutes() []string { return []string{"/recruit/customization"} }

// ─────────────────────────────────────────────
// Shared payload
// ─────────────────────────────────────────────

type CustomizationPayload struct {
	ID                 int64  `json:"id"`
	OwnerID            int64  `json:"owner_id"`
	Slug               string `json:"slug"`
	BrandName          string `json:"brand_name"`
	Tagline            string `json:"tagline"`
	AboutBlurb         string `json:"about_blurb"`
	LogoURL            string `json:"logo_url"`
	PrimaryColor       string `json:"primary_color"`
	AccentColor        string `json:"accent_color"`
	BgColor            string `json:"bg_color"`
	TextColor          string `json:"text_color"`
	MutedTextColor     string `json:"muted_text_color"`
	FontPairing        string `json:"font_pairing"`
	HeroImageURL       string `json:"hero_image_url"`
	FaviconURL         string `json:"favicon_url"`
	Template           string `json:"template"`
	HeaderHTML         string `json:"header_html"`
	MetaTitle          string `json:"meta_title"`
	MetaDescription    string `json:"meta_description"`
	PrivacyURL         string `json:"privacy_url"`
	TermsURL           string `json:"terms_url"`
	GAMeasurementID    string `json:"ga_measurement_id"`
	MetaPixelID        string `json:"meta_pixel_id"`
	RequireConsent     bool   `json:"require_consent"`
	NotificationEmail  string `json:"notification_email"`
	EmailNotifications bool   `json:"email_notifications"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

func toCustomizationPayload(c domain.Customization) *CustomizationPayload {
	return &CustomizationPayload{
		ID:                 c.ID,
		OwnerID:            c.OwnerID,
		Slug:               c.Slug,
		BrandName:          c.BrandName,
		Tagline:            c.Tagline,
		AboutBlurb:         c.AboutBlurb,
		LogoURL:            c.LogoURL,
		PrimaryColor:       c.PrimaryColor,
		AccentColor:        c.AccentColor,
		BgColor:            c.BgColor,
		TextColor:          c.TextColor,
		MutedTextColor:     c.MutedTextColor,
		FontPairing:        c.FontPairing,
		HeroImageURL:       c.HeroImageURL,
		FaviconURL:         c.FaviconURL,
		Template:           c.Template,
		HeaderHTML:         c.HeaderHTML,
		MetaTitle:          c.MetaTitle,
		MetaDescription:    c.MetaDescription,
		PrivacyURL:         c.PrivacyURL,
		TermsURL:           c.TermsURL,
		GAMeasurementID:    c.GAMeasurementID,
		MetaPixelID:        c.MetaPixelID,
		RequireConsent:     c.RequireConsent,
		NotificationEmail:  c.NotificationEmail,
		EmailNotifications: c.EmailNotifications,
		CreatedAt:          c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          c.UpdatedAt.Format(time.RFC3339),
	}
}

// ─────────────────────────────────────────────
// GET /recruit/customization — my branding
// ─────────────────────────────────────────────

type GetCustomizationHandler struct {
	CareersService *application.Service
}

type GetCustomizationRequest struct{}

type GetCustomizationResponse struct {
	Customization *CustomizationPayload `json:"customization"`
}

func (h *GetCustomizationHandler) Serve(ctx context.Context, _ *GetCustomizationRequest) (GetCustomizationResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetCustomizationResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return GetCustomizationResponse{}, onekiterrors.PermissionDenied()
	}

	c, err := h.CareersService.GetMyCustomization(ctx, businessID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			// First-time callers get an empty stub rather than a 404 —
			// lets the frontend render a "Set up your careers page"
			// empty state without branching on error shape.
			return GetCustomizationResponse{Customization: nil}, nil
		}
		return GetCustomizationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load customization")
	}
	return GetCustomizationResponse{Customization: toCustomizationPayload(c)}, nil
}

// ─────────────────────────────────────────────
// PUT /recruit/customization — upsert
// ─────────────────────────────────────────────

type PutCustomizationHandler struct {
	CareersService *application.Service
}

type PutCustomizationRequest struct {
	Slug               string `json:"slug"`
	BrandName          string `json:"brand_name"`
	Tagline            string `json:"tagline"`
	AboutBlurb         string `json:"about_blurb"`
	LogoURL            string `json:"logo_url"`
	PrimaryColor       string `json:"primary_color"`
	AccentColor        string `json:"accent_color"`
	BgColor            string `json:"bg_color"`
	TextColor          string `json:"text_color"`
	MutedTextColor     string `json:"muted_text_color"`
	FontPairing        string `json:"font_pairing"`
	HeroImageURL       string `json:"hero_image_url"`
	FaviconURL         string `json:"favicon_url"`
	Template           string `json:"template"`
	HeaderHTML         string `json:"header_html"`
	MetaTitle          string `json:"meta_title"`
	MetaDescription    string `json:"meta_description"`
	PrivacyURL         string `json:"privacy_url"`
	TermsURL           string `json:"terms_url"`
	GAMeasurementID    string `json:"ga_measurement_id"`
	MetaPixelID        string `json:"meta_pixel_id"`
	RequireConsent     bool   `json:"require_consent"`
	NotificationEmail  string `json:"notification_email"`
	EmailNotifications bool   `json:"email_notifications"`
}

type PutCustomizationResponse struct {
	Message string `json:"message"`
}

func (h *PutCustomizationHandler) Validate(_ context.Context, req *PutCustomizationRequest) error {
	if req.Slug == "" {
		return onekiterrors.New(http.StatusBadRequest, "Slug is required")
	}
	if req.BrandName == "" {
		return onekiterrors.New(http.StatusBadRequest, "Brand name is required")
	}
	return nil
}

func (h *PutCustomizationHandler) Serve(ctx context.Context, req *PutCustomizationRequest) (PutCustomizationResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return PutCustomizationResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return PutCustomizationResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.CustomizationInput{
		OwnerID:            authUser.UserID,
		BusinessID:         businessID,
		Slug:               req.Slug,
		BrandName:          req.BrandName,
		Tagline:            req.Tagline,
		AboutBlurb:         req.AboutBlurb,
		LogoURL:            req.LogoURL,
		PrimaryColor:       req.PrimaryColor,
		AccentColor:        req.AccentColor,
		BgColor:            req.BgColor,
		TextColor:          req.TextColor,
		MutedTextColor:     req.MutedTextColor,
		FontPairing:        req.FontPairing,
		HeroImageURL:       req.HeroImageURL,
		FaviconURL:         req.FaviconURL,
		Template:           req.Template,
		HeaderHTML:         req.HeaderHTML,
		MetaTitle:          req.MetaTitle,
		MetaDescription:    req.MetaDescription,
		PrivacyURL:         req.PrivacyURL,
		TermsURL:           req.TermsURL,
		GAMeasurementID:    req.GAMeasurementID,
		MetaPixelID:        req.MetaPixelID,
		RequireConsent:     req.RequireConsent,
		NotificationEmail:  req.NotificationEmail,
		EmailNotifications: req.EmailNotifications,
	}

	if err := h.CareersService.UpsertCustomization(ctx, input); err != nil {
		return PutCustomizationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to save customization")
	}
	invalidatePublicCareersCache(ctx, businessID)
	return PutCustomizationResponse{Message: "Customization saved"}, nil
}
