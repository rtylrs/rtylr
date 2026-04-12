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

func AccountRoutes() []string {
	return []string{
		"/account",
		"/user/account",
	}
}

type AccountHandler struct {
	UserService *application.Service
}

type GetProfileRequest struct {
	UserID int64 `json:"user_id"`
}

type ProfileUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetProfileResponse struct {
	User *ProfileUser `json:"user"`
}

func newProfileUser(user *domain.User) *ProfileUser {
	if user == nil {
		return nil
	}

	return &ProfileUser{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.FullName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (h *AccountHandler) Serve(ctx context.Context, _ *GetProfileRequest) (GetProfileResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	user, err := h.UserService.GetUserByID(ctx, device.User.UserID)
	if err != nil {
		return GetProfileResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user account",
		}
	}

	return GetProfileResponse{
		User: newProfileUser(user),
	}, nil
}
