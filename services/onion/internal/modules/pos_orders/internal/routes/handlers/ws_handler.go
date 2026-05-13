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

// POSWebSocketRoutes returns the single /pos/ws endpoint.
//
// This route is listed as anonymous at the middleware layer because browsers
// cannot set an Authorization header on a WebSocket handshake. The JWT is
// passed in the `?token=` query param and validated here before upgrading.
func POSWebSocketRoutes() []string { return []string{"/pos/ws"} }

// posWSClaims is the minimum claim set needed to identify the caller.
type posWSClaims struct {
	UserID     int64  `json:"user_id"`
	BusinessID int64  `json:"business_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// ServePOSWebSocket upgrades the HTTP request to a WebSocket and hands it
// off to the default wsbus hub. Registered as an anonymous chi route so the
// global JWT middleware is skipped; token validation is performed here.
func ServePOSWebSocket(w http.ResponseWriter, r *http.Request) {
	hub := wsbus.Default()
	if hub == nil {
		http.Error(w, "realtime not available", http.StatusServiceUnavailable)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		auth := r.Header.Get("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	jwtSecret, _ := r.Context().Value(jwtutils.JWTSecretKey).(string)
	if jwtSecret == "" {
		logger.Ctx(r.Context()).Error("pos wsbus: jwt_secret not available on ctx")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	helper := jwtutils.NewJWTHelper(jwtSecret, "onion-platform")

	var claims posWSClaims
	if valErr := helper.ValidateToken(r.Context(), token, &claims); valErr != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if claims.UserID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.WithValue(r.Context(), authPkg.AuthUserKey, &authPkg.AuthUser{
		UserID:     claims.UserID,
		BusinessID: claims.BusinessID,
	})
	hub.ServeWS(w, r.WithContext(ctx), claims.UserID)
}
