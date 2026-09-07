package api

import (
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// DashWSRoutes returns the /dash/ws endpoint.
//
// Anonymous at the middleware level because browsers cannot set an
// Authorization header on WebSocket handshakes. Existing browser clients
// pass a JWT via ?token= and it is validated here before upgrading. Flow's
// WebSocket path uses the newer one-time ticket flow.
func DashWSRoutes() []string { return []string{"/dash/ws"} }

// ServeDashWS upgrades the HTTP request to a WebSocket and hands it off
// to the global wsbus hub. Clients subscribe to user:{id} topics to
// receive real-time notification invalidation hints.
func ServeDashWS(w http.ResponseWriter, r *http.Request) {
	hub := wsbus.Default()
	if hub == nil {
		http.Error(w, "realtime not configured", http.StatusServiceUnavailable)
		return
	}

	authUser, err := authPkg.AuthUserFromWebSocketRequest(r)
	if errors.Is(err, authPkg.ErrJWTSecretMissing) {
		logger.Ctx(r.Context()).Error("dash wsbus: jwt_secret not available on ctx")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := authPkg.WithAuthUser(r.Context(), authUser)
	hub.ServeWS(w, r.WithContext(ctx), authUser.UserID)
}
