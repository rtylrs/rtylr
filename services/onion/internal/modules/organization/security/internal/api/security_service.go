package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/security/internal/application"
	"github.com/voxire/onion/internal/modules/organization/security/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubsecurityv1 "github.com/voxire/voxie/gen/go/hub/security/v1"
)

type SecurityAPI struct {
	Service securityService
}

// securityService is the application surface the handlers call. It is held
// behind an interface so the step-up responses can be asserted without a
// database.
type securityService interface {
	GetSettings(ctx context.Context, userID int64) (domain.SecuritySettings, error)
	GetLoginHistory(ctx context.Context, userID int64, page, pageSize int64) ([]domain.LoginEntry, int64, bool, error)
	InvalidateAllSessions(ctx context.Context, userID int64) error
	SetTwoFactor(ctx context.Context, userID int64, enabled bool) error
	BeginMFAEnrollment(ctx context.Context, userID int64) (string, string, error)
	CompleteMFAEnrollment(ctx context.Context, userID int64, code string) ([]string, error)
	VerifyMFA(ctx context.Context, userID int64, code string) error
	UseMFARecoveryCode(ctx context.Context, userID int64, code string) error
	GetOrganizationMFAEnforced(ctx context.Context, businessID int64) (bool, error)
	SetOrganizationMFAEnforced(ctx context.Context, businessID int64, enforced bool) error
}

func NewSecurityAPI(service *application.Service) *SecurityAPI {
	return &SecurityAPI{Service: service}
}

func requireAuthenticatedSecurityUser(ctx context.Context) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	return device.User, nil
}

// securityGuardError maps the step-up and attempt-limit failures raised before
// any state change to the status the client sees. A caller holding a valid
// session but no factor proof is forbidden, not unauthenticated, and a state
// read that could not be completed is a dependency outage rather than a
// permission problem. Every other error is left to the handler's own mapping.
func securityGuardError(err error) error {
	switch {
	case errors.Is(err, application.ErrStepUpRequired):
		return onekiterrors.New(http.StatusForbidden, "Recent MFA verification required before changing security settings")
	case errors.Is(err, application.ErrStepUpUnavailable):
		return onekiterrors.New(http.StatusServiceUnavailable, "Security settings are temporarily unavailable")
	case errors.Is(err, application.ErrMFAAttemptsExceeded):
		return onekiterrors.New(http.StatusTooManyRequests, "Too many MFA code attempts, try again later")
	default:
		return nil
	}
}

func canManageOrganizationMFA(user *authPkg.AuthUser) bool {
	return authPkg.HasDashboardPermission(user, "hub:security:manage") || authPkg.IsBusinessAdmin(user)
}

func requireOrganizationSecurityPermission(ctx context.Context) (*authPkg.AuthUser, error) {
	user, err := requireAuthenticatedSecurityUser(ctx)
	if err != nil {
		return nil, err
	}
	if !canManageOrganizationMFA(user) {
		return nil, onekiterrors.PermissionDenied()
	}
	return user, nil
}

func (a *SecurityAPI) GetSecurityState(ctx context.Context, _ *hubsecurityv1.GetSecurityStateRequest) (*hubsecurityv1.GetSecurityStateResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	settings, err := a.Service.GetSettings(ctx, user.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get security settings")
	}
	organizationMFAEnforced, err := a.Service.GetOrganizationMFAEnforced(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get organization security settings")
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
			TwoFactorEnabled:         settings.TwoFactorEnabled,
			MfaRequired:              settings.MFARequired || organizationMFAEnforced,
			MfaEnrolled:              settings.MFAEnrolled,
			OrganizationMfaEnforced:  organizationMFAEnforced,
			CanManageOrganizationMfa: canManageOrganizationMFA(user),
			ActiveSessions:           sessions,
		},
	}, nil
}

