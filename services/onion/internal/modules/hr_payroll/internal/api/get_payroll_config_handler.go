package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetPayrollConfigRoutes() []string { return []string{"/hr/payroll/config"} }

type GetPayrollConfigHandler struct{ Service *application.Service }

type GetPayrollConfigRequest struct{}

type GetPayrollConfigResponse struct {
	ID              int64  `json:"id"`
	OwnerID         int64  `json:"owner_id"`
	PeriodType      string `json:"period_type"`
	PayDayOfMonth   int32  `json:"pay_day_of_month"`
	OvertimeEnabled bool   `json:"overtime_enabled"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

func (h *GetPayrollConfigHandler) Serve(ctx context.Context, _ *GetPayrollConfigRequest) (GetPayrollConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetPayrollConfigResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requirePayrollPermission(ctx, "hr:payroll:view"); err != nil {
		return GetPayrollConfigResponse{}, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return GetPayrollConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.GetPayrollConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollNotConfigured) {
			return GetPayrollConfigResponse{}, onekiterrors.New(http.StatusNotFound, "Payroll not configured for this company")
		}
		logger.Ctx(ctx).Error("failed to get payroll config", logger.Any("business_id", businessID), logger.Error(err))
		return GetPayrollConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get payroll config")
	}

	return GetPayrollConfigResponse{
		ID:              cfg.ID,
		OwnerID:         cfg.OwnerID,
		PeriodType:      cfg.PeriodType,
		PayDayOfMonth:   cfg.PayDayOfMonth,
		OvertimeEnabled: cfg.OvertimeEnabled,
		CreatedAt:       cfg.CreatedAt.Unix(),
		UpdatedAt:       cfg.UpdatedAt.Unix(),
	}, nil
}
