package auth

import "strings"

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
	return HasAnyRole(user, "owner", "admin", "console_admin")
}
