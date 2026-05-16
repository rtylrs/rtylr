package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/netip"
	"os"
	"strings"

	"google.golang.org/api/idtoken"

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
	ErrGoogleEmailUnverified    = errors.New("google email unverified")
)

// tokenVerifier is the minimal surface the service needs; the real
// implementation is google.golang.org/api/idtoken.Validate but stubbing
// this in tests means we don't need network access or a real Google
// project during unit tests.
type tokenVerifier func(ctx context.Context, idToken, audience string) (*idtoken.Payload, error)

// Default to the real Google verifier. Test double can swap it.
var verifyGoogleToken tokenVerifier = idtoken.Validate

// GoogleAuth verifies a Google-issued ID token, then either logs in
// the matching user (by verified email) or provisions a new account.
//
// Account linking rule (locked in the plan):
//   - email_verified === true on the Google payload,
//   - AND a local user exists with that email,
//   - THEN issue a token to that existing user, flip auth_provider
//     to 'google' if it was previously 'password'.
//   - Otherwise create a new user with auth_provider='google' and a
//     random unusable password hash.
//
// The `signup_ip_address` and `device_info` mirror the register flow
// so we can forensically match signups across auth types.
func (s *Service) GoogleAuth(
	ctx context.Context,
	idTokenStr string,
	businessName string,
	signupIP netip.Addr,
	deviceInfo string,
) (*GoogleAuthResult, error) {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	if clientID == "" {
		return nil, ErrGoogleOAuthNotConfigured
	}

	payload, err := verifyGoogleToken(ctx, idTokenStr, clientID)
	if err != nil {
		logger.Ctx(ctx).Warn("google id token validation failed", logger.Error(err))
		return nil, ErrGoogleTokenInvalid
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	fullName, _ := payload.Claims["name"].(string)

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, ErrGoogleTokenInvalid
	}
	if !emailVerified {
		// We explicitly refuse to link unverified Google emails — it's
		// the one vector where someone could hijack an existing rtylr
		// account by claiming an email they don't own.
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

	// Create a new user. Password hash is a random 64 hex chars —
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
		FullName:         fullName,
		SignupIPAddress:  signupIP.String(),
		SignupDeviceInfo: deviceInfo,
	}

	userID, createErr := s.repo.CreateUser(ctx, newUser)
	if createErr != nil {
		return nil, createErr
	}

	// Optional: create a business if the signup flow collected a name.
	// Mirrors the existing Register flow — failure here is logged but
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
		}
	}

	token, tokenErr := s.GenerateToken(ctx, userID)
	if tokenErr != nil {
		return nil, tokenErr
	}

	return &GoogleAuthResult{UserID: userID, Token: token, IsNew: true}, nil
}
