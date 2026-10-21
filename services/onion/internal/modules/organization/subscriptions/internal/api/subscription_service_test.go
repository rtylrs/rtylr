package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestBillingUserRequiresResolvedBillingAccess(t *testing.T) {
	tests := []struct {
		name    string
		user    *authPkg.AuthUser
		allowed bool
	}{
		{
			name:    "owner",
			user:    &authPkg.AuthUser{UserID: 1, EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}},
			allowed: true,
		},
		{
			name:    "admin",
			user:    &authPkg.AuthUser{UserID: 2, EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}},
			allowed: true,
		},
		{
			name:    "billing permission",
			user:    &authPkg.AuthUser{UserID: 3, EffectiveAccess: &authorization.EffectiveAccess{DashPermissions: []string{"hub:billing:manage"}}},
			allowed: true,
		},
		{
			name: "unrelated permission",
			user: &authPkg.AuthUser{UserID: 4, EffectiveAccess: &authorization.EffectiveAccess{DashPermissions: []string{"team:members:read"}}},
		},
		{
			name: "console admin alone",
			user: &authPkg.AuthUser{UserID: 5, IsConsoleAdmin: true},
		},
		{name: "zero user id", user: &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.user != nil {
				ctx = authPkg.WithAuthUser(ctx, test.user)
			}
			user, err := billingUser(ctx)
			if test.allowed {
				if err != nil || user != test.user {
					t.Fatalf("billingUser() = (%v, %v), want authenticated user and no error", user, err)
				}
				return
			}
			if err == nil || user != nil {
				t.Fatalf("billingUser() = (%v, %v), want denied", user, err)
			}
		})
	}
}

func TestConsoleAdminUserRequiresConsoleAdminFlag(t *testing.T) {
	for _, test := range []struct {
		name    string
		user    *authPkg.AuthUser
		allowed bool
	}{
		{name: "console admin", user: &authPkg.AuthUser{UserID: 1, IsConsoleAdmin: true}, allowed: true},
		{name: "tenant owner", user: &authPkg.AuthUser{UserID: 2, EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}}},
		{name: "tenant admin", user: &authPkg.AuthUser{UserID: 3, EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}}},
		{name: "missing user"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.user != nil {
				ctx = authPkg.WithAuthUser(ctx, test.user)
			}
			user, err := consoleAdminUser(ctx)
			if test.allowed {
				if err != nil || user != test.user {
					t.Fatalf("consoleAdminUser() = (%v, %v), want authenticated user and no error", user, err)
				}
				return
			}
			if err == nil || user != nil {
				t.Fatalf("consoleAdminUser() = (%v, %v), want denied", user, err)
			}
		})
	}
}
