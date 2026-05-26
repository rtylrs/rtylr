package handlers

import (
	"context"
	"errors"
	"net/http"

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

func (h *ApprovePayrollRunHandler) Validate(_ context.Context, req *ApprovePayrollRunRequest) *errorModels.ServerError {
	if req.RunID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "run_id is required"}
	}
	return nil
}

func (h *ApprovePayrollRunHandler) Serve(ctx context.Context, req *ApprovePayrollRunRequest) (ApprovePayrollRunResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApprovePayrollRunResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.ApprovePayrollRun(ctx, device.User.UserID, req.RunID, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Payroll run not found"}
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Approval chain not configured for payroll"}
		}
		logger.Ctx(ctx).Error("failed to approve payroll run", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", req.RunID), logger.Error(err))
		return ApprovePayrollRunResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to approve payroll run"}
	}

	return ApprovePayrollRunResponse{Success: true}, nil
}
