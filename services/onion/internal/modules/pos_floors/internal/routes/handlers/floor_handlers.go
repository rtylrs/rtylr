package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/application"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
)

type FloorPayload struct {
	ID         int64  `json:"id"`
	BusinessID int64  `json:"business_id"`
	BranchID   int64  `json:"branch_id"`
	Name       string `json:"name"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
}

func newFloorPayload(f domain.Floor) FloorPayload {
	return FloorPayload{
		ID:         f.ID,
		BusinessID: f.BusinessID,
		BranchID:   f.BranchID,
		Name:       f.Name,
		SortOrder:  f.SortOrder,
		IsActive:   f.IsActive,
		CreatedAt:  f.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func FloorListRoutes() []string { return []string{"/pos/floors"} }

type ListFloorsHandler struct{ Service *application.Service }
type ListFloorsRequest struct {
	BranchID int64 `json:"branch_id"`
}
type ListFloorsResponse struct {
	Floors []FloorPayload `json:"floors"`
}

func (h *ListFloorsHandler) Serve(ctx context.Context, req *ListFloorsRequest) (ListFloorsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListFloorsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.BranchID == 0 {
		return ListFloorsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branch_id is required"}
	}
	floors, err := h.Service.List(ctx, device.User.UserID, req.BranchID)
	if err != nil {
		return ListFloorsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list floors"}
	}
	payloads := make([]FloorPayload, len(floors))
	for i, f := range floors {
		payloads[i] = newFloorPayload(f)
	}
	return ListFloorsResponse{Floors: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func FloorCreateRoutes() []string { return []string{"/pos/floors"} }

type CreateFloorHandler struct{ Service *application.Service }

type CreateFloorRequest struct {
	BranchID  int64  `json:"branch_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type CreateFloorResponse struct {
	FloorID int64  `json:"floor_id"`
	Message string `json:"message"`
}

func (h *CreateFloorHandler) Validate(_ context.Context, req *CreateFloorRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.BranchID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Branch ID is required"}
	}
	return nil
}

func (h *CreateFloorHandler) Serve(ctx context.Context, req *CreateFloorRequest) (CreateFloorResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateFloorResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.FloorInput{
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		Name:       req.Name,
		SortOrder:  req.SortOrder,
	})
	if err != nil {
		return CreateFloorResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create floor"}
	}
	return CreateFloorResponse{FloorID: id, Message: "Floor created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func FloorUpdateRoutes() []string { return []string{"/pos/floors/{id}"} }

type UpdateFloorHandler struct{ Service *application.Service }

type UpdateFloorRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type UpdateFloorResponse struct {
	Message string `json:"message"`
}

func (h *UpdateFloorHandler) Validate(_ context.Context, req *UpdateFloorRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateFloorHandler) Serve(ctx context.Context, req *UpdateFloorRequest) (UpdateFloorResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateFloorResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateFloorResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid floor ID"}
	}
	if updateErr := h.Service.Update(ctx, domain.FloorUpdateInput{
		ID:         id,
		BusinessID: device.User.UserID,
		Name:       req.Name,
		SortOrder:  req.SortOrder,
	}); updateErr != nil {
		return UpdateFloorResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update floor"}
	}
	return UpdateFloorResponse{Message: "Floor updated successfully"}, nil
}

// ─── Delete (soft) ───────────────────────────────────────────────────────────

func FloorDeleteRoutes() []string { return []string{"/pos/floors/{id}"} }

type DeleteFloorHandler struct{ Service *application.Service }
type DeleteFloorRequest struct{}
type DeleteFloorResponse struct {
	Message string `json:"message"`
}

func (h *DeleteFloorHandler) Serve(ctx context.Context, _ *DeleteFloorRequest) (DeleteFloorResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteFloorResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteFloorResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid floor ID"}
	}
	if deleteErr := h.Service.Delete(ctx, device.User.UserID, id); deleteErr != nil {
		return DeleteFloorResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete floor"}
	}
	return DeleteFloorResponse{Message: "Floor deleted successfully"}, nil
}
