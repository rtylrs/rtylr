package auth

import (
	"strings"

	"github.com/voxire/onion/internal/platform/authorization"
)

// HasAnyRole reports whether an authenticated user has one of the supplied
// tenant roles. Callers should still resolve and validate the tenant scope
// separately; this helper only answers the role portion of authorization.
func HasAnyRole(user *AuthUser, roles ...string) bool {
	if user == nil {
		return false
	}

	actual := strings.ToLower(strings.TrimSpace(user.Role))
	for _, role := range roles {
		if actual == strings.ToLower(strings.TrimSpace(role)) {
			return true
		}
	}
	return false
}

func IsBusinessAdmin(user *AuthUser) bool {
	if !hasActiveProvisioning(user) {
		return false
	}
	if user != nil && user.EffectiveAccess != nil {
		return user.EffectiveAccess.IsOwner || user.EffectiveAccess.IsAdmin
	}
	return HasAnyRole(user, "owner", "admin", "console_admin")
}

// HasAppPermission checks only server-resolved app access. It deliberately
// does not fall back to the legacy role field.
func HasAppPermission(user *AuthUser, appID authorization.AppID, permission string) bool {
	if !hasActiveProvisioning(user) || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID != appID || !app.Licensed || !app.HasAccess {
			continue
		}
		if LegacyAuthorizationMode() {
			return true
		}
		for _, granted := range app.Permissions {
			if granted == permission {
				return true
			}
		}
	}
	return false
}

// HasAppAccess checks the license and resolved assignment boundary without
// requiring a feature-specific permission. It is used by the HTTP boundary
// to prevent an authenticated tenant user from reaching an app they do not
// own before a module applies its finer-grained permission check.
func HasAppAccess(user *AuthUser, appID authorization.AppID) bool {
	if !hasActiveProvisioning(user) || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID == appID && app.Licensed && app.HasAccess {
			return true
		}
	}
	return false
}

func HasDashboardPermission(user *AuthUser, permission string) bool {
	if !hasActiveProvisioning(user) || user.EffectiveAccess == nil {
		return false
	}
	for _, granted := range user.EffectiveAccess.DashPermissions {
		if granted == permission {
			return true
		}
	}
	return false
}

// HasAnyAppScope reports whether a normalized app assignment has at least one
// resource scope. A zero authorization version is the legacy compatibility
// shape used while an installation is being migrated, so it retains the
// existing business-level behavior until normalized access is available.
func HasAnyAppScope(user *AuthUser, appID authorization.AppID) bool {
	if !hasActiveProvisioning(user) || user == nil || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID != appID || !app.Licensed || !app.HasAccess {
			continue
		}
		if user.EffectiveAccess.AuthorizationVersion == 0 {
			return true
		}
		// Owners and admins resolve through the system path, which grants every
		// permission without creating scoped assignments. HasAccess is only true
		// for them when their membership is active, so this cannot widen access.
		if user.EffectiveAccess.IsOwner || user.EffectiveAccess.IsAdmin {
			return true
		}
		return len(app.Scopes) > 0
	}
	return false
}

func hasActiveProvisioning(user *AuthUser) bool {
	if user == nil || user.ProvisioningStatus == nil {
		return user != nil
	}
	return user.ProvisioningStatus.State == authorization.ProvisioningActive
}
