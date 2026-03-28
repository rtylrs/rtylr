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
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

var (
	ErrInvalidCredentials = infrastructure.ErrInvalidCredentials
	ErrUserAlreadyExists  = infrastructure.ErrUserAlreadyExists
	ErrWeakPassword       = infrastructure.ErrWeakPassword
	ErrInvalidEmail       = infrastructure.ErrInvalidEmail
)

type Service struct {
	repo      *infrastructure.Repository
	validator *Validator
	JWT       *jwtutils.JWTHelper
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo:      repo,
		validator: NewValidator(),
	}
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

	return s.repo.CreateUser(ctx, user)
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
	_ = ctx

	claims := &AccessTokenClaims{
		UserID: userID,
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

func (s *Service) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

type AccessTokenClaims struct {
	UserID         int64  `json:"user_id"`
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
