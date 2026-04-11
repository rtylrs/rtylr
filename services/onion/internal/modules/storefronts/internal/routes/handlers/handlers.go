package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/storefronts/internal/application"
	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type Payload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"ownerId"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	LogoURL     string `json:"logoUrl"`
	ThemeID     int64  `json:"themeId"`
	IsPublished bool   `json:"isPublished"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func toPayload(s domain.Storefront) Payload {
	return Payload{
		ID: s.ID, OwnerID: s.OwnerID, Name: s.Name, Domain: s.Domain,
		LogoURL: s.LogoURL, ThemeID: s.ThemeID, IsPublished: s.IsPublished,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// --- List ---

func ListRoutes() []string { return []string{"/store/sites"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Sites []Payload `json:"sites"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	items, err := h.Service.List(ctx, authUser.UserID)
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

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	LogoURL     string `json:"logoUrl"`
	ThemeID     int64  `json:"themeId"`
	IsPublished bool   `json:"isPublished"`
}
type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	id, err := h.Service.Create(ctx, authUser.UserID, domain.StorefrontInput{
		Name: req.Name, Domain: req.Domain, LogoURL: req.LogoURL, ThemeID: req.ThemeID, IsPublished: req.IsPublished,
	})
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create site"}
	}
	return CreateResponse{ID: id, Message: "Site created successfully"}, nil
}

// --- Update ---

func UpdateRoutes() []string { return []string{"/store/sites"} }

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	LogoURL     string `json:"logoUrl"`
	ThemeID     int64  `json:"themeId"`
	IsPublished bool   `json:"isPublished"`
}
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.StorefrontInput{
		Name: req.Name, Domain: req.Domain, LogoURL: req.LogoURL, ThemeID: req.ThemeID, IsPublished: req.IsPublished,
	}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update site"}
	}
	return UpdateResponse{Message: "Site updated successfully"}, nil
}
