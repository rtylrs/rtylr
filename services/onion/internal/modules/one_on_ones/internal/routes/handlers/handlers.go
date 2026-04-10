package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/one_on_ones/internal/application"
	"github.com/voxire/onion/internal/modules/one_on_ones/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type OneOnOnePayload struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	ManagerID   int64     `json:"managerId"`
	StaffID     int64     `json:"staffId"`
	ScheduledAt time.Time `json:"scheduledAt"`
	Notes       string    `json:"notes"`
	ActionItems string    `json:"actionItems"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

func newPayload(o domain.OneOnOne) OneOnOnePayload {
	return OneOnOnePayload{
		ID: o.ID, OwnerID: o.OwnerID, ManagerID: o.ManagerID, StaffID: o.StaffID,
		ScheduledAt: o.ScheduledAt, Notes: o.Notes, ActionItems: o.ActionItems,
		Status: o.Status, CreatedAt: o.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/one-on-ones"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct {
	OneOnOnes []OneOnOnePayload `json:"oneOnOnes"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
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

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	ManagerID   int64  `json:"managerId"`
	StaffID     int64  `json:"staffId"`
	ScheduledAt string `json:"scheduledAt"`
	Notes       string `json:"notes"`
	ActionItems string `json:"actionItems"`
	Status      string `json:"status"`
}
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
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.OneOnOneInput{
		ManagerID: req.ManagerID, StaffID: req.StaffID, ScheduledAt: req.ScheduledAt,
		Notes: req.Notes, ActionItems: req.ActionItems, Status: req.Status,
	}
	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create one-on-one"}
	}
	return CreateResponse{ID: id, Message: "One-on-one created successfully"}, nil
}
