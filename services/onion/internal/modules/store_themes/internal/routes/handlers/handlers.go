package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/store_themes/internal/application"
	"github.com/voxire/onion/internal/modules/store_themes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type Payload struct {
	ID           int64  `json:"id"`
	OwnerID      int64  `json:"owner_id"`
	BusinessID   int64  `json:"business_id"`
	Name         string `json:"name"`
	PrimaryColor string `json:"primary_color"`
	Font         string `json:"font"`
	HeaderStyle  string `json:"header_style"`
	FooterStyle  string `json:"footer_style"`
	CustomCSS    string `json:"custom_css"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func toPayload(t domain.Theme) Payload {
	return Payload{
		ID: t.ID, OwnerID: t.OwnerID, BusinessID: t.BusinessID, Name: t.Name, PrimaryColor: t.PrimaryColor, Font: t.Font,
		HeaderStyle: t.HeaderStyle, FooterStyle: t.FooterStyle, CustomCSS: t.CustomCSS,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ListRoutes() []string   { return []string{"/store/themes"} }
func CreateRoutes() []string { return []string{"/store/themes"} }
func UpdateRoutes() []string { return []string{"/store/themes"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Themes []Payload `json:"themes"`
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
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list themes"}
	}
	out := make([]Payload, 0, len(items))
	for _, i := range items {
		out = append(out, toPayload(i))
	}
	return ListResponse{Themes: out}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name         string `json:"name"`
		PrimaryColor string `json:"primary_color"`
		Font         string `json:"font"`
		HeaderStyle  string `json:"header_style"`
		FooterStyle  string `json:"footer_style"`
		CustomCSS    string `json:"custom_css"`
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
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.ThemeInput{Name: req.Name, PrimaryColor: req.PrimaryColor, Font: req.Font, HeaderStyle: req.HeaderStyle, FooterStyle: req.FooterStyle, CustomCSS: req.CustomCSS})
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create theme"}
	}
	return CreateResponse{ID: id, Message: "Theme created"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		PrimaryColor string `json:"primary_color"`
		Font         string `json:"font"`
		HeaderStyle  string `json:"header_style"`
		FooterStyle  string `json:"footer_style"`
		CustomCSS    string `json:"custom_css"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
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
	if err := h.Service.Update(ctx, businessID, req.ID, domain.ThemeInput{Name: req.Name, PrimaryColor: req.PrimaryColor, Font: req.Font, HeaderStyle: req.HeaderStyle, FooterStyle: req.FooterStyle, CustomCSS: req.CustomCSS}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update theme"}
	}
	return UpdateResponse{Message: "Theme updated"}, nil
}
