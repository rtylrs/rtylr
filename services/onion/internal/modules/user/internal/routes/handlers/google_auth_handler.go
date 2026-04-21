package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"

	authv1 "github.com/voxire/voxie/gen/go/auth/v1"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GoogleAuthRoutes is the single endpoint the frontend already calls
// (`POST /auth/register/google`). The name is a little misleading —
// it handles both login and signup since we can't tell the difference
// until we've resolved the verified email against our users table.
func GoogleAuthRoutes() []string {
	return []string{
		"/auth/register/google",
	}
}

type GoogleAuthHandler struct {
	UserService *application.Service
	// OnUserCreated fires on first-time signups only (IsNew=true).
	// Keep this aligned with the password-register hook so the
	// personal-workspace provisioning logic runs once.
	OnUserCreated func(ctx context.Context, userID int64)
	OnUserLogin   func(ctx context.Context, userID int64)
}

type GoogleAuthRequest struct {
	IDToken      string `json:"id_token"`
	BusinessName string `json:"business_name"`
}

type GoogleAuthResponse = authv1.LoginResponse

func (h *GoogleAuthHandler) Validate(_ context.Context, req *GoogleAuthRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.IDToken) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "id_token is required",
		}
	}
	return nil
}

func (h *GoogleAuthHandler) Serve(ctx context.Context, req *GoogleAuthRequest) (GoogleAuthResponse, *errorModels.ServerError) {
	// Signup IP is optional — the platform auth middleware doesn't
	// always populate it for anonymous routes, so we best-effort it.
	var signupIP netip.Addr

	result, err := h.UserService.GoogleAuth(ctx, req.IDToken, req.BusinessName, signupIP, "")
	if err != nil {
		switch {
		case errors.Is(err, application.ErrGoogleOAuthNotConfigured):
			// 503: we'd like to help but we haven't finished setting
			// up Google OAuth on the server. Distinct from 400 so
			// frontend can surface a "temporarily unavailable"
			// message rather than "your token is bad".
			return GoogleAuthResponse{}, &errorModels.ServerError{
				Code:    http.StatusServiceUnavailable,
				Message: "Google login is not configured on this server",
			}
		case errors.Is(err, application.ErrGoogleEmailUnverified):
			return GoogleAuthResponse{}, &errorModels.ServerError{
				Code:    http.StatusForbidden,
				Message: "Your Google email is not verified",
			}
		case errors.Is(err, application.ErrGoogleTokenInvalid):
			return GoogleAuthResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Google token is invalid or expired",
			}
		default:
			return GoogleAuthResponse{}, &errorModels.ServerError{
				Code:    http.StatusInternalServerError,
				Message: "Google authentication failed",
			}
		}
	}

	// Fire new-user / login hooks asynchronously — these are
	// idempotent (personal-workspace creation) and should never block
	// the user's first response.
	if result.IsNew && h.OnUserCreated != nil {
		go h.OnUserCreated(context.WithoutCancel(ctx), result.UserID)
	}
	if !result.IsNew && h.OnUserLogin != nil {
		go h.OnUserLogin(context.WithoutCancel(ctx), result.UserID)
	}

	user, uErr := h.UserService.GetUserByID(ctx, result.UserID)
	if uErr != nil {
		return GoogleAuthResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to load user after google auth",
		}
	}

	return GoogleAuthResponse{
		Token:   result.Token,
		Message: "Google authentication successful",
		UserId:  result.UserID,
		Email:   user.Email,
		Name:    user.FullName,
	}, nil
}
