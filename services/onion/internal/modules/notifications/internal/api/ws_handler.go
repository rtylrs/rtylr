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
// Authorization header on WebSocket handshakes. Bearer tokens are no longer
// accepted in query parameters, so clients authenticate with a one-time ticket.
//
// TODO: this module has no /dash/ws-ticket issuer yet (workspaces serves
// /flow/ws-ticket and pos_orders serves /pos/ws-ticket), so dashboard browsers
// currently have no supported way to obtain a ticket and /dash/ws rejects
// them. Add the exchange and migrate the dashboard client.
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
