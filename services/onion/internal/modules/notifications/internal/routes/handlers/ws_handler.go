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

// DashWSRoutes returns the /dash/ws endpoint.
//
// Anonymous at the middleware level because browsers cannot set an
// Authorization header on WebSocket handshakes. The JWT is passed via
// ?token= and validated here before upgrading.
func DashWSRoutes() []string { return []string{"/dash/ws"} }

type dashWSClaims struct {
	UserID     int64  `json:"user_id"`
	BusinessID int64  `json:"business_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// ServeDashWS upgrades the HTTP request to a WebSocket and hands it off
// to the global wsbus hub. Clients subscribe to user:{id} topics to
// receive real-time notification invalidation hints.
func ServeDashWS(w http.ResponseWriter, r *http.Request) {
	hub := wsbus.Default()
	if hub == nil {
		http.Error(w, "realtime not configured", http.StatusServiceUnavailable)
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
		logger.Ctx(r.Context()).Error("dash wsbus: jwt_secret not available on ctx")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	helper := jwtutils.NewJWTHelper(jwtSecret, "onion-platform")

	var claims dashWSClaims
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
