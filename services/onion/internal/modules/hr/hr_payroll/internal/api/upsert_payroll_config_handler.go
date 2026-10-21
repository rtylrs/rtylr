package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (h *UpsertPayrollConfigHandler) Validate(_ context.Context, req *UpsertPayrollConfigRequest) error {
	if req.PeriodType != "" && req.PeriodType != "weekly" && req.PeriodType != "biweekly" && req.PeriodType != "monthly" {
		return onekiterrors.New(http.StatusBadRequest, "period_type must be weekly, biweekly, or monthly")
	}
	if req.PeriodType == "monthly" && req.PayDayOfMonth != 0 && (req.PayDayOfMonth < 1 || req.PayDayOfMonth > 28) {
		return onekiterrors.New(http.StatusBadRequest, "pay_day_of_month must be between 1 and 28")
	}
	return nil
}

func (h *UpsertPayrollConfigHandler) Serve(ctx context.Context, req *UpsertPayrollConfigRequest) (UpsertPayrollConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertPayrollConfigResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requirePayrollPermission(ctx, "hr:payroll:manage"); err != nil {
		return UpsertPayrollConfigResponse{}, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return UpsertPayrollConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg := domain.PayrollConfigRecord{
		PeriodType:      req.PeriodType,
		PayDayOfMonth:   req.PayDayOfMonth,
		OvertimeEnabled: req.OvertimeEnabled,
	}

	err := h.Service.UpsertPayrollConfig(ctx, businessID, cfg)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert payroll config", logger.Any("business_id", businessID), logger.Error(err))
		return UpsertPayrollConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update payroll config")
	}

	return UpsertPayrollConfigResponse{Success: true}, nil
}
