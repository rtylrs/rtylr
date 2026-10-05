package tenant

import (
	"context"
	"errors"
	"sync"
)

var ErrOwnerLookupFailed = errors.New("failed to resolve business owner")
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

// ResolveOwnerID is for billing aggregation and legacy owner columns, not tenant data scoping.
func ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) (int64, error) {
	return resolveOwnerID(ctx, currentResolver(), businessID, fallbackUserID)
}

func resolveOwnerID(ctx context.Context, reader ScopeResolver, businessID, _ int64) (int64, error) {
	if businessID <= 0 {
		return 0, ErrNoBusinessScope
	}
	if reader == nil {
		return 0, ErrOwnerLookupFailed
	}
	ownerID, err := reader.ActiveBusinessOwner(ctx, businessID)
	if err != nil || ownerID <= 0 {
		return 0, ErrOwnerLookupFailed
	}
	return ownerID, nil
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
