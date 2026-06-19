package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/storefronts/internal/application"
	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list sites"}
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

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.StorefrontInput{
		Name: req.Name, Domain: req.Domain, LogoURL: req.LogoURL, ThemeID: req.ThemeID, IsPublished: req.IsPublished,
	})
	if err != nil {
		return CreateResponse{}, errorModels.FromError(err, "Failed to create site")
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

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.StorefrontInput{
		Name: req.Name, Domain: req.Domain, LogoURL: req.LogoURL, ThemeID: req.ThemeID, IsPublished: req.IsPublished,
	}); err != nil {
		return UpdateResponse{}, errorModels.FromError(err, "Failed to update site")
	}
	return UpdateResponse{Message: "Site updated successfully"}, nil
}
