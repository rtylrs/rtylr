package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

type SetupStepPayload struct {
	ID          int64  `json:"id"`
	StepKey     string `json:"step_key"`
	StepLabel   string `json:"step_label"`
	IsCompleted bool   `json:"is_completed"`
	CompletedAt string `json:"completed_at"`
}

func stepToPayload(s domain.SetupStepRecord) *SetupStepPayload {
	completedAt := ""
	if s.IsCompleted {
		completedAt = s.CompletedAt.Format("2006-01-02T15:04:05Z")
	}
	return &SetupStepPayload{
		ID: s.ID, StepKey: s.StepKey, StepLabel: s.StepLabel,
		IsCompleted: s.IsCompleted, CompletedAt: completedAt,
	}
}

// --- Get Setup Status ---

func GetSetupStatusRoutes() []string { return []string{"/hr/setup/status"} }

type GetSetupStatusHandler struct{ Service *application.Service }
type GetSetupStatusRequest struct{}
type GetSetupStatusResponse struct {
	Steps []*SetupStepPayload `json:"steps"`
}

func (h *GetSetupStatusHandler) Serve(ctx context.Context, _ *GetSetupStatusRequest) (GetSetupStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetSetupStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	steps, err := h.Service.GetSetupStatus(ctx, device.User.UserID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get setup status", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetSetupStatusResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get setup status"}
	}

	payloads := make([]*SetupStepPayload, 0, len(steps))
	for _, s := range steps {
		payloads = append(payloads, stepToPayload(s))
	}
	return GetSetupStatusResponse{Steps: payloads}, nil
}

// --- Complete Setup Step ---

func CompleteSetupStepRoutes() []string { return []string{"/hr/setup/steps/{step_key}/complete"} }

type CompleteSetupStepHandler struct{ Service *application.Service }
type CompleteSetupStepRequest struct{}
type CompleteSetupStepResponse struct {
	Step *SetupStepPayload `json:"step"`
}

func (h *CompleteSetupStepHandler) Serve(ctx context.Context, _ *CompleteSetupStepRequest) (CompleteSetupStepResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CompleteSetupStepResponse{}, errorModels.PermissionDeniedServerError()
	}

	stepKey := chi.URLParamFromCtx(ctx, "step_key")
	if stepKey == "" {
		return CompleteSetupStepResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "step_key is required"}
	}

	step, err := h.Service.CompleteSetupStep(ctx, device.User.UserID, stepKey)
	if err != nil {
		logger.Ctx(ctx).Error("failed to complete setup step", logger.Any("owner_id", device.User.UserID), logger.String("step_key", stepKey), logger.Error(err))
		return CompleteSetupStepResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to complete setup step"}
	}

	return CompleteSetupStepResponse{Step: stepToPayload(*step)}, nil
}
