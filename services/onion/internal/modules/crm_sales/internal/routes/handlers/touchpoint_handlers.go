package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// POST /crm/sales/touchpoints
// ---------------------------------------------------------------------------

func CreateTouchpointRoutes() []string { return []string{"/crm/sales/touchpoints"} }

type CreateTouchpointHandler struct{ Service *application.Service }

type CreateTouchpointRequest struct {
	LeadID       int64   `json:"lead_id"`
	ContactID    *int64  `json:"contact_id"`
	Type         string  `json:"type"`
	Outcome      string  `json:"outcome"`
	Notes        string  `json:"notes"`
	NextFollowUp *string `json:"next_follow_up"`
}

type CreateTouchpointResponse struct {
	Touchpoint domain.Touchpoint `json:"touchpoint"`
}

func (h *CreateTouchpointHandler) Validate(_ context.Context, req *CreateTouchpointRequest) *errorModels.ServerError {
	if req.LeadID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "lead_id is required"}
	}
	return nil
}

func (h *CreateTouchpointHandler) Serve(ctx context.Context, req *CreateTouchpointRequest) (CreateTouchpointResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTouchpointResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTouchpointResponse{}, errorModels.PermissionDeniedServerError()
	}

	tp, err := h.Service.CreateTouchpoint(ctx, businessID, device.User.UserID, domain.CreateTouchpointRequest{
		LeadID:       req.LeadID,
		ContactID:    req.ContactID,
		Type:         req.Type,
		Outcome:      req.Outcome,
		Notes:        req.Notes,
		NextFollowUp: req.NextFollowUp,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create touchpoint", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateTouchpointResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create touchpoint"}
	}

	return CreateTouchpointResponse{Touchpoint: tp}, nil
}

// ---------------------------------------------------------------------------
// DELETE /crm/sales/touchpoints/{id}
// ---------------------------------------------------------------------------

func DeleteTouchpointRoutes() []string { return []string{"/crm/sales/touchpoints/{id}"} }

type (
	DeleteTouchpointHandler  struct{ Service *application.Service }
	DeleteTouchpointRequest  struct{}
	DeleteTouchpointResponse struct {
		Success bool `json:"success"`
	}
)

func (h *DeleteTouchpointHandler) Serve(ctx context.Context, _ *DeleteTouchpointRequest) (DeleteTouchpointResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTouchpointResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteTouchpointResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid touchpoint id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteTouchpointResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteTouchpoint(ctx, id, businessID); err != nil {
		if errors.Is(err, domain.ErrTouchpointNotFound) {
			return DeleteTouchpointResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Touchpoint not found"}
		}
		logger.Ctx(ctx).Error("failed to delete touchpoint", logger.Any("owner_id", device.User.UserID), logger.Any("touchpoint_id", id), logger.Error(err))
		return DeleteTouchpointResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete touchpoint"}
	}

	return DeleteTouchpointResponse{Success: true}, nil
}
