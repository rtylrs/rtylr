package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// IssueEmailVerificationCode mails a fresh code and its link form.
//
// The code is ten digits and the token carries the user id with it, so a client
// that only accepts six digits (the auth web app's typed-entry field does) can
// no longer submit what is in the email body. The link is therefore the primary
// call to action and the code is secondary copy: a user who follows the link
// never has to type. See the migration note on formatEmailVerificationToken.
func (s *Service) IssueEmailVerificationCode(ctx context.Context, emailAddress string) error {
	issued, err := s.repo.CreateEmailVerificationCode(ctx, emailAddress)
	if err != nil {
		return err
	}

	authAppURL := config.AuthAppURL
	if config.GlobalConfig != nil && config.GlobalConfig.AuthAppURL != "" {
		authAppURL = config.GlobalConfig.AuthAppURL
	}
	verifyLink := strings.TrimSuffix(authAppURL, "/") + "/verify?token=" + url.QueryEscape(issued.Token)

	if s.emails == nil {
		return errors.New("email notifications are not configured")
	}
	err = s.emails.SendVerification(ctx, strings.ToLower(strings.TrimSpace(emailAddress)), "Verify your rtylr email address", fmt.Sprintf(
		"<p>Confirm your email address to finish setting up your rtylr account.</p>"+
			"<p><a href=\"%s\">Verify my email</a></p>"+
			"<p>Or enter this code on the verification page: <strong>%s</strong></p>"+
			"<p>This expires in 10 minutes and can only be used once.</p>",
		verifyLink, issued.Code), fmt.Sprintf(
		"Confirm your email address to finish setting up your rtylr account.\n\nVerify my email: %s\n\nOr enter this code on the verification page: %s\n\nThis expires in 10 minutes and can only be used once.",
		verifyLink, issued.Code))
	return err
}

// VerifyEmailCode spends a verification token. The three failure modes are kept
// apart only where the message can help the legitimate holder: a code issued
// before the user binding existed needs a *new* code rather than a retry, and an
// exhausted attempt budget needs a new code too. Everything else collapses to
// one error so the anonymous endpoint cannot be used to probe which accounts
// have a pending code.
func (s *Service) VerifyEmailCode(ctx context.Context, token string) error {
	err := s.repo.ConsumeEmailVerificationCode(ctx, token)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrSupersededEmailVerificationCode):
		return ErrEmailCodeSuperseded
	case errors.Is(err, domain.ErrTooManyEmailVerificationAttempts):
		return ErrTooManyEmailCodeAttempts
	default:
		return ErrInvalidEmailCode
	}
}

// CreateBusinessForUser creates a new business and assigns the user as owner.
func (s *Service) CreateBusinessForUser(ctx context.Context, userID int64, businessName string) error {
	defer invalidateUserScope(ctx, userID)

	businessID, err := s.repo.CreateBusiness(ctx, businessName, userID)
	if err != nil {
		return err
	}

	if err := s.repo.SetUserBusinessID(ctx, userID, businessID); err != nil {
		return err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to load account details for HR employee provisioning",
			logger.Int("owner_id", int(userID)), logger.Error(err))
	} else if err := s.ensureEmployeeForUser(ctx, userID, businessID, userID, user.FullName, user.Email); err != nil {
		logger.Ctx(ctx).Error("failed to create employee record for business owner",
			logger.Int("owner_id", int(userID)), logger.Error(err))
	}

	if s.businessProvisioner != nil {
		if err := s.businessProvisioner.ProvisionBusiness(ctx, businessID, userID, businessName); err != nil {
			logger.Ctx(ctx).Error("failed to provision HR records for new business",
				logger.Int("business_id", int(businessID)), logger.Int("owner_id", int(userID)), logger.Error(err))
		}
	}

	return nil
}

