package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

// PasswordResetRequestRoutes returns the route for requesting a password reset email.
func PasswordResetRequestRoutes() []string {
	return []string{"/auth/password-reset/request"}
}

// PasswordResetConfirmRoutes returns the route for confirming a password reset.
func PasswordResetConfirmRoutes() []string {
	return []string{"/auth/password-reset/confirm"}
}

// PasswordResetRequestHandler handles POST /auth/password-reset/request.
// Always returns 200 regardless of whether the email exists, to prevent
// user enumeration.
type PasswordResetRequestHandler struct {
	UserService *application.Service
}

type PasswordResetRequestBody struct {
	Email string `json:"email"`
}

type PasswordResetRequestResponse struct {
	Message string `json:"message"`
}

func (h *PasswordResetRequestHandler) Validate(_ context.Context, req *PasswordResetRequestBody) *errorModels.ServerError {
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	return nil
}

func (h *PasswordResetRequestHandler) Serve(ctx context.Context, req *PasswordResetRequestBody) (PasswordResetRequestResponse, *errorModels.ServerError) {
	if err := h.UserService.RequestPasswordReset(ctx, req.Email); err != nil {
		// Always return 200 to prevent user enumeration; log the failure for ops.
		logger.Ctx(ctx).Error("password reset request failed",
			logger.String("email", req.Email),
			logger.Error(err),
		)
	}
	return PasswordResetRequestResponse{Message: "If that email is registered, a reset link has been sent."}, nil
}

// PasswordResetConfirmHandler handles POST /auth/password-reset/confirm.
type PasswordResetConfirmHandler struct {
	UserService *application.Service
}

type PasswordResetConfirmBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type PasswordResetConfirmResponse struct {
	Message string `json:"message"`
}

func (h *PasswordResetConfirmHandler) Validate(_ context.Context, req *PasswordResetConfirmBody) *errorModels.ServerError {
	if req.Token == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "token is required"}
	}
	if req.Password == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "password is required"}
	}
	return nil
}

func (h *PasswordResetConfirmHandler) Serve(ctx context.Context, req *PasswordResetConfirmBody) (PasswordResetConfirmResponse, *errorModels.ServerError) {
	err := h.UserService.ResetPassword(ctx, req.Token, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrResetTokenNotFound):
			return PasswordResetConfirmResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Reset link is invalid or has expired. Please request a new one.",
			}
		case errors.Is(err, application.ErrWeakPassword):
			return PasswordResetConfirmResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Password is too weak. Use at least 8 characters.",
			}
		default:
			return PasswordResetConfirmResponse{}, &errorModels.ServerError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to reset password.",
			}
		}
	}
	return PasswordResetConfirmResponse{Message: "Password reset successfully."}, nil
}
