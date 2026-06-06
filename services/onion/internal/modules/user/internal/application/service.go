package application

import (
	"context"
	"database/sql"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/modules/user/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const userProfileTTL = 5 * time.Minute

func invalidateUserScope(ctx context.Context, userID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(userID))
	}
}

var (
	ErrInvalidCredentials = infrastructure.ErrInvalidCredentials
	ErrUserAlreadyExists  = infrastructure.ErrUserAlreadyExists
	ErrWeakPassword       = infrastructure.ErrWeakPassword
	ErrInvalidEmail       = infrastructure.ErrInvalidEmail
)

type Service struct {
	repo           *infrastructure.Repository
	db             *sql.DB
	validator      *Validator
	JWT            *jwtutils.JWTHelper
	googleClientID string
	workOSProvider *WorkOSProvider
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
			logger.Ctx(ctx).Warn("failed to enqueue welcome email", logger.Error(qErr))
		}
	}

	return userID, nil
}

func (s *Service) RegisterLegacy(ctx context.Context, email, password, fullName string) (int64, error) {
	return s.Register(ctx, RegisterUserData{
		Email:    email,
		Password: password,
		FullName: fullName,
	})
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (int64, error) {
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}

		return 0, err
	}

	if !passwordutils.CheckPasswordHash(password, user.PasswordHash) {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}

// CreateBusinessForUser creates a new business and assigns the user as owner.
func (s *Service) CreateBusinessForUser(ctx context.Context, userID int64, businessName string) error {
	businessID, err := s.repo.CreateBusiness(ctx, businessName, userID)
	if err != nil {
		return err
	}

	if err := s.repo.SetUserBusinessID(ctx, userID, businessID); err != nil {
		return err
	}

	if err := s.repo.EnsureEmployeeForUser(ctx, userID, userID); err != nil {
		logger.Ctx(ctx).Error("failed to create employee record for business owner",
			logger.Int("owner_id", int(userID)), logger.Error(err))
	}

	return nil
}

type PlatformClaims struct {
	Platform       string
	PlatformUserID int64
	PlanID         int64
}

func (s *Service) GenerateToken(ctx context.Context, userID int64) (string, error) {
	_ = ctx

	return s.JWT.CreateJWT(&AccessTokenClaims{
		UserID: userID,
	})
}

func (s *Service) GenerateTokenWithPlatform(ctx context.Context, userID int64, pc *PlatformClaims) (string, error) {
	claims := &AccessTokenClaims{
		UserID: userID,
	}

	// Look up business_id and role
	businessID, _ := s.repo.GetUserBusinessID(ctx, userID)
	if businessID > 0 {
		claims.BusinessID = businessID
		role, _ := s.repo.GetUserRole(ctx, businessID, userID)
		claims.Role = role
	}

	if pc != nil {
		claims.Platform = pc.Platform
		claims.PlatformUserID = pc.PlatformUserID
		claims.PlanID = pc.PlanID
	}

	return s.JWT.CreateJWT(claims)
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	err := s.JWT.ValidateToken(ctx, tokenString, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (s *Service) RefreshToken(ctx context.Context, userID int64) (string, error) {
	_ = ctx

	return s.JWT.RefreshToken(&AccessTokenClaims{
		UserID: userID,
	})
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

func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	defer invalidateUserScope(ctx, userID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !passwordutils.CheckPasswordHash(currentPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	err = s.validator.ValidatePassword(newPassword)
	if err != nil {
		return err
	}

	hash, err := passwordutils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdateUserPassword(ctx, userID, hash)
}

func (s *Service) DeleteAccount(ctx context.Context, userID int64, password string) error {
	defer invalidateUserScope(ctx, userID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !passwordutils.CheckPasswordHash(password, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	return s.repo.DeleteUser(ctx, userID)
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// EmailExists reports whether an account with the given email already exists.
// The email is normalized (trimmed + lowercased) to match how accounts are
// stored, so the check is case-insensitive.
func (s *Service) EmailExists(ctx context.Context, email string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return false, nil
	}
	return s.repo.EmailExists(ctx, normalized)
}

func (s *Service) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return cacher.Of[*domain.User]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped("user", id, "profile"),
		func(ctx context.Context) (*domain.User, error) {
			return s.repo.GetUserByID(ctx, id)
		},
		cacher.WithExpiry(ctx, userProfileTTL),
	)
}

type AccessTokenClaims struct {
	UserID         int64  `json:"user_id"`
	BusinessID     int64  `json:"business_id,omitempty"`
	Role           string `json:"role,omitempty"`
	Platform       string `json:"platform,omitempty"`
	PlatformUserID int64  `json:"platform_user_id,omitempty"`
	PlanID         int64  `json:"plan_id,omitempty"`
	jwt.RegisteredClaims
}

func (c *AccessTokenClaims) SetIssuer(issuer string) {
	c.Issuer = issuer
}

func (c *AccessTokenClaims) SetExpiration(exp time.Time) {
	if !exp.IsZero() {
		c.ExpiresAt = jwt.NewNumericDate(exp)
	}
}

// RecordLogin delegates to the repository to write a login history entry.
func (s *Service) RecordLogin(ctx context.Context, userID int64, email, method, ip string, success bool) error {
	return s.repo.RecordLogin(ctx, userID, email, method, ip, success)
}

func (s *Service) AssignSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	return s.repo.AssignSubscription(ctx, ownerID, planSlug)
}
