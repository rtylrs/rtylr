package tenant

import (
	"context"
	"errors"
	"sync"
)

var ErrNoBusinessScope = errors.New("no business scope for request")

// ScopeResolver supplies verified active identity from the owning module.
// Reads must reflect current grants, including revocations. Unresolved scope returns an error.
type ScopeResolver interface {
	ActiveBusinessOwner(context.Context, int64) (int64, error)
	UserBusiness(context.Context, int64, int64) (int64, error)
}

var scopeResolver struct {
	sync.RWMutex
	reader ScopeResolver
}

// SetScopeResolver configures the process composition root; restore supports isolated test scopes.
func SetScopeResolver(reader ScopeResolver) (restore func()) {
	scopeResolver.Lock()
	previous := scopeResolver.reader
	scopeResolver.reader = reader
	scopeResolver.Unlock()
	return func() {
		scopeResolver.Lock()
		scopeResolver.reader = previous
		scopeResolver.Unlock()
	}
}

func currentResolver() ScopeResolver {
	scopeResolver.RLock()
	defer scopeResolver.RUnlock()
	return scopeResolver.reader
}

// ActiveBusinessID verifies that businessID names an active business, for
// business-scoped reads that run without a user, such as billing.
func ActiveBusinessID(ctx context.Context, businessID int64) (int64, error) {
	return activeBusinessID(ctx, currentResolver(), businessID)
}

func activeBusinessID(ctx context.Context, reader ScopeResolver, businessID int64) (int64, error) {
	if businessID <= 0 || reader == nil {
		return 0, ErrNoBusinessScope
	}
	if ownerID, err := reader.ActiveBusinessOwner(ctx, businessID); err != nil || ownerID <= 0 {
		return 0, ErrNoBusinessScope
	}
	return businessID, nil
}

// ResolveBusinessID preserves each business scope and rejects an unverified token or default business.
func ResolveBusinessID(ctx context.Context, businessID, fallbackUserID int64) (int64, error) {
	return resolveBusinessID(ctx, currentResolver(), businessID, fallbackUserID)
}

func resolveBusinessID(ctx context.Context, reader ScopeResolver, businessID, userID int64) (int64, error) {
	if reader == nil || userID <= 0 {
		return 0, ErrNoBusinessScope
	}
	resolved, err := reader.UserBusiness(ctx, businessID, userID)
	if err != nil || resolved <= 0 {
		return 0, ErrNoBusinessScope
	}
	return resolved, nil
}