func (a *SecurityAPI) LoginHistory(ctx context.Context, req *hubsecurityv1.LoginHistoryRequest) (*hubsecurityv1.LoginHistoryResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}

	page, pageSize := application.NormalizeLoginHistoryPagination(
		int64(req.GetPage()),
		int64(req.GetPageSize()),
	)
	entries, total, hasMore, err := a.Service.GetLoginHistory(ctx, user.UserID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get login history")
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

	return &hubsecurityv1.LoginHistoryResponse{
		History:  payloads,
		Total:    int32(total),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Page:     int32(page),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PageSize: int32(pageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		HasMore:  hasMore,
	}, nil
}

func (a *SecurityAPI) InvalidateSessions(ctx context.Context, _ *hubsecurityv1.InvalidateSessionsRequest) (*hubsecurityv1.InvalidateSessionsResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}

	if err := a.Service.InvalidateAllSessions(ctx, user.UserID); err != nil {
		if guardErr := securityGuardError(err); guardErr != nil {
			return nil, guardErr
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to invalidate sessions")
	}

	return &hubsecurityv1.InvalidateSessionsResponse{Message: "All sessions invalidated"}, nil
}

func (a *SecurityAPI) Update2FA(ctx context.Context, req *hubsecurityv1.Update2FARequest) (*hubsecurityv1.Update2FAResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}
	if !req.GetEnabled() {
		businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
		if businessErr != nil {
			return nil, onekiterrors.PermissionDenied()
		}
		organizationMFAEnforced, policyErr := a.Service.GetOrganizationMFAEnforced(ctx, businessID)
		if policyErr != nil {
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get organization security settings")
		}
		if organizationMFAEnforced {
			return nil, onekiterrors.New(http.StatusConflict, "MFA is required by your organization")
		}
	}

	if err := a.Service.SetTwoFactor(ctx, user.UserID, req.GetEnabled()); err != nil {
		// Step-up is evaluated first by the service, so its refusal is the more
		// specific answer; the org-MFA conflict is the fallback.
		if guardErr := securityGuardError(err); guardErr != nil {
			return nil, guardErr
		}
		if application.IsMFAEnforcedByOrganization(err) {
			return nil, onekiterrors.New(http.StatusConflict, "MFA is required by your organization")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update 2FA settings")
	}

	msg := "Two-factor authentication disabled"
	if req.GetEnabled() {
		msg = "Two-factor authentication enabled"
	}

	return &hubsecurityv1.Update2FAResponse{Message: msg}, nil
}

func (a *SecurityAPI) UpdateOrganizationMFAEnforcement(ctx context.Context, req *hubsecurityv1.UpdateOrganizationMFAEnforcementRequest) (*hubsecurityv1.UpdateOrganizationMFAEnforcementResponse, error) {
	user, authErr := requireOrganizationSecurityPermission(ctx)
	if authErr != nil {
		return nil, authErr
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.SetOrganizationMFAEnforced(ctx, businessID, req.GetEnabled()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update organization MFA policy")
	}
	return &hubsecurityv1.UpdateOrganizationMFAEnforcementResponse{Enforced: req.GetEnabled()}, nil
}

func (a *SecurityAPI) BeginMFAEnrollment(ctx context.Context, _ *hubsecurityv1.BeginMFAEnrollmentRequest) (*hubsecurityv1.BeginMFAEnrollmentResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}

	secret, uri, err := a.Service.BeginMFAEnrollment(ctx, user.UserID)
	if err != nil {
		if guardErr := securityGuardError(err); guardErr != nil {
			return nil, guardErr
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to begin MFA enrollment")
	}
	return &hubsecurityv1.BeginMFAEnrollmentResponse{Secret: secret, OtpauthUri: uri}, nil
}

func (a *SecurityAPI) CompleteMFAEnrollment(ctx context.Context, req *hubsecurityv1.CompleteMFAEnrollmentRequest) (*hubsecurityv1.CompleteMFAEnrollmentResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}

	recoveryCodes, err := a.Service.CompleteMFAEnrollment(ctx, user.UserID, req.GetCode())
	if err != nil {
		if guardErr := securityGuardError(err); guardErr != nil {
			return nil, guardErr
		}
		if errors.Is(err, application.ErrMFAInvalidCode) || errors.Is(err, application.ErrMFAEnrollmentMissing) {
			return nil, onekiterrors.New(http.StatusBadRequest, "MFA code is invalid or enrollment has expired")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to complete MFA enrollment")
	}
	return &hubsecurityv1.CompleteMFAEnrollmentResponse{Verified: true, RecoveryCodes: recoveryCodes}, nil
}

func (a *SecurityAPI) VerifyMFA(ctx context.Context, req *hubsecurityv1.VerifyMFARequest) (*hubsecurityv1.VerifyMFAResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}

	if err := a.Service.VerifyMFA(ctx, user.UserID, req.GetCode()); err != nil {
		if errors.Is(err, application.ErrMFAAttemptsExceeded) {
			return nil, onekiterrors.New(http.StatusTooManyRequests, "Too many MFA code attempts, try again later")
		}
		if errors.Is(err, application.ErrMFAInvalidCode) {
			return nil, onekiterrors.New(http.StatusUnauthorized, "MFA code is invalid")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify MFA")
	}
	return &hubsecurityv1.VerifyMFAResponse{Verified: true}, nil
}

func (a *SecurityAPI) UseMFARecoveryCode(ctx context.Context, req *hubsecurityv1.UseMFARecoveryCodeRequest) (*hubsecurityv1.UseMFARecoveryCodeResponse, error) {
	user, authErr := requireAuthenticatedSecurityUser(ctx)
	if authErr != nil {
		return nil, authErr
	}

	if err := a.Service.UseMFARecoveryCode(ctx, user.UserID, req.GetCode()); err != nil {
		if errors.Is(err, application.ErrMFAAttemptsExceeded) {
			return nil, onekiterrors.New(http.StatusTooManyRequests, "Too many MFA code attempts, try again later")
		}
		if errors.Is(err, application.ErrMFAInvalidCode) {
			return nil, onekiterrors.New(http.StatusUnauthorized, "MFA recovery code is invalid")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify MFA recovery code")
	}
	return &hubsecurityv1.UseMFARecoveryCodeResponse{Verified: true}, nil
}
