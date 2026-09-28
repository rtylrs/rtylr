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
// does not fall back to the legacy role field. Licensing is plan metadata and
// is not part of this decision: access comes from the member's role
// assignments, which the resolver has already turned into an app entry.
func HasAppPermission(user *AuthUser, appID authorization.AppID, permission string) bool {
	if !hasActiveProvisioning(user) || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID != appID || !app.HasAccess {
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
		if app.AppID == appID && app.HasAccess {
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

// appRequiresScope reports whether an app's assignments are resource scoped.
// Only POS (locations) and Flow (projects) model scopes today, and the
// chk_auth_assignment_scope_type constraint permits no other scope type, so
// requiring a scope on any other app could never be satisfied and would lock
// every non-owner member out of it. The allowlist is deliberately explicit:
// adding an app here without adding a permitted scope type would deny access
// rather than widen it.
func appRequiresScope(appID authorization.AppID) bool {
	return appID == authorization.AppPOS || appID == authorization.AppFlow
}

// HasAnyAppScope reports whether an assignment for a resource scoped app
// carries at least one scope. For apps that model no scopes it reports true
// once the app is resolved, leaving per-feature permissions and the
// business_id predicate as the boundary. A zero authorization version is the
// legacy compatibility shape used while an installation is being migrated, so
// it retains the existing business-level behavior until normalized access is
// available.
func HasAnyAppScope(user *AuthUser, appID authorization.AppID) bool {
	if !hasActiveProvisioning(user) || user == nil || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID != appID || !app.HasAccess {
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
		if !appRequiresScope(appID) {
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
