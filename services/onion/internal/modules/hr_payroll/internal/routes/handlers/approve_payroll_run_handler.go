package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func ApprovePayrollRunRoutes() []string { return []string{"/hr/payroll/runs/{id}/approve"} }

type ApprovePayrollRunHandler struct{ Service *application.Service }

type ApprovePayrollRunRequest struct {
	RunID int64 `json:"run_id"`
}

type ApprovePayrollRunResponse struct {
	Success bool `json:"success"`
}

func (h *ApprovePayrollRunHandler) Serve(ctx context.Context, req *ApprovePayrollRunRequest) (ApprovePayrollRunResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApprovePayrollRunResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	runID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid payroll run id"}
	}

	err := h.Service.ApprovePayrollRun(ctx, device.User.UserID, runID, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Payroll run not found"}
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Approval chain not configured for payroll"}
		}
		logger.Ctx(ctx).Error("failed to approve payroll run", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", runID), logger.Error(err))
		return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to approve payroll run"}
	}

	return ApprovePayrollRunResponse{Success: true}, nil
}
