package handlers

import (
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// POSWebSocketRoutes returns the single /pos/ws endpoint.
//
// This route is listed as anonymous at the middleware layer because browsers
// cannot set an Authorization header on a WebSocket handshake. Existing
// browser clients pass a JWT in the `?token=` query param and it is validated
// here before upgrading. Flow's WebSocket path uses the newer one-time ticket
// flow.
func POSWebSocketRoutes() []string { return []string{"/pos/ws"} }

// ServePOSWebSocket upgrades the HTTP request to a WebSocket and hands it
// off to the default wsbus hub. Registered as an anonymous chi route so the
// global JWT middleware is skipped; token validation is performed here.
func ServePOSWebSocket(w http.ResponseWriter, r *http.Request) {
	hub := wsbus.Default()
	if hub == nil {
		http.Error(w, "realtime not available", http.StatusServiceUnavailable)
		return
	}

	authUser, err := authPkg.AuthUserFromWebSocketRequest(r)
	if errors.Is(err, authPkg.ErrJWTSecretMissing) {
		logger.Ctx(r.Context()).Error("pos wsbus: jwt_secret not available on ctx")
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
