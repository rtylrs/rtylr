package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

var (
	ErrResetTokenNotFound = errors.New("password reset token not found or expired")
	ErrResetTokenInvalid  = errors.New("password reset token invalid")
)

const resetTokenTTL = time.Hour

func resetCacheKey(token string) string {
	return "pwd_reset:" + token
}

// RequestPasswordReset generates a 1-hour reset token, stores it in Redis,
// and queues a reset link. Unknown addresses still return nil; queue failures
// are returned so the caller can record the delivery failure.
func (s *Service) RequestPasswordReset(ctx context.Context, emailAddr string) error {
	emailAddr = strings.ToLower(strings.TrimSpace(emailAddr))

	user, err := s.repo.GetUserByEmail(ctx, emailAddr)
	if err != nil || user == nil || user.ID == 0 {
		return nil
	}

	raw := make([]byte, 32)
	if _, randErr := rand.Read(raw); randErr != nil {
		return randErr
	}
	token := hex.EncodeToString(raw)

	cache := cacher.GetRedisCache()
	if cache == nil {
		return errors.New("cache unavailable for password reset")
	}
	cache.Put(ctx, resetCacheKey(token), []byte(fmt.Sprintf("%d", user.ID)), resetTokenTTL)

	authAppURL := "https://auth.rtylr.com"
	if config.GlobalConfig != nil && config.GlobalConfig.AuthAppURL != "" {
		authAppURL = config.GlobalConfig.AuthAppURL
	}
	resetLink := fmt.Sprintf("%s/password-reset/confirm?token=%s", authAppURL, token)
	if s.db != nil {
		if emailErr := email.EnqueueTemplate(ctx, s.db, email.PriorityCritical, emailAddr, "Reset your rtylr password", email.TemplatePasswordReset, email.PasswordResetData{
			Subject:   "Reset your rtylr password",
			ResetLink: resetLink,
		}); emailErr != nil {
			logger.Ctx(ctx).Error("failed to enqueue password reset email", logger.Error(emailErr))
			return emailErr
		}
	}
	return nil
}

// ResetPassword validates a reset token and updates the user's password.
// The token is consumed on first use.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	cache := cacher.GetRedisCache()
	if cache == nil {
		return ErrResetTokenNotFound
	}

	val := cache.Get(ctx, resetCacheKey(token))
	if len(val) == 0 {
		return ErrResetTokenNotFound
	}

	var userID int64
	if _, scanErr := fmt.Sscanf(string(val), "%d", &userID); scanErr != nil || userID <= 0 {
		return ErrResetTokenInvalid
	}

	if err := s.validator.ValidatePassword(newPassword); err != nil {
		return err
	}

	pwHash, hashErr := passwordutils.HashPassword(newPassword)
	if hashErr != nil {
		return hashErr
	}

	if err := s.repo.UpdateUserPassword(ctx, userID, pwHash); err != nil {
		return err
	}

	cache.Delete(ctx, []string{resetCacheKey(token)})
	return nil
}
