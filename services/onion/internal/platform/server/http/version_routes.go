package http

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/go-chi/chi/v5"
)

// These are injected at build time via -ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

type versionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func RegisterVersionRoutes(r *chi.Mux) {
	r.Get("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(versionResponse{
			Version:   Version,
			Commit:    Commit,
			BuildTime: BuildTime,
			GoVersion: runtime.Version(),
		}); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})
}
