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

func ExportPayrollRoutes() []string { return []string{"/hr/payroll/runs/{id}/export"} }

type ExportPayrollHandler struct{ Service *application.Service }

type ExportPayrollRequest struct {
	RunID int64 `json:"run_id"`
}

type ExportPayrollResponse struct {
	Success bool `json:"success"`
}

func (h *ExportPayrollHandler) Serve(ctx context.Context, req *ExportPayrollRequest) (ExportPayrollResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ExportPayrollResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.ExportPayrollToFinance(ctx, device.User.UserID, req.RunID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollRunNotFound) {
			return ExportPayrollResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Payroll run not found"}
		}
		if errors.Is(err, domain.ErrPayrollRunNotApproved) {
			return ExportPayrollResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnprocessableEntity, Code: http.StatusUnprocessableEntity, Message: "Payroll run must be approved before export"}
		}
		if errors.Is(err, domain.ErrPayrollAlreadyExported) {
			return ExportPayrollResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Payroll run has already been exported"}
		}
		logger.Ctx(ctx).Error("failed to export payroll", logger.Any("owner_id", device.User.UserID), logger.Any("run_id", req.RunID), logger.Error(err))
		return ExportPayrollResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to export payroll"}
	}

	return ExportPayrollResponse{Success: true}, nil
}
