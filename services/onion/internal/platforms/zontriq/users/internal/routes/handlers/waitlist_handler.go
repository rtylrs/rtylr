package handlers

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
)

func WaitlistRoutes() []string {
	return []string{"/zontriq/waitlist"}
}

type WaitlistRequest struct {
	Email  string `json:"email"`
	Source string `json:"source"`
}

type WaitlistResponse struct {
	Success bool `json:"success"`
}

type WaitlistHandler struct {
	UsersService *application.Service
}

func (h *WaitlistHandler) Serve(ctx context.Context, req *WaitlistRequest) (WaitlistResponse, *errorModels.ServerError) {
	if req.Email == "" {
		return WaitlistResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "email is required",
		}
	}

	source := req.Source
	if source == "" {
		source = "landing"
	}

	err := h.UsersService.CreateWaitlistEntry(ctx, req.Email, source)
	if err != nil {
		return WaitlistResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to join waitlist",
		}
	}

	return WaitlistResponse{Success: true}, nil
}
