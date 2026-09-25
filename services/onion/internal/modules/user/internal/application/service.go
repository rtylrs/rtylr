package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/email"
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
	ErrInvalidCredentials   = infrastructure.ErrInvalidCredentials
	ErrUserAlreadyExists    = infrastructure.ErrUserAlreadyExists
	ErrWeakPassword         = infrastructure.ErrWeakPassword
	ErrInvalidEmail         = infrastructure.ErrInvalidEmail
	ErrBusinessAccessDenied = errors.New("business access denied")
	ErrBusinessNameRequired = errors.New("business name is required")
	ErrBusinessNameTaken    = errors.New("business name already taken")
	ErrInvalidEmailCode     = errors.New("invalid or expired email verification code")
)

type Service struct {
	repo                *infrastructure.Repository
	db                  *sql.DB
	validator           *Validator
	JWT                 *jwtutils.JWTHelper
	googleClientID      string
	workOSProvider      *WorkOSProvider
	businessProvisioner BusinessProvisioner
}

type BusinessProvisioner interface {
	ProvisionBusiness(ctx context.Context, businessID, ownerUserID int64, companyName string) error
}

func NewService(_ context.Context, db *sql.DB, googleClientID string) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo:           repo,
		db:             db,
		validator:      NewValidator(),
		googleClientID: googleClientID,
	}
}

// WithWorkOSProvider attaches a WorkOS provider to the service.
// Call this after NewService when WorkOS credentials are available.
func (s *Service) WithWorkOSProvider(p *WorkOSProvider) *Service {
	s.workOSProvider = p
	return s
}

func (s *Service) WithBusinessProvisioner(provisioner BusinessProvisioner) *Service {
	s.businessProvisioner = provisioner
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

	if !errors.Is(err, sql.ErrNoRows) {
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

	if businessName != "" {
		if err := s.CreateBusinessForUser(ctx, userID, businessName); err != nil {
			return 0, fmt.Errorf("create registration business: %w", err)
		}
	}

	if s.db != nil {
		greeting := normalizedFullName
		if greeting == "" {
			greeting = "there"
		}
		if qErr := email.EnqueueTemplate(ctx, s.db, email.PriorityNormal, normalizedEmail, "Welcome to rtylr", email.TemplateWelcome,
			email.WelcomeData{Subject: "Welcome to rtylr", Name: greeting},
			email.WithFrom("Mohamad at rtylr <mohamad@voxire.com>"),
			email.WithReplyTo("mohamad@voxire.com"),
		); qErr != nil {
			logger.Ctx(ctx).Error("failed to enqueue welcome email", logger.Error(qErr))
			return 0, qErr
		}
	}

	return userID, nil
}
func (s *Service) Authenticate(ctx context.Context, identifier, password string) (int64, error) {
	user, err := s.repo.GetUserByIdentifier(ctx, strings.ToLower(strings.TrimSpace(identifier)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
