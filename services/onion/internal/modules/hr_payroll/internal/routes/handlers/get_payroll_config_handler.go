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

func (h *GetPayrollConfigHandler) Serve(ctx context.Context, _ *GetPayrollConfigRequest) (GetPayrollConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetPayrollConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	cfg, err := h.Service.GetPayrollConfig(ctx, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrPayrollNotConfigured) {
			return GetPayrollConfigResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Payroll not configured for this company"}
		}
		logger.Ctx(ctx).Error("failed to get payroll config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetPayrollConfigResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get payroll config"}
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
