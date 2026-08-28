package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, error) {
	device, authErr := requireSecurityPermission(ctx)
	if authErr != nil {
		return GetSettingsResponse{}, authErr
	}

	settings, err := h.Service.GetSettings(ctx, device.UserID)
	if err != nil {
		return GetSettingsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get security settings")
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

func (h *Toggle2FAHandler) Serve(ctx context.Context, req *Toggle2FARequest) (Toggle2FAResponse, error) {
	device, authErr := requireSecurityPermission(ctx)
	if authErr != nil {
		return Toggle2FAResponse{}, authErr
	}

	if err := h.Service.SetTwoFactor(ctx, device.UserID, req.Enabled); err != nil {
		return Toggle2FAResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update 2FA settings")
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
	LoginHistoryRequest struct {
		Page     int64 `json:"page"`
		PageSize int64 `json:"page_size"`
	}
)

type LoginHistoryResponse struct {
	History  []*hubsecurityv1.LoginEntry `json:"history"`
	Total    int32                       `json:"total"`
	Page     int32                       `json:"page"`
	PageSize int32                       `json:"page_size"`
	HasMore  bool                        `json:"has_more"`
}

func (h *LoginHistoryHandler) Serve(ctx context.Context, req *LoginHistoryRequest) (LoginHistoryResponse, error) {
	device, authErr := requireSecurityPermission(ctx)
	if authErr != nil {
		return LoginHistoryResponse{}, authErr
	}

	page, pageSize := application.NormalizeLoginHistoryPagination(req.Page, req.PageSize)
	entries, total, hasMore, err := h.Service.GetLoginHistory(ctx, device.UserID, page, pageSize)
	if err != nil {
		return LoginHistoryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get login history")
	}

	payloads := make([]*hubsecurityv1.LoginEntry, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, &hubsecurityv1.LoginEntry{
			Id:        e.ID,
			Email:     e.Email,
			Method:    e.Method,
			Ip:        e.IP,
			UserAgent: e.UserAgent,
			Success:   e.Success,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return LoginHistoryResponse{
		History:  payloads,
		Total:    int32(total),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Page:     int32(page),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PageSize: int32(pageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		HasMore:  hasMore,
	}, nil
}

// -- Invalidate All Sessions --

type (
	InvalidateSessionsHandler  struct{ Service *application.Service }
	InvalidateSessionsRequest  struct{}
	InvalidateSessionsResponse struct {
		Message string `json:"message"`
	}
)

func (h *InvalidateSessionsHandler) Serve(ctx context.Context, _ *InvalidateSessionsRequest) (InvalidateSessionsResponse, error) {
	device, authErr := requireSecurityPermission(ctx)
	if authErr != nil {
		return InvalidateSessionsResponse{}, authErr
	}

	if err := h.Service.InvalidateAllSessions(ctx, device.UserID); err != nil {
		return InvalidateSessionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to invalidate sessions")
	}

	return InvalidateSessionsResponse{Message: "All sessions invalidated"}, nil
}
