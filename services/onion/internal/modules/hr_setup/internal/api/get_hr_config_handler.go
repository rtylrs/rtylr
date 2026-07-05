package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func GetHRConfigRoutes() []string { return []string{"/hr/config"} }

type GetHRConfigHandler struct{ Service *application.Service }

type GetHRConfigRequest struct{}

type HRConfigPayload struct {
	ID               int64  `json:"id"`
	OwnerID          int64  `json:"owner_id"`
	SetupCompletedAt *int64 `json:"setup_completed_at"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

type GetHRConfigResponse struct {
	Config *HRConfigPayload `json:"config"`
}

func (h *GetHRConfigHandler) Serve(ctx context.Context, _ *GetHRConfigRequest) (GetHRConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.GetHRConfig(ctx, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return GetHRConfigResponse{Config: nil}, nil
		}
		logger.Ctx(ctx).Error("failed to get hr config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get HR config")
	}

	return GetHRConfigResponse{Config: toConfigPayload(cfg)}, nil
}

func toConfigPayload(cfg *domain.HRConfig) *HRConfigPayload {
	p := &HRConfigPayload{
		ID:        cfg.ID,
		OwnerID:   cfg.OwnerID,
		CreatedAt: cfg.CreatedAt.Unix(),
		UpdatedAt: cfg.UpdatedAt.Unix(),
	}
	if cfg.SetupCompletedAt != nil {
		unix := cfg.SetupCompletedAt.Unix()
		p.SetupCompletedAt = &unix
	}
	return p
}
