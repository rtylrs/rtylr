package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/netip"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
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

// GoogleAuth exchanges a Google auth code for user info, then either logs
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
	if s.googleOAuthProvider == nil || s.googleClientID == "" {
		return nil, ErrGoogleOAuthNotConfigured
	}

	googleUser, err := s.googleOAuthProvider.AuthenticateWithCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrGoogleOAuthNotConfigured) {
			return nil, ErrGoogleOAuthNotConfigured
		}
		logger.Ctx(ctx).Warn("google code exchange failed", logger.Error(err))
		if googleOAuthUnavailable(err) {
			return nil, ErrGoogleOAuthUnavailable
		}
		return nil, ErrGoogleTokenInvalid
	}
	return s.linkGoogleUser(ctx, googleUser, businessName, signupIP, deviceInfo)
}

// linkGoogleUser verifies the provider identity and links or provisions the
// local account. Split from GoogleAuth so the verification gate is unit-testable.
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

	// Only a genuine miss may provision a new account. Any other lookup failure
	// (connection reset, context deadline, lock wait timeout) used to fall
	// through to CreateUser for an email that already exists, hit the UNIQUE
	// email index, and surface a raw duplicate-key error instead of the lookup
	// failure; without that index it would create a second account.
	if findErr != nil && !errors.Is(findErr, domain.ErrUserNotFound) {
		return nil, findErr
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
		SignupDeviceInfo: deviceInfo,
	}
	// netip.Addr{}.String() is the literal "invalid IP"; leave the column NULL
	// when the client address could not be parsed rather than storing that.
	if signupIP.IsValid() {
		newUser.SignupIPAddress = signupIP.String()
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
		} else if empErr := s.ensureEmployeeForUser(ctx, userID, bizID, userID, newUser.FullName, newUser.Email); empErr != nil {
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

// GoogleAuthorizationURL returns the Google-hosted consent URL.
func (s *Service) GoogleAuthorizationURL() (string, error) {
	if s.googleOAuthProvider == nil || s.googleClientID == "" {
		return "", ErrGoogleOAuthNotConfigured
	}
	return s.googleOAuthProvider.GetAuthorizationURL()
}
