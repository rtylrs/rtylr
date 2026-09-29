package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

// resetTokenStore is the whole surface the reset flow is allowed to use.
// GetDel is the only read that also spends the token, so a redemption cannot be
// expressed as Get followed by Delete. cacher.RedisCache satisfies it.
type resetTokenStore interface {
	Put(ctx context.Context, key string, val []byte, ttl time.Duration)
	GetDel(ctx context.Context, key string) []byte
}

// resetTokenCache is a variable so the store decision stays injectable for a
// regression test without a live Redis, mirroring how the websocket ticket
// store is tested. Production always resolves the process cache, per call, so a
// runtime Redis swap is still honoured.
var resetTokenCache = func() resetTokenStore { return cacher.GetRedisCache() }

// resetCacheKey hashes the token so a dump or scan of the cache keyspace never
// yields a usable reset credential.
func resetCacheKey(token string) string {
	digest := sha256.Sum256([]byte(token))

	return "pwd_reset:" + hex.EncodeToString(digest[:])
}

// consumeResetToken atomically reads and spends a reset token, returning the
// user it was issued for. GETDEL is the single gate: with a plain Get followed
// by a later Delete the token stayed valid for the whole window, so two
// concurrent requests carrying it both changed the password.
func consumeResetToken(ctx context.Context, store resetTokenStore, token string) (int64, error) {
	if store == nil {
		return 0, ErrResetTokenNotFound
	}
	if strings.TrimSpace(token) == "" {
		return 0, ErrResetTokenNotFound
	}

	val := store.GetDel(ctx, resetCacheKey(token))
	if len(val) == 0 {
		return 0, ErrResetTokenNotFound
	}

	var userID int64
	if _, scanErr := fmt.Sscanf(string(val), "%d", &userID); scanErr != nil || userID <= 0 {
		return 0, ErrResetTokenInvalid
	}

	return userID, nil
}

// newResetToken mints a 256-bit token from crypto/rand.
func newResetToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return hex.EncodeToString(raw), nil
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

	token, tokenErr := newResetToken()
	if tokenErr != nil {
		return tokenErr
	}

	cache := resetTokenCache()
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

// ResetPassword redeems a reset token and updates the user's password. The
// token is spent by a single atomic read-and-delete, so it can be redeemed
// exactly once no matter how many requests race for it.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	// The policy check runs before the token is spent, so a rejected password
	// does not burn the user's only link, and before any hashing, so an
	// unauthenticated caller cannot make the process burn bcrypt work.
	if err := s.validator.ValidatePassword(newPassword); err != nil {
		// Reported as the weak-password sentinel so callers map a policy
		// rejection to a 4xx; the validator's own error is not that sentinel.
		return fmt.Errorf("%w: %w", ErrWeakPassword, err)
	}

	userID, err := consumeResetToken(ctx, resetTokenCache(), token)
	if err != nil {
		return err
	}

	pwHash, hashErr := passwordutils.HashPassword(newPassword)
	if hashErr != nil {
		return hashErr
	}

	// The token is already spent. A failed write must not hand it back: giving
	// it back would restore the very replay window the atomic consume closed.
	return s.repo.UpdateUserPassword(ctx, userID, pwHash)
}
