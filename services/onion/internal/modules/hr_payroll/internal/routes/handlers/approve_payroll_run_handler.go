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
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func ApprovePayrollRunRoutes() []string { return []string{"/hr/payroll/runs/{id}/approve"} }

type ApprovePayrollRunHandler struct{ Service *application.Service }

type ApprovePayrollRunRequest struct {
	RunID int64 `json:"run_id"`
}

type ApprovePayrollRunResponse struct {
	Success bool `json:"success"`
}

func (h *ApprovePayrollRunHandler) Serve(ctx context.Context, _ *ApprovePayrollRunRequest) (ApprovePayrollRunResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApprovePayrollRunResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	runID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ApprovePayrollRunResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid payroll run id")
	}

	err := h.Service.ApprovePayrollRun(ctx, device.User.UserID, runID, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return ApprovePayrollRunResponse{}, onekiterrors.New(http.StatusNotFound, "Payroll run not found")
		}
		if errors.Is(err, domain.ErrApprovalChainNotConfigured) {
			return ApprovePayrollRunResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Approval chain not configured for payroll")
		}
		logger.Ctx(ctx).Error("failed to approve payroll run", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", runID), logger.Error(err))
		return ApprovePayrollRunResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to approve payroll run")
	}

	return ApprovePayrollRunResponse{Success: true}, nil
}
