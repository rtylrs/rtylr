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

func ExportPayrollRoutes() []string { return []string{"/hr/payroll/runs/{id}/export"} }

type ExportPayrollHandler struct{ Service *application.Service }

type ExportPayrollRequest struct {
	RunID int64 `json:"run_id"`
}

type ExportPayrollResponse struct {
	Success bool `json:"success"`
}

func (h *ExportPayrollHandler) Serve(ctx context.Context, _ *ExportPayrollRequest) (ExportPayrollResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ExportPayrollResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	runID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ExportPayrollResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid payroll run id")
	}

	err := h.Service.ExportPayrollToFinance(ctx, device.User.UserID, runID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return ExportPayrollResponse{}, onekiterrors.New(http.StatusNotFound, "Payroll run not found")
		}
		if errors.Is(err, domain.ErrPayrollRunNotApproved) {
			return ExportPayrollResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Payroll run must be approved before export")
		}
		if errors.Is(err, domain.ErrPayrollAlreadyExported) {
			return ExportPayrollResponse{}, onekiterrors.New(http.StatusConflict, "Payroll run has already been exported")
		}
		if errors.Is(err, domain.ErrPayrollExportInProgress) {
			return ExportPayrollResponse{}, onekiterrors.New(http.StatusConflict, "Payroll run export is already in progress")
		}
		if errors.Is(err, domain.ErrPayrollEmployeeInactive) {
			return ExportPayrollResponse{}, onekiterrors.New(http.StatusConflict, "Payroll contains an inactive or missing employee")
		}
		if errors.Is(err, domain.ErrFinanceIntegrationUnavailable) {
			return ExportPayrollResponse{}, onekiterrors.New(http.StatusServiceUnavailable, "Finance integration is unavailable")
		}
		logger.Ctx(ctx).Error("failed to export payroll", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", runID), logger.Error(err))
		return ExportPayrollResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export payroll")
	}

	return ExportPayrollResponse{Success: true}, nil
}