func (s *Service) ensureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, fullName, email string) error {
	if s.employeeProvisioner == nil {
		return errors.New("HR employee provisioner is not configured")
	}
	name := strings.Fields(strings.TrimSpace(fullName))
	firstName, lastName := "", ""
	if len(name) > 0 {
		firstName = name[0]
	}
	if len(name) > 1 {
		lastName = strings.Join(name[1:], " ")
	}
	return s.employeeProvisioner.EnsureEmployeeForUser(ctx, ownerID, businessID, userID, firstName, lastName, email)
}

type PlatformClaims struct {
	Platform       string
	PlatformUserID int64
	PlanID         int64
}

func (s *Service) GenerateToken(ctx context.Context, userID int64) (string, error) {
	return s.JWT.CreateJWT(s.buildAccessTokenClaims(ctx, userID))
}

func (s *Service) GenerateTokenWithPlatform(ctx context.Context, userID int64, pc *PlatformClaims) (string, error) {
	claims := s.buildAccessTokenClaims(ctx, userID)

	if pc != nil {
		claims.Platform = pc.Platform
		claims.PlatformUserID = pc.PlatformUserID
		claims.PlanID = pc.PlanID
	}

	return s.JWT.CreateJWT(claims)
}

func (s *Service) buildAccessTokenClaims(ctx context.Context, userID int64) *AccessTokenClaims {
	claims := &AccessTokenClaims{
		UserID: userID,
	}
	s.hydrateAccessTokenClaims(ctx, claims)

	return claims
}

func (s *Service) hydrateAccessTokenClaims(ctx context.Context, claims *AccessTokenClaims) *domain.User {
	if claims == nil || claims.UserID <= 0 {
		return nil
	}

	// Authentication state must bypass the profile cache. Otherwise a ban or
	// deactivation performed by another process can leave a valid token alive
	// for the cache TTL.
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err == nil && user != nil {
		// The epoch is the revocation anchor: tokens whose embedded epoch is
		// behind the stored one were issued before a password change or
		// credential reset and must stop validating.
		claims.TokenEpoch = user.TokenEpoch
	}

	if claims.BusinessID <= 0 {
		if businessID, err := s.getUserBusinessID(ctx, claims.UserID); err == nil && businessID > 0 {
			claims.BusinessID = businessID
		}
	}

	if err == nil && user != nil {
		claims.IsConsoleAdmin = user.IsConsoleAdmin
	}

	return user
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	err := s.JWT.ValidateToken(ctx, tokenString, claims)
	if err != nil {
		return nil, err
	}

	// Capture the epoch exactly as embedded at issuance: hydration refreshes
	// claims from the database and would otherwise mask a stale token.
	tokenEpoch := claims.TokenEpoch

	user := s.hydrateAccessTokenClaims(ctx, claims)

	// Deleted or missing users never validate, closing the deleted-token hole
	// alongside epoch-based revocation.
	if user == nil || tokenEpoch != user.TokenEpoch || !s.hasActiveAccountAccess(ctx, user) {
		return nil, ErrInvalidCredentials
	}

	return claims, nil
}

func (s *Service) hasActiveAccountAccess(ctx context.Context, user *domain.User) bool {
	if user == nil || user.IsBanned {
		return false
	}
	if user.IsConsoleAdmin {
		return true
	}

	state, err := s.repo.GetUserAccessState(ctx, user.ID)
	if err != nil {
		// Fail closed when the database cannot confirm current access.
		return false
	}
	// Users without a permission row are allowed through onboarding. Once an
	// account has been provisioned into a business, at least one active access
	// path must remain before it can authenticate.
	return !state.HasPermissions || state.HasActiveAccess
}

func (s *Service) RefreshToken(ctx context.Context, userID int64) (string, error) {
	return s.JWT.RefreshToken(s.buildAccessTokenClaims(ctx, userID))
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, name string) (*domain.User, error) {
	defer invalidateUserScope(ctx, userID)
	if err := s.validator.ValidateFullName(name); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.FullName = s.validator.NormalizeFullName(name)

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.repo.GetUserByID(ctx, userID)
}
