package handlers

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// RealtimeRoutes returns the single /flow/ws endpoint.
//
// This route is intentionally listed as anonymous at the middleware level
// because browsers cannot set an Authorization header on a WebSocket
// handshake (the `WebSocket` constructor accepts no headers). The client
// therefore passes the JWT in the `?token=` query param and we validate
// it here inside the handler before upgrading.
func RealtimeRoutes() []string { return []string{"/flow/ws"} }

// accessTokenClaims mirrors the minimum set of fields we need to identify
// the caller. Defined locally so this handler doesn't depend on the user
// module's private types.
type accessTokenClaims struct {
	UserID     int64  `json:"user_id"`
	BusinessID int64  `json:"business_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// ServeRealtime upgrades the HTTP request to a WebSocket and hands it
// off to the default wsbus hub. Standard chi handler signature — wire it
// as an anonymous route via RegisterRaw or similar.
func ServeRealtime(w http.ResponseWriter, r *http.Request) {
	hub := wsbus.Default()
	if hub == nil {
		http.Error(w, "realtime not configured", http.StatusServiceUnavailable)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		// Fall back to Authorization header for server-to-server callers
		// that can set it (curl, tests).
		auth := r.Header.Get("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// The JWT secret and issuer live on the request context — the platform
	// middleware seeds them before handing off to any route.
	jwtSecret, _ := r.Context().Value(jwtutils.JWTSecretKey).(string)
	if jwtSecret == "" {
		logger.Ctx(r.Context()).Error("wsbus: jwt_secret not available on ctx")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	helper := jwtutils.NewJWTHelper(jwtSecret, "onion-platform")

	var claims accessTokenClaims
	if valErr := helper.ValidateToken(r.Context(), token, &claims); valErr != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.UserID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Attach the userID to the request context so downstream observers
	// (log, metrics) can see it.
	ctx := context.WithValue(r.Context(), authPkg.AuthUserKey, &authPkg.AuthUser{
		UserID: claims.UserID,
	})
	hub.ServeWS(w, r.WithContext(ctx), claims.UserID)
}
