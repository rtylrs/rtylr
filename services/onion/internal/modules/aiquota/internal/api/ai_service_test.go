package api

import (
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestCanUpdateAiQuotaSettings(t *testing.T) {
	tests := []struct {
		name string
		user *authPkg.AuthUser
		want bool
	}{
		{name: "owner", user: &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}}, want: true},
		{name: "admin", user: &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}}, want: true},
		{name: "quota permission", user: &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{DashPermissions: []string{"hub:ai:quota:manage"}}}, want: true},
		{name: "unrelated permission", user: &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{DashPermissions: []string{"hub:team:members:read"}}}},
		{name: "console admin alone", user: &authPkg.AuthUser{IsConsoleAdmin: true}},
		{name: "missing user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canUpdateAiQuotaSettings(tt.user); got != tt.want {
				t.Fatalf("canUpdateAiQuotaSettings(%+v) = %v, want %v", tt.user, got, tt.want)
			}
		})
	}
}
