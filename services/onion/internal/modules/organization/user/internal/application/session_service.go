package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
)

const refreshSessionTTL = 30 * 24 * time.Hour

func sessionTokenHash(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func (s *Service) CreateSession(ctx context.Context, userID int64, method, ip, userAgent string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	if err := s.repo.CreateSession(ctx, userID, sessionTokenHash(token), ip, userAgent, method, time.Now().Add(refreshSessionTTL)); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) RefreshSession(ctx context.Context, token string, resolve func(context.Context, int64) *PlatformClaims) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ErrInvalidCredentials
	}
	userID, err := s.repo.GetSessionUserID(ctx, sessionTokenHash(token))
	if errors.Is(err, domain.ErrSessionNotFound) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}
	if userID <= 0 {
		return "", ErrInvalidCredentials
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if errors.Is(err, domain.ErrUserNotFound) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}
	active, err := s.accountHasActiveAccess(ctx, user)
	if err != nil {
		return "", err
	}
	if !active {
		return "", ErrInvalidCredentials
	}
	var pc *PlatformClaims
	if resolve != nil {
		pc = resolve(ctx, userID)
	}
	return s.GenerateTokenWithPlatform(ctx, userID, pc)
}

func (s *Service) RevokeSession(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	return s.repo.RevokeSession(ctx, sessionTokenHash(token))
}
