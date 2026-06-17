package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/application"
	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type ModifierPayload struct {
	ID         int64   `json:"id"`
	OwnerID    int64   `json:"owner_id"`
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
	GroupName  string  `json:"group_name"`
	IsRequired bool    `json:"is_required"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func newModifierPayload(m domain.Modifier) ModifierPayload {
	return ModifierPayload{
		ID:         m.ID,
		OwnerID:    m.OwnerID,
		Name:       m.Name,
		PriceDelta: m.PriceDelta,
		GroupName:  m.GroupName,
		IsRequired: m.IsRequired,
		CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List ---

func ListModifiersRoutes() []string { return []string{"/pos/menu/modifiers", "/pos/modifiers"} }

type (
	ListModifiersHandler  struct{ Service *application.Service }
	ListModifiersRequest  struct{}
	ListModifiersResponse struct {
		Modifiers []ModifierPayload `json:"modifiers"`
	}
)

func (h *ListModifiersHandler) Serve(ctx context.Context, _ *ListModifiersRequest) (ListModifiersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListModifiersResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)
	items, err := h.Service.List(ctx, ownerID)
	if err != nil {
		return ListModifiersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list modifiers"}
	}
	payloads := make([]ModifierPayload, len(items))
	for i, m := range items {
		payloads[i] = newModifierPayload(m)
	}
	return ListModifiersResponse{Modifiers: payloads}, nil
}

// --- Create ---

func CreateModifierRoutes() []string { return []string{"/pos/menu/modifiers", "/pos/modifiers"} }

type CreateModifierHandler struct{ Service *application.Service }

type CreateModifierRequest struct {
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
	GroupName  string  `json:"group_name"`
	IsRequired bool    `json:"is_required"`
}

type CreateModifierResponse struct {
	ModifierID int64  `json:"modifier_id"`
	Message    string `json:"message"`
}

func (h *CreateModifierHandler) Validate(_ context.Context, req *CreateModifierRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateModifierHandler) Serve(ctx context.Context, req *CreateModifierRequest) (CreateModifierResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateModifierResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)
	id, err := h.Service.Create(ctx, ownerID, domain.ModifierInput{
		Name:       req.Name,
		PriceDelta: req.PriceDelta,
		GroupName:  req.GroupName,
		IsRequired: req.IsRequired,
	})
	if err != nil {
		return CreateModifierResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create modifier"}
	}
	return CreateModifierResponse{ModifierID: id, Message: "Modifier created successfully"}, nil
}

// --- Update ---

func UpdateModifierRoutes() []string { return []string{"/pos/menu/modifiers/{id}"} }

type UpdateModifierHandler struct{ Service *application.Service }

type UpdateModifierRequest struct {
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
	GroupName  string  `json:"group_name"`
	IsRequired bool    `json:"is_required"`
}

type UpdateModifierResponse struct {
	Message string `json:"message"`
}

func (h *UpdateModifierHandler) Validate(_ context.Context, req *UpdateModifierRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateModifierHandler) Serve(ctx context.Context, req *UpdateModifierRequest) (UpdateModifierResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateModifierResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateModifierResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid modifier ID"}
	}
	if updateErr := h.Service.Update(ctx, ownerID, id, domain.ModifierInput{
		Name:       req.Name,
		PriceDelta: req.PriceDelta,
		GroupName:  req.GroupName,
		IsRequired: req.IsRequired,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrModifierNotFound) {
			return UpdateModifierResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Modifier not found"}
		}
		return UpdateModifierResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update modifier"}
	}
	return UpdateModifierResponse{Message: "Modifier updated successfully"}, nil
}

// --- Delete ---

func DeleteModifierRoutes() []string { return []string{"/pos/menu/modifiers/{id}"} }

type (
	DeleteModifierHandler  struct{ Service *application.Service }
	DeleteModifierRequest  struct{}
	DeleteModifierResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteModifierHandler) Serve(ctx context.Context, _ *DeleteModifierRequest) (DeleteModifierResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteModifierResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteModifierResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid modifier ID"}
	}
	if deleteErr := h.Service.Delete(ctx, ownerID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrModifierNotFound) {
			return DeleteModifierResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Modifier not found"}
		}
		return DeleteModifierResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete modifier"}
	}
	return DeleteModifierResponse{Message: "Modifier deleted successfully"}, nil
}
