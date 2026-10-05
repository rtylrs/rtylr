package realtime

import (
	"encoding/json"
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// RealtimeRoutes returns the browser WebSocket endpoint. It is intentionally
// anonymous at the middleware level because browsers cannot set an
// Authorization header on a WebSocket handshake. The handler authenticates
// the upgrade with a short-lived, single-use ticket instead.
//
// IssueRealtimeTicket exchanges the normal authenticated API session for an
// opaque, short-lived WebSocket ticket. This route is deliberately not part
// of GetAnonymousRoutes, so the standard Bearer middleware authenticates it.
//
//nolint:revive // The name matches the module route registration convention.
func IssueRealtimeTicket(w http.ResponseWriter, r *http.Request) {
	authUser, ok := authPkg.GetAuthUser(r.Context())
	if !ok || authUser == nil || authUser.UserID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ticket, err := authPkg.IssueWebSocketTicket(r.Context(), authUser)
	if err != nil {
		logger.Ctx(r.Context()).Error("wsbus: failed to issue websocket ticket", logger.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Ticket string `json:"ticket"`
	}{Ticket: ticket}); err != nil {
		logger.Ctx(r.Context()).Warn("wsbus: failed to write websocket ticket", logger.Error(err))
	}
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
