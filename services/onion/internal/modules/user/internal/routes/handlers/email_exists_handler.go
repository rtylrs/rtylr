package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// EmailExistsRoutes is the anonymous endpoint the auth frontend calls to drive
// inline validation: "email already exists" on register, "email doesn't exist"
// on login. It only ever returns a boolean — never any account details.
//
// Note: this is a deliberate user-enumeration surface (any caller can probe
// whether an email is registered). It is gated by the global rate limiter and
// the tighter per-IP auth-endpoint budget applied at registration time.
func EmailExistsRoutes() []string {
	return []string{
		"/get-email-exists",
	}
}

type EmailExistsHandler struct {
	UserService *application.Service
}

// EmailExistsRequest reads the email from the `?email=` query param (GET query
// params are mapped onto json-tagged fields by the platform router).
type EmailExistsRequest struct {
	Email string `json:"email"`
}

type EmailExistsResponse struct {
	Exists bool `json:"exists"`
}

func (h *EmailExistsHandler) Validate(_ context.Context, req *EmailExistsRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Email) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "email is required",
		}
	}
	return nil
}

func (h *EmailExistsHandler) Serve(ctx context.Context, req *EmailExistsRequest) (EmailExistsResponse, *errorModels.ServerError) {
	exists, err := h.UserService.EmailExists(ctx, req.Email)
	if err != nil {
		return EmailExistsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to check email",
		}
	}
	return EmailExistsResponse{Exists: exists}, nil
}
