package auth

import (
	"strings"
)

const (
	OAuthAccessRead  = "read"
	OAuthAccessWrite = "write"
)

// OAuthApps are the apps an OAuth client can be granted access to.
var OAuthApps = []string{"flow", "crm", "finance", "erp", "pos"}

func OAuthScope(app, access string) string {
	return app + ":" + access
}

// ParseOAuthScopes splits a space-delimited scope string, dropping duplicates.
func ParseOAuthScopes(scope string) []string {
	fields := strings.Fields(scope)
	seen := make(map[string]struct{}, len(fields))
	scopes := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		scopes = append(scopes, field)
	}
	return scopes
}

// IsOAuthToken reports whether the request was authenticated with an OAuth
// access token, which is limited to its granted scopes.
func (u *AuthUser) IsOAuthToken() bool {
	return u != nil && u.OAuthClientID != ""
}

// HasOAuthScope reports whether a scope-limited token grants the required
// scope. Write access implies read access to the same app.
func (u *AuthUser) HasOAuthScope(required string) bool {
	if u == nil {
		return false
	}
	app, access, ok := strings.Cut(required, ":")
	if !ok {
		return false
	}
	for _, granted := range u.OAuthScopes {
		if granted == required {
			return true
		}
		if access == OAuthAccessRead && granted == OAuthScope(app, OAuthAccessWrite) {
			return true
		}
	}
	return false
}
