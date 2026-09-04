package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/netip"
	"strings"

	"github.com/workos/workos-go/v4/pkg/workos_errors"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

// GoogleAuthResult is what the handler returns to the frontend. IsNew
// lets callers run first-signup hooks (welcome email, onboarding
// flag) without a round trip.
type GoogleAuthResult struct {
	UserID int64
	Token  string
	IsNew  bool
}

// Errors surfaced by GoogleAuth. Handler translates into HTTP codes.
var (
	ErrGoogleOAuthNotConfigured = errors.New("google oauth not configured")
	ErrGoogleTokenInvalid       = errors.New("google token invalid")
	ErrGoogleOAuthUnavailable   = errors.New("google oauth provider unavailable")
	ErrGoogleEmailUnverified    = errors.New("google email unverified")
)

// GoogleAuth exchanges a WorkOS auth code for user info, then either logs
// in the matching user (by verified email) or provisions a new account.
//
// Account linking rule (locked in the plan):
//   - a local user exists with that email,
//   - THEN issue a token to that existing user, flip auth_provider
//     to 'google' if it was previously 'password'.
//   - Otherwise create a new user with auth_provider='google' and a
//     random unusable password hash.
//
// The `signup_ip_address` and `device_info` mirror the register flow
// so we can forensically match signups across auth types.
func (s *Service) GoogleAuth(
	ctx context.Context,
	code string,
	businessName string,
	signupIP netip.Addr,
	deviceInfo string,
) (*GoogleAuthResult, error) {
	if s.workOSProvider == nil || s.workOSProvider.clientID == "" {
		return nil, ErrGoogleOAuthNotConfigured
	}

	googleUser, err := s.workOSProvider.AuthenticateWithCode(ctx, code)
	if err != nil {
		logger.Ctx(ctx).Warn("workos code exchange failed", logger.Error(err))
		if isWorkOSUnavailable(err) {
			return nil, ErrGoogleOAuthUnavailable
		}
		return nil, ErrGoogleTokenInvalid
	}
	return s.linkGoogleUser(ctx, googleUser, businessName, signupIP, deviceInfo)
}

func isWorkOSUnavailable(err error) bool {
	status, isHTTPError := workOSHTTPStatus(err)
	if isHTTPError {
		return status == 429 || status >= 500
	}
	return true
}

func workOSHTTPStatus(err error) (int, bool) {
	var httpErr workos_errors.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Code, true
	}

	// WorkOS uses structured authentication errors for some 4xx responses;
	// each embeds the HTTP status but has its own Error implementation.
	var emailVerification *workos_errors.EmailVerificationRequiredError
	if errors.As(err, &emailVerification) {
		return emailVerification.HTTPError.Code, true
	}
	var mfaEnrollment *workos_errors.MFAEnrollmentError
	if errors.As(err, &mfaEnrollment) {
		return mfaEnrollment.HTTPError.Code, true
	}
	var mfaChallenge *workos_errors.MFAChallengeError
	if errors.As(err, &mfaChallenge) {
		return mfaChallenge.HTTPError.Code, true
	}
	var organizationSelection *workos_errors.OrganizationSelectionRequiredError
	if errors.As(err, &organizationSelection) {
		return organizationSelection.HTTPError.Code, true
	}
	var ssoRequired *workos_errors.SSORequiredError
	if errors.As(err, &ssoRequired) {
		return ssoRequired.HTTPError.Code, true
	}
	var organizationMethods *workos_errors.OrganizationAuthenticationMethodsRequiredError
	if errors.As(err, &organizationMethods) {
		return organizationMethods.HTTPError.Code, true
	}

	return 0, false
}

