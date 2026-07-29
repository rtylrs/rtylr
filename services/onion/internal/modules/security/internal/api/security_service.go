package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubsecurityv1 "github.com/voxire/voxie/gen/go/hub/security/v1"
)

type SecurityAPI struct {
	Service *application.Service
}

func NewSecurityAPI(service *application.Service) *SecurityAPI {
	return &SecurityAPI{Service: service}
}

func (a *SecurityAPI) GetSecurityState(ctx context.Context, _ *hubsecurityv1.GetSecurityStateRequest) (*hubsecurityv1.GetSecurityStateResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	settings, err := a.Service.GetSettings(ctx, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get security settings")
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

	return &hubsecurityv1.GetSecurityStateResponse{
		State: &hubsecurityv1.SecurityState{
			TwoFactorEnabled: settings.TwoFactorEnabled,
			ActiveSessions:   sessions,
		},
	}, nil
}

func (a *SecurityAPI) LoginHistory(ctx context.Context, _ *hubsecurityv1.LoginHistoryRequest) (*hubsecurityv1.LoginHistoryResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	entries, err := a.Service.GetLoginHistory(ctx, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get login history")
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

	return &hubsecurityv1.LoginHistoryResponse{History: payloads}, nil
}

func (a *SecurityAPI) InvalidateSessions(ctx context.Context, _ *hubsecurityv1.InvalidateSessionsRequest) (*hubsecurityv1.InvalidateSessionsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.InvalidateAllSessions(ctx, device.User.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to invalidate sessions")
	}

	return &hubsecurityv1.InvalidateSessionsResponse{Message: "All sessions invalidated"}, nil
}

func (a *SecurityAPI) Update2FA(ctx context.Context, req *hubsecurityv1.Update2FARequest) (*hubsecurityv1.Update2FAResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.SetTwoFactor(ctx, device.User.UserID, req.GetEnabled()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update 2FA settings")
	}

	msg := "Two-factor authentication disabled"
	if req.GetEnabled() {
		msg = "Two-factor authentication enabled"
	}

	return &hubsecurityv1.Update2FAResponse{Message: msg}, nil
}
