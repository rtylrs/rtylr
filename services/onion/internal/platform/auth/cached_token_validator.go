package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var errNilTokenClaims = errors.New("token validator returned nil claims")

type cachedTokenClaims struct {
	claims    TokenClaims
	expiresAt time.Time
}

type cachingTokenValidator struct {
	inner TokenValidator
	ttl   time.Duration

	mu        sync.RWMutex
	entries   map[string]cachedTokenClaims
	lastPrune time.Time
	group     singleflight.Group
}

func NewCachingTokenValidator(inner TokenValidator, ttl time.Duration) TokenValidator {
	if inner == nil || ttl <= 0 {
		return inner
	}

	return &cachingTokenValidator{
		inner:   inner,
		ttl:     ttl,
		entries: make(map[string]cachedTokenClaims),
	}
}

func (v *cachingTokenValidator) ValidateToken(ctx context.Context, token string) (int64, error) {
	claims, err := v.ValidateTokenFull(ctx, token)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

func (v *cachingTokenValidator) ValidateTokenFull(ctx context.Context, token string) (*TokenClaims, error) {
	now := time.Now()
	key := tokenValidationCacheKey(token)

	if claims, ok := v.get(key, now); ok {
		return claims, nil
	}

	resultCh := v.group.DoChan(key, func() (any, error) {
		now := time.Now()
		if claims, ok := v.get(key, now); ok {
			return claims, nil
		}

		validationCtx, cancel := tokenValidationContext(ctx)
		defer cancel()

		claims, err := v.inner.ValidateTokenFull(validationCtx, token)
		if err != nil {
			return nil, err
		}
		if claims == nil {
			return nil, errNilTokenClaims
		}

		v.put(key, claims, now)
		return cloneTokenClaims(claims), nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}

		claims, ok := result.Val.(*TokenClaims)
		if !ok || claims == nil {
			return nil, errNilTokenClaims
		}
		return cloneTokenClaims(claims), nil
	}
}

func tokenValidationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(ctx)
	if deadline, ok := ctx.Deadline(); ok {
		return context.WithDeadline(base, deadline)
	}

	return base, func() {}
}

func (v *cachingTokenValidator) get(key string, now time.Time) (*TokenClaims, bool) {
	v.mu.RLock()
	entry, ok := v.entries[key]
	v.mu.RUnlock()
	if !ok || !entry.expiresAt.After(now) {
		return nil, false
	}

	return cloneTokenClaims(&entry.claims), true
}

func (v *cachingTokenValidator) put(key string, claims *TokenClaims, now time.Time) {
	expiresAt := now.Add(v.ttl)
	if !claims.ExpiresAt.IsZero() && claims.ExpiresAt.Before(expiresAt) {
		expiresAt = claims.ExpiresAt
	}
	if !expiresAt.After(now) {
		return
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.lastPrune.IsZero() || now.Sub(v.lastPrune) >= v.ttl {
		for entryKey, entry := range v.entries {
			if !entry.expiresAt.After(now) {
				delete(v.entries, entryKey)
			}
		}
		v.lastPrune = now
	}

	v.entries[key] = cachedTokenClaims{
		claims:    *cloneTokenClaims(claims),
		expiresAt: expiresAt,
	}
}

func tokenValidationCacheKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func cloneTokenClaims(claims *TokenClaims) *TokenClaims {
	if claims == nil {
		return nil
	}
	cloned := *claims
	return &cloned
}