// linkGoogleUser verifies the provider identity and links or provisions the
// local account. Split from GoogleAuth so the verification gate is unit-
// testable without the WorkOS SDK (#815).
func (s *Service) linkGoogleUser(
	ctx context.Context,
	googleUser *GoogleUser,
	businessName string,
	signupIP netip.Addr,
	deviceInfo string,
) (*GoogleAuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(googleUser.Email))
	if email == "" {
		return nil, ErrGoogleTokenInvalid
	}
	// Never link or provision an account off an unverified email — a
	// provider that allows unverified addresses would otherwise let an
	// attacker take over the matching local account.
	if !googleUser.EmailVerified {
		return nil, ErrGoogleEmailUnverified
	}

	// Try to link by email first.
	existing, findErr := s.repo.GetUserByEmail(ctx, email)
	if findErr == nil && existing != nil && existing.ID > 0 {
		token, tokenErr := s.GenerateToken(ctx, existing.ID)
		if tokenErr != nil {
			return nil, tokenErr
		}
		return &GoogleAuthResult{UserID: existing.ID, Token: token, IsNew: false}, nil
	}

	// Create a new user. Password hash is a random 64 hex chars --
	// bcrypt of that string, never handed out. User can set a real
	// password later via the "forgot password" flow.
	randomFiller := make([]byte, 32)
	if _, rErr := rand.Read(randomFiller); rErr != nil {
		return nil, rErr
	}
	pwHash, hashErr := passwordutils.HashPassword(hex.EncodeToString(randomFiller))
	if hashErr != nil {
		return nil, hashErr
	}

	newUser := &domain.User{
		Email:            email,
		PasswordHash:     pwHash,
		FullName:         googleUser.Name,
		SignupIPAddress:  signupIP.String(),
		SignupDeviceInfo: deviceInfo,
	}

	userID, createErr := s.repo.CreateUser(ctx, newUser)
	if createErr != nil {
		return nil, createErr
	}

	// Optional: create a business if the signup flow collected a name.
	// Mirrors the existing Register flow -- failure here is logged but
	// non-fatal; user can complete business setup from the dashboard.
	businessName = strings.TrimSpace(businessName)
	if businessName != "" {
		bizID, bizErr := s.repo.CreateBusiness(ctx, businessName, userID)
		if bizErr != nil {
			logger.Ctx(ctx).Warn("failed to create business during google signup",
				logger.Error(bizErr),
				logger.Any("user_id", userID),
			)
		} else if setErr := s.repo.SetUserBusinessID(ctx, userID, bizID); setErr != nil {
			logger.Ctx(ctx).Warn("failed to link business to google-signup user",
				logger.Error(setErr),
				logger.Any("user_id", userID),
			)
		} else if authErr := s.repo.EnsureOwnerAuthorizationMembership(ctx, bizID, userID); authErr != nil {
			logger.Ctx(ctx).Warn("failed to write the owner authorization membership for google signup",
				logger.Error(authErr),
				logger.Any("user_id", userID),
			)
		} else if empErr := s.repo.EnsureEmployeeForUser(ctx, userID, bizID, userID); empErr != nil {
			logger.Ctx(ctx).Warn("failed to create employee record for google-signup owner",
				logger.Error(empErr),
				logger.Any("user_id", userID),
			)
		}
	}

	if subErr := s.AssignSubscription(ctx, userID, DefaultSignupPlanSlug); subErr != nil {
		logger.Ctx(ctx).Warn("failed to assign subscription during google signup",
			logger.Error(subErr),
			logger.Any("user_id", userID),
			logger.String("plan_slug", DefaultSignupPlanSlug),
		)
	}

	token, tokenErr := s.GenerateToken(ctx, userID)
	if tokenErr != nil {
		return nil, tokenErr
	}

	return &GoogleAuthResult{UserID: userID, Token: token, IsNew: true}, nil
}

// SocialAuthorizationURL returns the WorkOS-hosted consent URL for a social
// provider (google, github, microsoft, apple). The frontend redirects the user
// there; the provider returns to the callback with a code that GoogleAuth
// exchanges (the exchange is provider-agnostic).
func (s *Service) SocialAuthorizationURL(provider string) (string, error) {
	if s.workOSProvider == nil || s.workOSProvider.clientID == "" {
		return "", ErrGoogleOAuthNotConfigured
	}
	return s.workOSProvider.GetAuthorizationURL(provider)
}

// GoogleAuthorizationURL is the Google-specific shorthand, kept for callers
// that predate multi-provider support.
func (s *Service) GoogleAuthorizationURL() (string, error) {
	return s.SocialAuthorizationURL("google")
}
