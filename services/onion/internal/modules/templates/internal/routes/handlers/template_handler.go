package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/templates/internal/application"
	"github.com/voxire/onion/internal/modules/templates/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func TemplateListRoutes() []string {
	return []string{"/crm/templates"}
}

func TemplateCreateRoutes() []string {
	return []string{"/crm/templates"}
}

func TemplateUpdateRoutes() []string {
	return []string{"/crm/templates"}
}

func TemplateDeleteRoutes() []string {
	return []string{"/crm/templates"}
}

type TemplatePayload struct {
	ID        int64           `json:"id"`
	OwnerID   int64           `json:"owner_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Subject   string          `json:"subject"`
	Body      string          `json:"body"`
	Variables json.RawMessage `json:"variables"`
	IsActive  bool            `json:"is_active"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func newTemplatePayload(t domain.Template) TemplatePayload {
	return TemplatePayload{
		ID:        t.ID,
		OwnerID:   t.OwnerID,
		Name:      t.Name,
		Type:      t.Type,
		Subject:   t.Subject,
		Body:      t.Body,
		Variables: json.RawMessage(t.Variables),
		IsActive:  t.IsActive,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// -- LIST --------------------------------------------------------------------

type ListTemplatesHandler struct {
	Service *application.Service
}

type ListTemplatesRequest struct{}

type ListTemplatesResponse struct {
	Templates []TemplatePayload `json:"templates"`
}

func (h *ListTemplatesHandler) Serve(ctx context.Context, _ *ListTemplatesRequest) (ListTemplatesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTemplatesResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTemplatesResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListTemplatesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list templates",
		}
	}

	payloads := make([]TemplatePayload, len(items))
	for i, t := range items {
		payloads[i] = newTemplatePayload(t)
	}

	return ListTemplatesResponse{Templates: payloads}, nil
}

// -- CREATE ------------------------------------------------------------------

type CreateTemplateHandler struct {
	Service *application.Service
}

type CreateTemplateRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Subject   string          `json:"subject"`
	Body      string          `json:"body"`
	Variables json.RawMessage `json:"variables"`
	IsActive  bool            `json:"is_active"`
}

type CreateTemplateResponse struct {
	TemplateID int64  `json:"template_id"`
	Message    string `json:"message"`
}

func (h *CreateTemplateHandler) Validate(_ context.Context, req *CreateTemplateRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Template name is required",
		}
	}

	if strings.TrimSpace(req.Type) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Template type is required",
		}
	}

	if strings.TrimSpace(req.Body) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Template body is required",
		}
	}

	return nil
}

func (h *CreateTemplateHandler) Serve(ctx context.Context, req *CreateTemplateRequest) (CreateTemplateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	variables := []byte(req.Variables)
	if variables == nil {
		variables = []byte("[]")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.TemplateInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:      req.Name,
		Type:      req.Type,
		Subject:   req.Subject,
		Body:      req.Body,
		Variables: variables,
		IsActive:  req.IsActive,
	})
	if err != nil {
		return CreateTemplateResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create template",
		}
	}

	return CreateTemplateResponse{TemplateID: id, Message: "Template created successfully"}, nil
}

// -- UPDATE ------------------------------------------------------------------

type UpdateTemplateHandler struct {
	Service *application.Service
}

type UpdateTemplateRequest struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Subject   string          `json:"subject"`
	Body      string          `json:"body"`
	Variables json.RawMessage `json:"variables"`
	IsActive  bool            `json:"is_active"`
}

type UpdateTemplateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTemplateHandler) Validate(_ context.Context, req *UpdateTemplateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Template ID is required",
		}
	}

	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Template name is required",
		}
	}

	return nil
}

func (h *UpdateTemplateHandler) Serve(ctx context.Context, req *UpdateTemplateRequest) (UpdateTemplateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	variables := []byte(req.Variables)
	if variables == nil {
		variables = []byte("[]")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Update(ctx, domain.TemplateUpdateInput{
		ID:         req.ID,
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:      req.Name,
		Type:      req.Type,
		Subject:   req.Subject,
		Body:      req.Body,
		Variables: variables,
		IsActive:  req.IsActive,
	}); err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return UpdateTemplateResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Template not found",
			}
		}

		return UpdateTemplateResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update template",
		}
	}

	return UpdateTemplateResponse{Message: "Template updated successfully"}, nil
}

// -- DELETE ------------------------------------------------------------------

type DeleteTemplateHandler struct {
	Service *application.Service
}

type DeleteTemplateRequest struct {
	ID int64 `json:"id"`
}

type DeleteTemplateResponse struct {
	Message string `json:"message"`
}

func (h *DeleteTemplateHandler) Validate(_ context.Context, req *DeleteTemplateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Template ID is required",
		}
	}

	return nil
}

func (h *DeleteTemplateHandler) Serve(ctx context.Context, req *DeleteTemplateRequest) (DeleteTemplateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return DeleteTemplateResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Template not found",
			}
		}

		return DeleteTemplateResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete template",
		}
	}

	return DeleteTemplateResponse{Message: "Template deleted successfully"}, nil
}
