package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	"github.com/voxire/onion/internal/modules/user/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func UpdateProfileRoutes() []string {
	return []string{
		"/account/profile",
	}
}

type UpdateProfileHandler struct {
	UserService *application.Service
}

type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdatedProfileUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateProfileResponse struct {
	User *UpdatedProfileUser `json:"user"`
}

func newUpdatedProfileUser(user *domain.User) *UpdatedProfileUser {
	if user == nil {
		return nil
	}

	return &UpdatedProfileUser{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.FullName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (h *UpdateProfileHandler) Validate(_ context.Context, req *UpdateProfileRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
	}

	return nil
}

func (h *UpdateProfileHandler) Serve(ctx context.Context, req *UpdateProfileRequest) (UpdateProfileResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	user, err := h.UserService.UpdateProfile(ctx, device.User.UserID, req.Name)
	if err != nil {
		return UpdateProfileResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update profile",
		}
	}

	return UpdateProfileResponse{
		User: newUpdatedProfileUser(user),
	}, nil
}
