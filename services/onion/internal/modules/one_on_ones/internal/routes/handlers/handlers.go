package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/one_on_ones/internal/application"
	"github.com/voxire/onion/internal/modules/one_on_ones/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

var scheduledAtLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func isValidScheduledAt(s string) bool {
	for _, layout := range scheduledAtLayouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

type OneOnOnePayload struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	ManagerID   int64     `json:"manager_id"`
	StaffID     int64     `json:"staff_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Notes       string    `json:"notes"`
	ActionItems string    `json:"action_items"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func newPayload(o domain.OneOnOne) OneOnOnePayload {
	return OneOnOnePayload{
		ID: o.ID, OwnerID: o.OwnerID, ManagerID: o.ManagerID, StaffID: o.StaffID,
		ScheduledAt: o.ScheduledAt, Notes: o.Notes, ActionItems: o.ActionItems,
		Status: o.Status, CreatedAt: o.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/one-on-ones"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		OneOnOnes []OneOnOnePayload `json:"one_on_ones"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list one-on-ones"}
	}
	payloads := make([]OneOnOnePayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{OneOnOnes: payloads}, nil
}

func CreateRoutes() []string { return []string{"/hr/one-on-ones"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		ManagerID   int64  `json:"manager_id"`
		StaffID     int64  `json:"staff_id"`
		ScheduledAt string `json:"scheduled_at"`
		Notes       string `json:"notes"`
		ActionItems string `json:"action_items"`
		Status      string `json:"status"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.ManagerID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "managerId is required"}
	}
	if req.StaffID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "staffId is required"}
	}
	if req.ScheduledAt == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "scheduledAt is required"}
	}
	if !isValidScheduledAt(req.ScheduledAt) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "scheduledAt is not a valid date"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.OneOnOneInput{
		ManagerID: req.ManagerID, StaffID: req.StaffID, ScheduledAt: req.ScheduledAt,
		Notes: req.Notes, ActionItems: req.ActionItems, Status: req.Status,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create one-on-one"}
	}
	return CreateResponse{ID: id, Message: "One-on-one created successfully"}, nil
}
