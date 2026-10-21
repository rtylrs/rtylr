package auth

import "github.com/voxire/onion/internal/platform/authorization"

func IsBusinessAdmin(user *AuthUser) bool {
	if !hasActiveProvisioning(user) {
		return false
	}
	if user != nil && user.EffectiveAccess != nil {
		return user.EffectiveAccess.IsOwner || user.EffectiveAccess.IsAdmin
	}
	return false
}

func IsConsoleAdmin(user *AuthUser) bool {
	return user != nil && user.IsConsoleAdmin
}

// HasAppPermission checks only server-resolved app access. It deliberately
// does not fall back to the legacy role field. The resolver requires both an
// active license and a permission profile assignment for regular members.
func HasAppPermission(user *AuthUser, appID authorization.AppID, permission string) bool {
	if !hasActiveProvisioning(user) || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID == appID && app.Licensed && app.HasAccess && (user.EffectiveAccess.IsOwner || user.EffectiveAccess.IsAdmin) {
			return true
		}
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID != appID || !app.Licensed || !app.HasAccess {
			continue
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
	if user.EffectiveAccess.IsOwner || user.EffectiveAccess.IsAdmin {
		return true
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

// HasAnyAppScope reports whether the resolved app grants any resource access.
// An empty scope list is a business-wide grant; a non-empty list is restricted
// to those resources. For apps that do not model scopes, resolved app access
// is sufficient and the business_id predicate remains the tenant boundary.
func HasAnyAppScope(user *AuthUser, appID authorization.AppID) bool {
	if !hasActiveProvisioning(user) || user == nil || user.EffectiveAccess == nil {
		return false
	}
	for _, app := range user.EffectiveAccess.Apps {
		if app.AppID != appID || !app.Licensed || !app.HasAccess {
			continue
		}
		// Owners and admins receive business-wide scope through the system path.
		if user.EffectiveAccess.IsOwner || user.EffectiveAccess.IsAdmin {
			return true
		}
		if !appRequiresScope(appID) {
			return true
		}
		return len(app.PermissionScopes) > 0
	}
	return false
}

func hasActiveProvisioning(user *AuthUser) bool {
	if user == nil || user.ProvisioningStatus == nil {
		return user != nil
	}
	return user.ProvisioningStatus.State == authorization.ProvisioningActive
}
