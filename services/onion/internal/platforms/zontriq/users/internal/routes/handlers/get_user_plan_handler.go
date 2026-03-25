package handlers

import (
	"context"
	"net/http"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
)

func GetUserPlanRoutes() []string {
	return []string{"/zontriq/user/plan"}
}

type GetUserPlanHandler struct {
	UsersService *application.Service
}

func (h *GetUserPlanHandler) Serve(ctx context.Context, _ *usersv1.GetUserPlanRequest) (usersv1.GetUserPlanResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return usersv1.GetUserPlanResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	userID := int(device.User.UserID)

	plan, err := h.UsersService.GetUserPlanByUserID(ctx, userID)
	if err != nil {
		return usersv1.GetUserPlanResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Plan not found",
		}
	}

	return usersv1.GetUserPlanResponse{
		Plan: &usersv1.Plan{
			Id:                int64(plan.ID),
			Name:              plan.Name,
			DisplayName:       plan.DisplayName,
			PriceCents:        int32(plan.PriceCents),
			MaxVideosPerMonth: int32(plan.MaxVideosPerMonth),
			MaxChannels:       int32(plan.MaxChannels),
			MaxSeries:         int32(plan.MaxSeries),
			MaxAutomations:    int32(plan.MaxAutomations),
		},
	}, nil
}
