package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpsertOvertimeConfigRoutes() []string { return []string{"/hr/overtime/config"} }

type UpsertOvertimeConfigHandler struct{ Service *application.Service }

type UpsertOvertimeConfigRequest struct {
	DailyThresholdMinutes int32   `json:"daily_threshold_minutes"`
	DefaultMultiplier     float64 `json:"default_multiplier"`
	AutoDetectEnabled     bool    `json:"auto_detect_enabled"`
}

type UpsertOvertimeConfigResponse struct {
	Success bool `json:"success"`
}

func (h *UpsertOvertimeConfigHandler) Validate(_ context.Context, req *UpsertOvertimeConfigRequest) error {
	if req.DailyThresholdMinutes <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "daily_threshold_minutes must be greater than 0")
	}
	if req.DefaultMultiplier <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "default_multiplier must be greater than 0")
	}
	return nil
}

func (h *UpsertOvertimeConfigHandler) Serve(ctx context.Context, req *UpsertOvertimeConfigRequest) (UpsertOvertimeConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertOvertimeConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpsertOvertimeConfigResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.UpsertOvertimeConfig(ctx, businessID, device.User.UserID, domain.OvertimeConfigRecord{
		DailyThresholdMinutes: req.DailyThresholdMinutes,
		DefaultMultiplier:     req.DefaultMultiplier,
		AutoDetectEnabled:     req.AutoDetectEnabled,
	})
	if err != nil {
		if errors.Is(err, domain.ErrOvertimeAccessDenied) {
			return UpsertOvertimeConfigResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to upsert overtime config", logger.Any("business_id", businessID), logger.Error(err))
		return UpsertOvertimeConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update overtime config")
	}

	return UpsertOvertimeConfigResponse{Success: true}, nil
}
