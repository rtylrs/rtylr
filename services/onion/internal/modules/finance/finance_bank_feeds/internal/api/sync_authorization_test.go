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
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           8001,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func TestSyncAuthorizationDoesNotAcceptViewOnlyPermission(t *testing.T) {
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
	for _, permission := range []string{bankFeedSyncPermission, bankFeedReconciliationPermission} {
		if !canSyncBankFeeds(bankFeedAuthUser(permission)) {
			t.Errorf("mutation permission %q was rejected", permission)
		}
	}
}

func TestSyncAuthorizationRequiresMutationPermission(t *testing.T) {
	user := bankFeedAuthUser("finance:bank_feeds:view")
	if canSyncBankFeeds(user) {
		t.Fatal("view-only permission unexpectedly authorizes sync")
	}
}
