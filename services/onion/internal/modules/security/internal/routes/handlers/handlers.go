package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	"github.com/voxire/onion/internal/modules/security/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SecurityRoutes() []string       { return []string{"/security"} }
func TwoFARoutes() []string          { return []string{"/security/2fa"} }
func LoginHistoryRoutes() []string   { return []string{"/security/login-history"} }
func InvalidateRoutes() []string     { return []string{"/security/invalidate-sessions"} }

// -- Get Security Settings --

type GetSettingsHandler struct{ Service *application.Service }
type GetSettingsRequest struct{}

type SessionPayload struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	CreatedAt string `json:"created_at"`
	IsCurrent bool   `json:"is_current"`
}

type GetSettingsResponse struct {
	TwoFactorEnabled bool             `json:"two_factor_enabled"`
	ActiveSessions   []SessionPayload `json:"active_sessions"`
}

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	settings, err := h.Service.GetSettings(ctx, device.User.UserID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get security settings",
		}
	}

	sessions := make([]SessionPayload, 0, len(settings.ActiveSessions))
	for _, s := range settings.ActiveSessions {
		sessions = append(sessions, SessionPayload{
			ID:        s.ID,
			IP:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			IsCurrent: s.IsCurrent,
		})
	}

	return GetSettingsResponse{
		TwoFactorEnabled: settings.TwoFactorEnabled,
		ActiveSessions:   sessions,
	}, nil
}

// -- Toggle 2FA --

type Toggle2FAHandler struct{ Service *application.Service }

type Toggle2FARequest struct {
	Enabled bool `json:"enabled"`
}

type Toggle2FAResponse struct {
	Message string `json:"message"`
}

func (h *Toggle2FAHandler) Serve(ctx context.Context, req *Toggle2FARequest) (Toggle2FAResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return Toggle2FAResponse{}, errorModels.PermissionDeniedServerError()
	}

	_ = domain.TwoFactorInput{Enabled: req.Enabled}

	if err := h.Service.SetTwoFactor(ctx, device.User.UserID, req.Enabled); err != nil {
		return Toggle2FAResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update 2FA settings",
		}
	}

	msg := "Two-factor authentication disabled"
	if req.Enabled {
		msg = "Two-factor authentication enabled"
	}

	return Toggle2FAResponse{Message: msg}, nil
}

// -- Login History --

type LoginHistoryHandler struct{ Service *application.Service }
type LoginHistoryRequest struct{}

type LoginEntryPayload struct {
	Email     string `json:"email"`
	Method    string `json:"method"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Success   bool   `json:"success"`
	CreatedAt string `json:"created_at"`
}

type LoginHistoryResponse struct {
	History []LoginEntryPayload `json:"history"`
}

func (h *LoginHistoryHandler) Serve(ctx context.Context, _ *LoginHistoryRequest) (LoginHistoryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return LoginHistoryResponse{}, errorModels.PermissionDeniedServerError()
	}

	entries, err := h.Service.GetLoginHistory(ctx, device.User.UserID)
	if err != nil {
		return LoginHistoryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get login history",
		}
	}

	payloads := make([]LoginEntryPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, LoginEntryPayload{
			Email:     e.Email,
			Method:    e.Method,
			IP:        e.IP,
			UserAgent: e.UserAgent,
			Success:   e.Success,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return LoginHistoryResponse{History: payloads}, nil
}

// -- Invalidate All Sessions --

type InvalidateSessionsHandler struct{ Service *application.Service }
type InvalidateSessionsRequest struct{}
type InvalidateSessionsResponse struct {
	Message string `json:"message"`
}

func (h *InvalidateSessionsHandler) Serve(ctx context.Context, _ *InvalidateSessionsRequest) (InvalidateSessionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return InvalidateSessionsResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.InvalidateAllSessions(ctx, device.User.UserID); err != nil {
		return InvalidateSessionsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to invalidate sessions",
		}
	}

	return InvalidateSessionsResponse{Message: "All sessions invalidated"}, nil
}
