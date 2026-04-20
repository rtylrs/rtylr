package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	ID           int64  `json:"id"`
	OwnerID      int64  `json:"owner_id"`
	Slug         string `json:"slug"`
	BrandName    string `json:"brand_name"`
	Tagline      string `json:"tagline"`
	AboutBlurb   string `json:"about_blurb"`
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	Template     string `json:"template"`
	HeaderHTML   string `json:"header_html"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func toCustomizationPayload(c domain.Customization) *CustomizationPayload {
	return &CustomizationPayload{
		ID:           c.ID,
		OwnerID:      c.OwnerID,
		Slug:         c.Slug,
		BrandName:    c.BrandName,
		Tagline:      c.Tagline,
		AboutBlurb:   c.AboutBlurb,
		LogoURL:      c.LogoURL,
		PrimaryColor: c.PrimaryColor,
		Template:     c.Template,
		HeaderHTML:   c.HeaderHTML,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
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

func (h *GetCustomizationHandler) Serve(ctx context.Context, _ *GetCustomizationRequest) (GetCustomizationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetCustomizationResponse{}, errorModels.PermissionDeniedServerError()
	}

	c, err := h.CareersService.GetMyCustomization(ctx, authUser.UserID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			// First-time callers get an empty stub rather than a 404 —
			// lets the frontend render a "Set up your careers page"
			// empty state without branching on error shape.
			return GetCustomizationResponse{Customization: nil}, nil
		}
		return GetCustomizationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to load customization",
		}
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
	Slug         string `json:"slug"`
	BrandName    string `json:"brand_name"`
	Tagline      string `json:"tagline"`
	AboutBlurb   string `json:"about_blurb"`
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	Template     string `json:"template"`
	HeaderHTML   string `json:"header_html"`
}

type PutCustomizationResponse struct {
	Message string `json:"message"`
}

func (h *PutCustomizationHandler) Validate(_ context.Context, req *PutCustomizationRequest) *errorModels.ServerError {
	if req.Slug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Slug is required"}
	}
	if req.BrandName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Brand name is required"}
	}
	return nil
}

func (h *PutCustomizationHandler) Serve(ctx context.Context, req *PutCustomizationRequest) (PutCustomizationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return PutCustomizationResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.CustomizationInput{
		OwnerID:      authUser.UserID,
		Slug:         req.Slug,
		BrandName:    req.BrandName,
		Tagline:      req.Tagline,
		AboutBlurb:   req.AboutBlurb,
		LogoURL:      req.LogoURL,
		PrimaryColor: req.PrimaryColor,
		Template:     req.Template,
		HeaderHTML:   req.HeaderHTML,
	}

	if err := h.CareersService.UpsertCustomization(ctx, input); err != nil {
		return PutCustomizationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to save customization",
		}
	}
	return PutCustomizationResponse{Message: "Customization saved"}, nil
}
