package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func ImportFlowRoutes() []string { return []string{"/hr/time-entries/import-flow"} }

type ImportFlowHandler struct{ Service *application.Service }

type ImportFlowRequest struct {
	EmployeeID  int64  `json:"employee_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

type ImportFlowResponse struct {
	Imported []*hrtimeentryv1.TimeEntryPayload `json:"imported"`
	Count    int32                             `json:"count"`
}

func (h *ImportFlowHandler) Validate(_ context.Context, req *ImportFlowRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.PeriodStart == "" {
		return onekiterrors.New(http.StatusBadRequest, "period_start is required")
	}
	if req.PeriodEnd == "" {
		return onekiterrors.New(http.StatusBadRequest, "period_end is required")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_end must be in YYYY-MM-DD format")
	}
	return nil
}

func (h *ImportFlowHandler) Serve(ctx context.Context, req *ImportFlowRequest) (ImportFlowResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ImportFlowResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ImportFlowResponse{}, onekiterrors.PermissionDenied()
	}

	entries, err := h.Service.ImportFlowTasks(ctx, businessID, req.EmployeeID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		logger.Ctx(ctx).Error("failed to import flow tasks",
			logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ImportFlowResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to import flow tasks")
	}

	payloads := make([]*hrtimeentryv1.TimeEntryPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, timeEntryToProto(e))
	}

	return ImportFlowResponse{Imported: payloads, Count: int32(len(payloads))}, nil
}
