package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func UpsertPayrollConfigRoutes() []string { return []string{"/hr/payroll/config"} }

type UpsertPayrollConfigHandler struct{ Service *application.Service }

type UpsertPayrollConfigRequest struct {
	PeriodType      string `json:"period_type"`
	PayDayOfMonth   int32  `json:"pay_day_of_month"`
	OvertimeEnabled bool   `json:"overtime_enabled"`
}

type UpsertPayrollConfigResponse struct {
	Success bool `json:"success"`
}

func (h *UpsertPayrollConfigHandler) Validate(_ context.Context, req *UpsertPayrollConfigRequest) *errorModels.ServerError {
	if req.PeriodType != "" && req.PeriodType != "weekly" && req.PeriodType != "biweekly" && req.PeriodType != "monthly" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_type must be weekly, biweekly, or monthly"}
	}
	if req.PeriodType == "monthly" && req.PayDayOfMonth != 0 && (req.PayDayOfMonth < 1 || req.PayDayOfMonth > 28) {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "pay_day_of_month must be between 1 and 28"}
	}
	return nil
}

func (h *UpsertPayrollConfigHandler) Serve(ctx context.Context, req *UpsertPayrollConfigRequest) (UpsertPayrollConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertPayrollConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	cfg := domain.PayrollConfigRecord{
		PeriodType:      req.PeriodType,
		PayDayOfMonth:   req.PayDayOfMonth,
		OvertimeEnabled: req.OvertimeEnabled,
	}

	err := h.Service.UpsertPayrollConfig(ctx, device.User.UserID, cfg)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert payroll config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UpsertPayrollConfigResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update payroll config"}
	}

	return UpsertPayrollConfigResponse{Success: true}, nil
}
