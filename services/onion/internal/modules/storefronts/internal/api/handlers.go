package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/storefronts/internal/application"
	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	tenant "github.com/voxire/onion/internal/platform/tenant"
)

type Payload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"owner_id"`
	BusinessID  int64  `json:"business_id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	LogoURL     string `json:"logo_url"`
	ThemeID     int64  `json:"theme_id"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toPayload(s domain.Storefront) Payload {
	return Payload{
		ID: s.ID, OwnerID: s.OwnerID, BusinessID: s.BusinessID, Name: s.Name, Domain: s.Domain,
		LogoURL: s.LogoURL, ThemeID: s.ThemeID, IsPublished: s.IsPublished,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// --- List ---

func ListRoutes() []string { return []string{"/store/sites"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Sites []Payload `json:"sites"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list sites")
	}
	out := make([]Payload, 0, len(items))
	for _, s := range items {
		out = append(out, toPayload(s))
	}
	return ListResponse{Sites: out}, nil
}

// --- Create ---

func CreateRoutes() []string { return []string{"/store/sites"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name        string `json:"name"`
		Domain      string `json:"domain"`
		LogoURL     string `json:"logo_url"`
		ThemeID     int64  `json:"theme_id"`
		IsPublished bool   `json:"is_published"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.StorefrontInput{
		Name: req.Name, Domain: req.Domain, LogoURL: req.LogoURL, ThemeID: req.ThemeID, IsPublished: req.IsPublished,
	})
	if err != nil {
		if errors.Is(err, domain.ErrThemeNotFound) {
			return CreateResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create site")
	}
	return CreateResponse{ID: id, Message: "Site created successfully"}, nil
}

// --- Update ---

func UpdateRoutes() []string { return []string{"/store/sites"} }

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Domain      string `json:"domain"`
		LogoURL     string `json:"logo_url"`
		ThemeID     int64  `json:"theme_id"`
		IsPublished bool   `json:"is_published"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.StorefrontInput{
		Name: req.Name, Domain: req.Domain, LogoURL: req.LogoURL, ThemeID: req.ThemeID, IsPublished: req.IsPublished,
	}); err != nil {
		if errors.Is(err, domain.ErrThemeNotFound) {
			return UpdateResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update site")
	}
	return UpdateResponse{Message: "Site updated successfully"}, nil
}
