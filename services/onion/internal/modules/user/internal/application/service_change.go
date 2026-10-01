package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

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

type AccessTokenClaims struct {
	UserID         int64  `json:"user_id"`
	BusinessID     int64  `json:"business_id,omitempty"`
	IsConsoleAdmin bool   `json:"is_console_admin,omitempty"`
	Platform       string `json:"platform,omitempty"`
	PlatformUserID int64  `json:"platform_user_id,omitempty"`
	PlanID         int64  `json:"plan_id,omitempty"`
	TokenEpoch     int64  `json:"token_epoch,omitempty"`
	jwt.RegisteredClaims
}

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
	for _, field := range []struct {
		name string
		raw  json.RawMessage
		dest *int64
	}{
		{name: "user_id", raw: aux.UserID, dest: &c.UserID},
		{name: "business_id", raw: aux.BusinessID, dest: &c.BusinessID},
		{name: "platform_user_id", raw: aux.PlatformUserID, dest: &c.PlatformUserID},
		{name: "plan_id", raw: aux.PlanID, dest: &c.PlanID},
	} {
		if len(field.raw) == 0 || string(field.raw) == "null" {
			continue
		}
		value, err := parseAccessTokenInt64(field.raw, field.name)
		if err != nil {
			return err
		}
		*field.dest = value
	}
	return nil
}

func parseAccessTokenInt64(raw json.RawMessage, field string) (int64, error) {
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
