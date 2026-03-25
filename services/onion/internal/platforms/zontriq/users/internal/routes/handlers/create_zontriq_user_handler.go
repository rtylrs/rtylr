package handlers

import (
	"context"
	"net/http"
	"time"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
)

func CreateZontriqUserRoutes() []string {
	return []string{
		"/zontriq/user",
	}
}

type CreateZontriqUserHandler struct {
	UsersService *application.Service
}

func (h *CreateZontriqUserHandler) Serve(ctx context.Context, req *usersv1.CreateZontriqUserRequest) (usersv1.CreateZontriqUserResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return usersv1.CreateZontriqUserResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	userID := int(device.User.UserID)
	planID := int(req.PlanId)

	if planID == 0 {
		return usersv1.CreateZontriqUserResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "plan_id is required",
		}
	}

	created, err := h.UsersService.CreateZontriqUser(ctx, userID, planID)
	if err != nil {
		return usersv1.CreateZontriqUserResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create zontriq user",
		}
	}

	return usersv1.CreateZontriqUserResponse{
		User: &usersv1.ZontriqUser{
			Id:        int64(created.ID),
			UserId:    int64(created.UserID),
			Status:    int32(created.Status),
			CreatedAt: created.CreatedAt.Format(time.RFC3339),
			UpdatedAt: created.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}
