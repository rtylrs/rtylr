package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/cycles/internal/application"
	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type CyclePayload struct {
	ID          int64   `json:"id"`
	OwnerID     int64   `json:"owner_id"`
	WorkspaceID int64   `json:"workspace_id"`
	Name        string  `json:"name"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func newCyclePayload(c domain.Cycle) CyclePayload {
	p := CyclePayload{
		ID:          c.ID,
		OwnerID:     c.OwnerID,
		WorkspaceID: c.WorkspaceID,
		Name:        c.Name,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.StartDate != nil {
		s := c.StartDate.Format("2006-01-02")
		p.StartDate = &s
	}
	if c.EndDate != nil {
		s := c.EndDate.Format("2006-01-02")
		p.EndDate = &s
	}
	return p
}

// --- List Cycles ---

func CycleListRoutes() []string { return []string{"/flow/cycles"} }

type (
	ListCyclesHandler  struct{ Service *application.Service }
	ListCyclesRequest  struct{}
	ListCyclesResponse struct {
		Cycles []CyclePayload `json:"cycles"`
	}
)

func (h *ListCyclesHandler) Validate(_ context.Context, _ *ListCyclesRequest) *errorModels.ServerError {
	return nil
}

func (h *ListCyclesHandler) Serve(ctx context.Context, _ *ListCyclesRequest) (ListCyclesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCyclesResponse{}, errorModels.PermissionDeniedServerError()
	}

	var workspaceID *int64
	if wsStr := chi.URLParamFromCtx(ctx, "workspaceId"); wsStr != "" {
		if ws, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			workspaceID = &ws
		}
	}

	cycles, err := h.Service.List(ctx, device.User.UserID, workspaceID)
	if err != nil {
		return ListCyclesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list cycles"}
	}
	payloads := make([]CyclePayload, len(cycles))
	for i, c := range cycles {
		payloads[i] = newCyclePayload(c)
	}
	return ListCyclesResponse{Cycles: payloads}, nil
}

// --- Create Cycle ---

func CycleCreateRoutes() []string { return []string{"/flow/cycles"} }

type CreateCycleHandler struct{ Service *application.Service }

type CreateCycleRequest struct {
	WorkspaceID int64  `json:"workspace_id"`
	Name        string `json:"name"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type CreateCycleResponse struct {
	CycleID int64  `json:"cycle_id"`
	Message string `json:"message"`
}

func (h *CreateCycleHandler) Validate(_ context.Context, req *CreateCycleRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}
	return nil
}

func (h *CreateCycleHandler) Serve(ctx context.Context, req *CreateCycleRequest) (CreateCycleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCycleResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.CycleInput{
		OwnerID:     device.User.UserID,
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			input.StartDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			input.EndDate = &t
		}
	}
	id, err := h.Service.Create(ctx, input)
	if err != nil {
		return CreateCycleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create cycle"}
	}
	return CreateCycleResponse{CycleID: id, Message: "Cycle created successfully"}, nil
}

// --- Update Cycle ---

func CycleUpdateRoutes() []string { return []string{"/flow/cycles/{id}"} }

type UpdateCycleHandler struct{ Service *application.Service }

type UpdateCycleRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
}

type UpdateCycleResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCycleHandler) Validate(_ context.Context, req *UpdateCycleRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateCycleHandler) Serve(ctx context.Context, req *UpdateCycleRequest) (UpdateCycleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCycleResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateCycleResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid cycle ID"}
	}
	input := domain.CycleUpdateInput{
		ID:      id,
		OwnerID: device.User.UserID,
		Name:    req.Name,
		Status:  req.Status,
	}
	if req.StartDate != "" {
		if t, parseErr := time.Parse("2006-01-02", req.StartDate); parseErr == nil {
			input.StartDate = &t
		}
	}
	if req.EndDate != "" {
		if t, parseErr := time.Parse("2006-01-02", req.EndDate); parseErr == nil {
			input.EndDate = &t
		}
	}
	if updateErr := h.Service.Update(ctx, input); updateErr != nil {
		if errors.Is(updateErr, domain.ErrCycleNotFound) {
			return UpdateCycleResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Cycle not found"}
		}
		return UpdateCycleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update cycle"}
	}
	return UpdateCycleResponse{Message: "Cycle updated successfully"}, nil
}
