package handlers

import (
	"context"
	"net/http"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
)

func CompleteOnboardingRoutes() []string {
	return []string{"/zontriq/user/onboard"}
}

type CompleteOnboardingHandler struct {
	UsersService *application.Service
}

func (h *CompleteOnboardingHandler) Serve(ctx context.Context, _ *usersv1.CompleteOnboardingRequest) (usersv1.CompleteOnboardingResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return usersv1.CompleteOnboardingResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	err := h.UsersService.CompleteOnboarding(ctx, int(device.User.UserID))
	if err != nil {
		return usersv1.CompleteOnboardingResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to complete onboarding",
		}
	}

	return usersv1.CompleteOnboardingResponse{Onboarded: true}, nil
}
