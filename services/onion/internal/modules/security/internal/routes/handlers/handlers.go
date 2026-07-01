package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	hubsecurityv1 "github.com/voxire/voxie/gen/go/hub/security/v1"
)

func SecurityRoutes() []string     { return []string{"/security"} }
func TwoFARoutes() []string        { return []string{"/security/2fa"} }
func LoginHistoryRoutes() []string { return []string{"/security/login-history"} }
func InvalidateRoutes() []string   { return []string{"/security/invalidate-sessions"} }

// -- Get Security Settings --

type (
	GetSettingsHandler struct{ Service *application.Service }
	GetSettingsRequest struct{}
)

type GetSettingsResponse struct {
	State *hubsecurityv1.SecurityState `json:"state"`
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

	sessions := make([]*hubsecurityv1.ActiveSession, 0, len(settings.ActiveSessions))
	for _, s := range settings.ActiveSessions {
		sessions = append(sessions, &hubsecurityv1.ActiveSession{
			Id:        s.ID,
			Ip:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			IsCurrent: s.IsCurrent,
		})
	}

	return GetSettingsResponse{
		State: &hubsecurityv1.SecurityState{
			TwoFactorEnabled: settings.TwoFactorEnabled,
			ActiveSessions:   sessions,
		},
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

type (
	LoginHistoryHandler struct{ Service *application.Service }
	LoginHistoryRequest struct{}
)

type LoginHistoryResponse struct {
	History []*hubsecurityv1.LoginEntry `json:"history"`
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

	payloads := make([]*hubsecurityv1.LoginEntry, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, &hubsecurityv1.LoginEntry{
			Email:     e.Email,
			Method:    e.Method,
			Ip:        e.IP,
			UserAgent: e.UserAgent,
			Success:   e.Success,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return LoginHistoryResponse{History: payloads}, nil
}

// -- Invalidate All Sessions --

type (
	InvalidateSessionsHandler  struct{ Service *application.Service }
	InvalidateSessionsRequest  struct{}
	InvalidateSessionsResponse struct {
		Message string `json:"message"`
	}
)

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
