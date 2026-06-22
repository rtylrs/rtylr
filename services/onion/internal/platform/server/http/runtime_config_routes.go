package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	platformconfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
)

// runtimeApps is the set of rtylr frontend apps a deployment serves. Frontends
// read this (plus the URLs) at boot so prebuilt images learn their environment
// instead of having it baked in.
var runtimeApps = []string{
	"dash", "pos", "erp", "hr", "crm", "finance", "flow", "insights", "recruit", "menu",
}

// runtimeAppDeps encodes app dependencies — enabling an app implies the others
// (e.g. recruit needs hr). Frontends/console use this to keep enablement coherent.
var runtimeAppDeps = map[string][]string{
	"recruit": {"hr"},
}

type runtimeConfigResponse struct {
	DeploymentMode     string               `json:"deployment_mode"`
	SelfHosted         bool                 `json:"self_hosted"`
	Version            string               `json:"version"`
	URLs               map[string]string    `json:"urls"`
	Apps               runtimeAppsBlock     `json:"apps"`
	AllFeaturesEnabled bool                 `json:"all_features_enabled"`
	License            license.StatusReport `json:"license"`
}

type runtimeAppsBlock struct {
	Enabled      []string            `json:"enabled"`
	Dependencies map[string][]string `json:"dependencies"`
}

// RegisterRuntimeConfigRoutes serves the per-deployment runtime configuration
// (URLs, enabled apps, self-host flags, license summary) to frontends.
//
// GET /runtime-config and GET /.well-known/rtylr.json (both anonymous, no
// secrets — never includes the license key). In self-host mode all_features_
// enabled is true so frontends hide the Addons tab and treat every feature
// flag as on.
func RegisterRuntimeConfigRoutes(r *chi.Mux, validator *license.CachedValidator) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		cfg := platformconfig.GlobalConfig
		selfHosted := license.IsSelfHosted(cfg)
		mode := "hosted"
		if selfHosted {
			mode = "self_hosted"
		}
		resp := runtimeConfigResponse{
			DeploymentMode:     mode,
			SelfHosted:         selfHosted,
			Version:            Version,
			URLs:               platformconfig.RuntimeURLs(cfg),
			AllFeaturesEnabled: selfHosted,
			License:            validator.Current(req.Context()),
			Apps: runtimeAppsBlock{
				Enabled:      runtimeApps,
				Dependencies: runtimeAppDeps,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
	r.Get("/runtime-config", handler)
	r.Get("/.well-known/rtylr.json", handler)
}
