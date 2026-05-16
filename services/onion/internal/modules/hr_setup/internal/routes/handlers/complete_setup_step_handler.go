package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func CompleteSetupStepRoutes() []string { return []string{"/hr/setup/steps/{stepKey}/complete"} }

type CompleteSetupStepHandler struct{ Service *application.Service }

type CompleteSetupStepRequest struct{}

type CompleteSetupStepResponse struct{}

func (h *CompleteSetupStepHandler) Serve(ctx context.Context, _ *CompleteSetupStepRequest) (CompleteSetupStepResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CompleteSetupStepResponse{}, errorModels.PermissionDeniedServerError()
	}

	stepKey := chi.URLParamFromCtx(ctx, "stepKey")

	err := h.Service.CompleteSetupStep(ctx, device.User.UserID, stepKey)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidStepKey) {
			return CompleteSetupStepResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Invalid step key",
			}
		}
		logger.Ctx(ctx).Error("failed to complete setup step", logger.Any("owner_id", device.User.UserID), logger.String("step_key", stepKey), logger.Error(err))
		return CompleteSetupStepResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to complete setup step",
		}
	}

	return CompleteSetupStepResponse{}, nil
}
