package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
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
			logger.Ctx(ctx).Error("failed to enqueue welcome email", logger.Error(qErr))
			return 0, qErr
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

func (s *Service) IssueEmailVerificationCode(ctx context.Context, emailAddress string) error {
	code, err := s.repo.CreateEmailVerificationCode(ctx, emailAddress)
	if err != nil {
		return err
	}
	_, err = email.Send(ctx, email.SendParams{
		To:      []string{strings.ToLower(strings.TrimSpace(emailAddress))},
		Subject: "Your rtylr email verification code",
		HTML:    fmt.Sprintf("<p>Your rtylr email verification code is <strong>%s</strong>.</p><p>This code expires in 10 minutes and can only be used once.</p>", code),
		Text:    fmt.Sprintf("Your rtylr email verification code is %s. It expires in 10 minutes and can only be used once.", code),
	})
	return err
}

func (s *Service) VerifyEmailCode(ctx context.Context, code string) error {
	if err := s.repo.ConsumeEmailVerificationCode(ctx, code); err != nil {
		return ErrInvalidEmailCode
	}
	return nil
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

	// The owner's authorization row. Without it GetAuthorizationSnapshot finds
	// nothing for the person who just created the business, EffectiveAccess.IsOwner
	// stays false, and every requireHubPermission check answers 403 to them.
	//
	// Logged rather than fatal, matching the employee record above: the business
	// and its owner exist and are usable through the legacy path, and failing
	// creation at this point would leave a half-made business behind.
	if err := s.repo.EnsureOwnerAuthorizationMembership(ctx, businessID, userID); err != nil {
		logger.Ctx(ctx).Error("failed to write the owner authorization membership",
			logger.Int("business_id", int(businessID)), logger.Int("user_id", int(userID)), logger.Error(err))
	}

	if err := s.repo.EnsureEmployeeForUser(ctx, userID, businessID, userID); err != nil {
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

	if claims.BusinessID > 0 {
		// Re-read the role on every validation instead of trusting the JWT:
		// a demoted member must not keep "admin" until the token expires.
		if role, err := s.repo.GetUserRole(ctx, claims.BusinessID, claims.UserID); err == nil {
			claims.Role = role
		}
	}

	// console_admin is additive while the flag is set, but a stale claim from
	// before a demotion must be cleared rather than carried forward.
	if err == nil && user != nil {
		if user.IsConsoleAdmin {
			claims.Role = "console_admin"
		} else if claims.Role == "console_admin" {
			claims.Role = ""
		}
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

// BusinessNameExists reports whether a business with the given name already
// exists. Business names are globally unique; this drives onboarding's inline
// availability check.
func (s *Service) BusinessNameExists(ctx context.Context, name string) (bool, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return false, nil
	}
	return s.repo.BusinessNameExists(ctx, trimmed)
}

// OnboardCreateBusiness creates the authenticated user's first business during
// onboarding, sets it as their current business, and returns a fresh JWT that
// carries the new business_id. Enforces the global business-name uniqueness.
func (s *Service) OnboardCreateBusiness(ctx context.Context, userID int64, name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrBusinessNameRequired
	}

	exists, err := s.repo.BusinessNameExists(ctx, trimmed)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrBusinessNameTaken
	}

	if err := s.CreateBusinessForUser(ctx, userID, trimmed); err != nil {
		return "", err
	}

	businessID, err := s.repo.GetUserBusinessID(ctx, userID)
	if err != nil {
		return "", err
	}

	claims := &AccessTokenClaims{UserID: userID, BusinessID: businessID}
	s.hydrateAccessTokenClaims(ctx, claims)
	return s.JWT.CreateJWT(claims)
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

func (s *Service) getUserBusinessID(ctx context.Context, userID int64) (int64, error) {
	return cacher.Of[int64]().GetOrComputeSingle(ctx,
		cacheKeys.UserBusinessID(userID),
		func(ctx context.Context) (int64, error) {
			return s.repo.GetUserBusinessID(ctx, userID)
		},
		cacher.WithExpiry(ctx, userProfileTTL),
	)
}

func (s *Service) getUserBusinessRole(ctx context.Context, businessID, userID int64) (string, error) {
	return cacher.Of[string]().GetOrComputeSingle(ctx,
		cacheKeys.UserBusinessRole(userID, businessID),
		func(ctx context.Context) (string, error) {
			return s.repo.GetUserRole(ctx, businessID, userID)
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
	TokenEpoch     int64  `json:"token_epoch,omitempty"`
	jwt.RegisteredClaims
}

// UnmarshalJSON accepts both the current numeric claim representation and the
// string-encoded representation issued by older platform clients. The JWT
// signature and standard claims are still validated by jwt.ParseWithClaims;
// this only preserves compatibility for the integer claim wire format.
func (c *AccessTokenClaims) UnmarshalJSON(data []byte) error {
	type alias AccessTokenClaims
	aux := struct {
		*alias
		UserID         json.RawMessage `json:"user_id"`
		BusinessID     json.RawMessage `json:"business_id"`
		PlatformUserID json.RawMessage `json:"platform_user_id"`
		PlanID         json.RawMessage `json:"plan_id"`
	}{alias: (*alias)(c)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	if len(aux.UserID) > 0 {
		if c.UserID, err = parseAccessTokenInt64(aux.UserID, "user_id"); err != nil {
			return err
		}
	}
	if len(aux.BusinessID) > 0 {
		if c.BusinessID, err = parseAccessTokenInt64(aux.BusinessID, "business_id"); err != nil {
			return err
		}
	}
	if len(aux.PlatformUserID) > 0 {
		if c.PlatformUserID, err = parseAccessTokenInt64(aux.PlatformUserID, "platform_user_id"); err != nil {
			return err
		}
	}
	if len(aux.PlanID) > 0 {
		if c.PlanID, err = parseAccessTokenInt64(aux.PlanID, "plan_id"); err != nil {
			return err
		}
	}

	return nil
}

func parseAccessTokenInt64(raw json.RawMessage, field string) (int64, error) {
	if string(raw) == "null" {
		return 0, nil
	}

	var numeric int64
	if err := json.Unmarshal(raw, &numeric); err == nil {
		return numeric, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, fmt.Errorf("%s must be an integer or numeric string: %w", field, err)
	}
	parsed, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer or numeric string: %w", field, err)
	}
	return parsed, nil
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
func (s *Service) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	return s.repo.RecordLogin(ctx, userID, email, method, ip, userAgent, success)
}

func (s *Service) AssignSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	return s.repo.AssignSubscription(ctx, ownerID, planSlug)
}

// SwitchBusiness validates the user's access to businessID and issues
// a new JWT scoped to that business.
func (s *Service) SwitchBusiness(ctx context.Context, userID, businessID int64) (string, error) {
	hasAccess, err := s.repo.CheckBusinessAccess(ctx, userID, businessID)
	if err != nil {
		return "", err
	}
	if !hasAccess {
		return "", ErrBusinessAccessDenied
	}

	claims := &AccessTokenClaims{
		UserID:     userID,
		BusinessID: businessID,
	}
	s.hydrateAccessTokenClaims(ctx, claims)
	return s.JWT.CreateJWT(claims)
}
