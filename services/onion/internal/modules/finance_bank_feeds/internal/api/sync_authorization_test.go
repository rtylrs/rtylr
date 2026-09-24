package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func bankFeedAuthUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     7001,
		BusinessID: 8001,
		Role:       "member",
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           8001,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
				Scopes:      []authorization.Scope{{Type: "finance", ID: "8001"}},
			}},
		},
	}
}

func TestSyncAuthorizationDoesNotAcceptViewOnlyPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	if canSyncBankFeeds(bankFeedAuthUser("finance:bank_feeds:view")) {
		t.Fatal("view-only bank-feed permission unexpectedly authorizes sync")
	}

	_, err := (&SyncHandler{}).Serve(
		authPkg.WithAuthUser(context.Background(), bankFeedAuthUser("finance:bank_feeds:view")),
		&SyncRequest{},
	)
	if err == nil {
		t.Fatal("view-only sync request unexpectedly succeeded")
	}
}

func TestSyncAuthorizationAcceptsMutationGrants(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	for _, permission := range []string{bankFeedSyncPermission, bankFeedReconciliationPermission} {
		if !canSyncBankFeeds(bankFeedAuthUser(permission)) {
			t.Errorf("mutation permission %q was rejected", permission)
		}
	}
}

func TestSyncAuthorizationDoesNotUseLegacyAppMembershipFallback(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "legacy")

	user := bankFeedAuthUser("finance:bank_feeds:view")
	user.Role = "finance.viewer"
	if canSyncBankFeeds(user) {
		t.Fatal("legacy view-only membership unexpectedly authorizes sync")
	}
}
