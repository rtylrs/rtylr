package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetOvertimeConfigRoutes() []string { return []string{"/hr/overtime/config"} }

type GetOvertimeConfigHandler struct{ Service *application.Service }

type GetOvertimeConfigRequest struct{}

type GetOvertimeConfigResponse struct {
	DailyThresholdMinutes int32   `json:"daily_threshold_minutes"`
	DefaultMultiplier     float64 `json:"default_multiplier"`
	AutoDetectEnabled     bool    `json:"auto_detect_enabled"`
}

func (h *GetOvertimeConfigHandler) Serve(ctx context.Context, _ *GetOvertimeConfigRequest) (GetOvertimeConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetOvertimeConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetOvertimeConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.GetOvertimeConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrOvertimeConfigNotFound) {
			return GetOvertimeConfigResponse{
				DailyThresholdMinutes: 480,
				DefaultMultiplier:     1.5,
				AutoDetectEnabled:     false,
			}, nil
		}
		return GetOvertimeConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get overtime config")
	}

	return GetOvertimeConfigResponse{
		DailyThresholdMinutes: cfg.DailyThresholdMinutes,
		DefaultMultiplier:     cfg.DefaultMultiplier,
		AutoDetectEnabled:     cfg.AutoDetectEnabled,
	}, nil
}
