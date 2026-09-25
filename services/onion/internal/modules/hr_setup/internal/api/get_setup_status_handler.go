package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetSetupStatusRoutes() []string { return []string{"/hr/setup/status"} }

type GetSetupStatusHandler struct{ Service *application.Service }

type GetSetupStatusRequest struct{}

type SetupStepPayload struct {
	StepKey     string `json:"step_key"`
	CompletedAt *int64 `json:"completed_at"`
}

type GetSetupStatusResponse struct {
	IsComplete bool                `json:"is_complete"`
	Steps      []*SetupStepPayload `json:"steps"`
}

func (h *GetSetupStatusHandler) Serve(ctx context.Context, _ *GetSetupStatusRequest) (GetSetupStatusResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetSetupStatusResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetSetupStatusResponse{}, onekiterrors.PermissionDenied()
	}

	steps, isComplete, err := h.Service.GetSetupStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get setup status", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetSetupStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get setup status")
	}

	payloads := make([]*SetupStepPayload, 0, len(steps))
	for _, s := range steps {
		p := &SetupStepPayload{StepKey: s.StepKey}
		if s.CompletedAt != nil {
			unix := s.CompletedAt.Unix()
			p.CompletedAt = &unix
		}
		payloads = append(payloads, p)
	}

	return GetSetupStatusResponse{IsComplete: isComplete, Steps: payloads}, nil
}
