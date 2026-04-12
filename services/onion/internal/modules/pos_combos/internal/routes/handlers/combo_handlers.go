package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/application"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
)

type ComboPayload struct {
	ID          int64   `json:"id"`
	OwnerID     int64   `json:"owner_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Items       string  `json:"items"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func newComboPayload(c domain.Combo) ComboPayload {
	return ComboPayload{
		ID:          c.ID,
		OwnerID:     c.OwnerID,
		Name:        c.Name,
		Description: c.Description,
		Price:       c.Price,
		Items:       c.Items,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List ---

func ListCombosRoutes() []string { return []string{"/pos/menu/combos"} }

type ListCombosHandler struct{ Service *application.Service }
type ListCombosRequest struct{}
type ListCombosResponse struct {
	Combos []ComboPayload `json:"combos"`
}

func (h *ListCombosHandler) Serve(ctx context.Context, _ *ListCombosRequest) (ListCombosResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCombosResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListCombosResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list combos"}
	}
	payloads := make([]ComboPayload, len(items))
	for i, c := range items {
		payloads[i] = newComboPayload(c)
	}
	return ListCombosResponse{Combos: payloads}, nil
}

// --- Create ---

func CreateComboRoutes() []string { return []string{"/pos/menu/combos"} }

type CreateComboHandler struct{ Service *application.Service }

type CreateComboRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Items       string  `json:"items"`
	IsActive    bool    `json:"is_active"`
}

type CreateComboResponse struct {
	ComboID int64  `json:"combo_id"`
	Message string `json:"message"`
}

func (h *CreateComboHandler) Validate(_ context.Context, req *CreateComboRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateComboHandler) Serve(ctx context.Context, req *CreateComboRequest) (CreateComboResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateComboResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, device.User.UserID, domain.ComboInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Items:       req.Items,
		IsActive:    req.IsActive,
	})
	if err != nil {
		return CreateComboResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create combo"}
	}
	return CreateComboResponse{ComboID: id, Message: "Combo created successfully"}, nil
}

// --- Update ---

func UpdateComboRoutes() []string { return []string{"/pos/menu/combos/{id}"} }

type UpdateComboHandler struct{ Service *application.Service }

type UpdateComboRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Items       string  `json:"items"`
	IsActive    bool    `json:"is_active"`
}

type UpdateComboResponse struct {
	Message string `json:"message"`
}

func (h *UpdateComboHandler) Validate(_ context.Context, req *UpdateComboRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateComboHandler) Serve(ctx context.Context, req *UpdateComboRequest) (UpdateComboResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateComboResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateComboResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid combo ID"}
	}
	if updateErr := h.Service.Update(ctx, device.User.UserID, id, domain.ComboInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Items:       req.Items,
		IsActive:    req.IsActive,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrComboNotFound) {
			return UpdateComboResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Combo not found"}
		}
		return UpdateComboResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update combo"}
	}
	return UpdateComboResponse{Message: "Combo updated successfully"}, nil
}

// --- Delete ---

func DeleteComboRoutes() []string { return []string{"/pos/menu/combos/{id}"} }

type DeleteComboHandler struct{ Service *application.Service }
type DeleteComboRequest struct{}
type DeleteComboResponse struct {
	Message string `json:"message"`
}

func (h *DeleteComboHandler) Serve(ctx context.Context, _ *DeleteComboRequest) (DeleteComboResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteComboResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteComboResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid combo ID"}
	}
	if deleteErr := h.Service.Delete(ctx, device.User.UserID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrComboNotFound) {
			return DeleteComboResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Combo not found"}
		}
		return DeleteComboResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete combo"}
	}
	return DeleteComboResponse{Message: "Combo deleted successfully"}, nil
}
