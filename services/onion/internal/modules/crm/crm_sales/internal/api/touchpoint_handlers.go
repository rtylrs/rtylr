package api

import (
	"context"
	"errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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
	Touchpoint touchpointPayload `json:"touchpoint"`
}

func (h *CreateTouchpointHandler) Validate(_ context.Context, req *CreateTouchpointRequest) error {
	if req.LeadID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "lead_id is required")
	}
	return nil
}

func (h *CreateTouchpointHandler) Serve(ctx context.Context, req *CreateTouchpointRequest) (CreateTouchpointResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTouchpointResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
		return CreateTouchpointResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTouchpointResponse{}, onekiterrors.PermissionDenied()
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
		if errors.Is(err, domain.ErrLeadNotFound) {
			return CreateTouchpointResponse{}, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		if errors.Is(err, domain.ErrContactNotFound) {
			return CreateTouchpointResponse{}, onekiterrors.New(http.StatusNotFound, "Contact not found")
		}
		logger.Ctx(ctx).Error("failed to create touchpoint", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateTouchpointResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create touchpoint")
	}

	return CreateTouchpointResponse{Touchpoint: payloadTouchpoint(tp)}, nil
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

func (h *DeleteTouchpointHandler) Serve(ctx context.Context, _ *DeleteTouchpointRequest) (DeleteTouchpointResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTouchpointResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:delete"); err != nil {
		return DeleteTouchpointResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteTouchpointResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid touchpoint id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteTouchpointResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.DeleteTouchpoint(ctx, id, businessID); err != nil {
		if errors.Is(err, domain.ErrTouchpointNotFound) {
			return DeleteTouchpointResponse{}, onekiterrors.New(http.StatusNotFound, "Touchpoint not found")
		}
		logger.Ctx(ctx).Error("failed to delete touchpoint", logger.Any("owner_id", device.User.UserID), logger.Any("touchpoint_id", id), logger.Error(err))
		return DeleteTouchpointResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete touchpoint")
	}

	return DeleteTouchpointResponse{Success: true}, nil
}
