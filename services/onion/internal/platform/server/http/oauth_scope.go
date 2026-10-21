package http

import (
	"context"
	"net/http"
	pathpkg "path"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// requiredOAuthScope maps a request to the scope an OAuth access token needs.
// allowed is false for routes OAuth tokens may never reach; an empty scope with
// allowed true means any valid OAuth token may call the route.
func requiredOAuthScope(method, requestPath string) (scope string, allowed bool) {
	normalized := strings.ToLower(pathpkg.Clean("/" + requestPath))
	switch normalized {
	case "/common/oauth/v1/me", "/common/oauth/v1/revoke":
		return "", true
	}
	app, _, _ := strings.Cut(strings.TrimPrefix(normalized, "/"), "/")
	for _, oauthApp := range authPkg.OAuthApps {
		if app != oauthApp {
			continue
		}
		access := authPkg.OAuthAccessWrite
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			access = authPkg.OAuthAccessRead
		}
		return authPkg.OAuthScope(app, access), true
	}
	return "", false
}

func writeInsufficientScope(ctx context.Context, w http.ResponseWriter, required string) {
	challenge := `Bearer error="insufficient_scope"`
	if required != "" {
		challenge += `, scope="` + required + `"`
	}
	w.Header().Set("WWW-Authenticate", challenge)
	WriteOneKitError(ctx, w, onekiterrors.New(http.StatusForbidden, "insufficient_scope"))
}
