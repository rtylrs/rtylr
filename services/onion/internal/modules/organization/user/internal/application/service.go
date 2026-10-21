package application

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	userProfileTTL        = 5 * time.Minute
	DefaultSignupPlanSlug = "free"
)

func invalidateUserScope(ctx context.Context, userID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.UserPrefix(userID))
}

var (
	ErrInvalidCredentials   = domain.ErrInvalidCredentials
	ErrUserAlreadyExists    = domain.ErrUserAlreadyExists
	ErrWeakPassword         = domain.ErrWeakPassword
	ErrInvalidEmail         = domain.ErrInvalidEmail
	ErrBusinessAccessDenied = errors.New("business access denied")
	ErrBusinessNameRequired = errors.New("business name is required")
	ErrBusinessNameTaken    = errors.New("business name already taken")
	ErrInvalidEmailCode     = domain.ErrInvalidEmailVerificationCode
	// ErrEmailCodeSuperseded is a code minted before codes were bound to an
	// account. Retrying it can never succeed, so the caller is told to request
	// a new one instead of being told the code was wrong.
	ErrEmailCodeSuperseded = domain.ErrSupersededEmailVerificationCode
	// ErrTooManyEmailCodeAttempts is the per-user attempt cap. Answering with
	// the generic invalid-code error would leave a legitimate user with a
	// correct code and no way forward.
	ErrTooManyEmailCodeAttempts = domain.ErrTooManyEmailVerificationAttempts
)

type Service struct {
	repo                repository
	emails              emailNotifications
	validator           *Validator
	JWT                 *jwtutils.JWTHelper
	OAuthJWT            *jwtutils.JWTHelper
	googleClientID      string
	googleOAuthProvider *GoogleOAuthProvider
	businessProvisioner BusinessProvisioner
	employeeProvisioner EmployeeProvisioner
}

type BusinessProvisioner interface {
	ProvisionBusiness(ctx context.Context, businessID, ownerUserID int64, companyName string) error
}

type EmployeeProvisioner interface {
	EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, firstName, lastName, email string) error
}

func NewService(repo repository, googleClientID string) *Service {
	return &Service{
		repo:           repo,
		validator:      NewValidator(),
		googleClientID: googleClientID,
	}
}

func (s *Service) WithEmailNotifications(emails emailNotifications) *Service {
	s.emails = emails
	return s
}

// WithGoogleOAuthProvider attaches the configured Google OAuth client.
func (s *Service) WithGoogleOAuthProvider(p *GoogleOAuthProvider) *Service {
	s.googleOAuthProvider = p
	return s
}

func (s *Service) WithBusinessProvisioner(provisioner BusinessProvisioner) *Service {
	s.businessProvisioner = provisioner
	return s
}

func (s *Service) WithEmployeeProvisioner(provisioner EmployeeProvisioner) *Service {
	s.employeeProvisioner = provisioner
	return s
}

func (s *Service) Initialize(ctx context.Context) bool {
	jwtSecret, ok := ctx.Value(jwtutils.JWTSecretKey).(string)
	if !ok || jwtSecret == "" {
		logger.Ctx(ctx).Error("jwt_secret not found or not a string in context")

		return false
	}

	jwtDuration, ok := ctx.Value(jwtutils.JWTDurationKey).(time.Duration)
	if !ok {
		logger.Ctx(ctx).Error("jwt_duration not found or not a time.Duration in context")

		return false
	}

	s.JWT = jwtutils.NewJWTHelperWithDuration(jwtSecret, "onion-platform", jwtDuration)
	s.OAuthJWT = jwtutils.NewJWTHelperWithDuration(jwtSecret, OAuthTokenIssuer, jwtDuration)

	return true
}

func (s *Service) ValidateEmail(email string) error {
	return s.validator.ValidateEmail(email)
}

func (s *Service) ValidatePassword(password string) error {
	return s.validator.ValidatePassword(password)
}

type RegisterUserData struct {
	Email           string
	Password        string
	FullName        string
	BusinessName    string
	SignupIPAddress netip.Addr
	DeviceInfo      string
}

func (s *Service) Register(ctx context.Context, data RegisterUserData) (int64, error) {
	if err := s.validator.ValidateUserData(data); err != nil {
		return 0, err
	}

	normalizedEmail := s.validator.NormalizeEmail(data.Email)
	normalizedFullName := s.validator.NormalizeFullName(data.FullName)

	_, err := s.repo.GetUserByEmail(ctx, normalizedEmail)
	if err == nil {
		return 0, ErrUserAlreadyExists
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return 0, err
	}

	// businesses.name is globally UNIQUE, so check the requested name before the
	// user row exists. Failing afterwards left an account behind whose password
	// the user chose while the form reported an error, and retrying the same
	// email then hit ErrUserAlreadyExists forever. OnboardCreateBusiness
	// pre-checks the same way.
	businessName := strings.TrimSpace(data.BusinessName)
	if businessName != "" {
		exists, nameErr := s.repo.BusinessNameExists(ctx, businessName)
		if nameErr != nil {
			return 0, nameErr
		}
		if exists {
			return 0, ErrBusinessNameTaken
		}
	}

	hash, err := passwordutils.HashPassword(data.Password)
	if err != nil {
		return 0, err
	}

	user := domain.NewUser(
		normalizedEmail,
		hash,
		normalizedFullName,
		data.SignupIPAddress.String(),
		data.DeviceInfo,
	)

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	var businessID int64
	if businessName != "" {
		if err := s.CreateBusinessForUser(ctx, userID, businessName); err != nil {
			return 0, fmt.Errorf("create registration business: %w", err)
		}
		if businessID, err = s.repo.GetUserBusinessID(ctx, userID); err != nil {
			return 0, fmt.Errorf("read registration business: %w", err)
		}
	}

	if err := s.repo.CreateSignupProvisioning(ctx, userID, businessID, false); err != nil {
		return 0, fmt.Errorf("create signup provisioning: %w", err)
	}

	if s.emails != nil {
		greeting := normalizedFullName
		if greeting == "" {
			greeting = "there"
		}
		if qErr := s.emails.QueueWelcome(ctx, normalizedEmail, greeting); qErr != nil {
			logger.Ctx(ctx).Error("failed to enqueue welcome email", logger.Error(qErr))
			return 0, qErr
		}
	}

	// The account stays on the verification screen until this code is used.
	// A failed send is not fatal: that screen can send another.
	if err := s.IssueEmailVerificationCode(ctx, normalizedEmail); err != nil {
		logger.Ctx(ctx).Error("failed to send signup verification email", logger.Error(err))
	}

	return userID, nil
}

func (s *Service) Authenticate(ctx context.Context, identifier, password string) (int64, error) {
	user, err := s.repo.GetUserByIdentifier(ctx, strings.ToLower(strings.TrimSpace(identifier)))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Burn the same bcrypt time as the hash-compare branch below so
			// response latency cannot reveal whether the account exists.
			passwordutils.CheckDummyHash(password)

			return 0, ErrInvalidCredentials
		}

		return 0, err
	}

	if !passwordutils.CheckPasswordHash(password, user.PasswordHash) {
		return 0, ErrInvalidCredentials
	}
	if !s.hasActiveAccountAccess(ctx, user) {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}
