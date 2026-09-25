package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
)

// RegisterLicenseStateRoutes exposes the current license status so frontends
// can show an admin-only notice when a self-host license is missing, expiring,
// expired, or the authority is unreachable. In hosted mode this always reports
// a valid hosted status and never touches the network.
//
// GET /license/state -> license.StatusReport (anonymous; carries no secrets —
// the license key itself is never returned).
func RegisterLicenseStateRoutes(r *chi.Mux, validator *license.CachedValidator) {
	r.Get("/license/state", func(w http.ResponseWriter, req *http.Request) {
		report := validator.Current(req.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(report); err != nil {
			logger.Ctx(req.Context()).Warn("failed to encode license state response", logger.Error(err))
		}
	})
}
