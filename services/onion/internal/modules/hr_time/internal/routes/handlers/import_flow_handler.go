package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *ImportFlowHandler) Validate(_ context.Context, req *ImportFlowRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if req.PeriodStart == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_start is required"}
	}
	if req.PeriodEnd == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_end is required"}
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_start must be in YYYY-MM-DD format"}
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_end must be in YYYY-MM-DD format"}
	}
	return nil
}

func (h *ImportFlowHandler) Serve(ctx context.Context, req *ImportFlowRequest) (ImportFlowResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ImportFlowResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ImportFlowResponse{}, errorModels.PermissionDeniedServerError()
	}

	entries, err := h.Service.ImportFlowTasks(ctx, businessID, req.EmployeeID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		logger.Ctx(ctx).Error("failed to import flow tasks",
			logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ImportFlowResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to import flow tasks"}
	}

	payloads := make([]*hrtimeentryv1.TimeEntryPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, timeEntryToProto(e))
	}

	return ImportFlowResponse{Imported: payloads, Count: int32(len(payloads))}, nil
}
