package realtime

import (
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
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

// ServeRealtime upgrades the HTTP request to a WebSocket and hands it
// off to the default wsbus hub. Standard chi handler signature — wire it
// as an anonymous route via RegisterRaw or similar.
func ServeRealtime(w http.ResponseWriter, r *http.Request) {
	hub := wsbus.Default()
	if hub == nil {
		http.Error(w, "realtime not configured", http.StatusServiceUnavailable)
		return
	}

	authUser, err := authPkg.AuthUserFromWebSocketRequest(r)
	if errors.Is(err, authPkg.ErrJWTSecretMissing) {
		logger.Ctx(r.Context()).Error("wsbus: jwt_secret not available on ctx")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Attach the userID to the request context so downstream observers
	// (log, metrics) can see it.
	ctx := authPkg.WithAuthUser(r.Context(), authUser)
	hub.ServeWS(w, r.WithContext(ctx), authUser.UserID)
}
