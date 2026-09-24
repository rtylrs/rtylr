package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	platformconfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
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
	DeploymentMode     string                `json:"deployment_mode"`
	SelfHosted         bool                  `json:"self_hosted"`
	Version            string                `json:"version"`
	URLs               map[string]string     `json:"urls"`
	Apps               runtimeAppsBlock      `json:"apps"`
	AllFeaturesEnabled bool                  `json:"all_features_enabled"`
	License            runtimeLicenseSummary `json:"license"`
}

type runtimeLicenseSummary struct {
	Valid      bool       `json:"valid"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason,omitempty"`
	Plan       string     `json:"plan,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	GraceUntil *time.Time `json:"grace_until,omitempty"`
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
		urls, err := runtimeURLsForMode(cfg, selfHosted, req)
		if err != nil {
			http.Error(w, "self-hosted runtime configuration requires a non-hosted API URL", http.StatusInternalServerError)
			return
		}

		resp := runtimeConfigResponse{
			DeploymentMode:     mode,
			SelfHosted:         selfHosted,
			Version:            Version,
			URLs:               urls,
			AllFeaturesEnabled: selfHosted,
			License:            runtimeLicenseFrom(validator.Current(req.Context()), selfHosted),
			Apps: runtimeAppsBlock{
				Enabled:      runtimeApps,
				Dependencies: runtimeAppDeps,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Ctx(req.Context()).Warn("failed to encode runtime config response", logger.Error(err))
		}
	}
	r.Get("/runtime-config", handler)
	r.Get("/.well-known/rtylr.json", handler)
}

// runtimeURLsForMode keeps hosted defaults for hosted deployments, but never
// lets a self-hosted response point a prebuilt frontend at the hosted Voxire
// service. A self-host API URL is preferred; when it is omitted, derive a safe
// local origin from the request host so single-origin self-host installations
// still receive a usable response.
func runtimeURLsForMode(cfg *platformconfig.AppConfig, selfHosted bool, req *http.Request) (map[string]string, error) {
	urls := platformconfig.RuntimeURLs(cfg)
	if !selfHosted {
		return urls, nil
	}

	apiURL, err := selfHostedAPIURL(req)
	if err != nil {
		return nil, err
	}
	apiURL = strings.TrimRight(apiURL, "/")

	// RuntimeURLs supplies the hosted constants for omitted variables. Replace
	// every such default with the self-host API origin, while preserving any
	// explicitly configured customer URL.
	for key, value := range urls {
		if key == "api" || isHostedRuntimeURL(value) {
			urls[key] = apiURL
		}
	}
	urls["api"] = apiURL
	return urls, nil
}

func selfHostedAPIURL(req *http.Request) (string, error) {
	raw := strings.TrimSpace(os.Getenv("API_URL"))
	if raw == "" {
		if req == nil {
			return "", errors.New("self-hosted API URL is missing")
		}
		host := strings.TrimSpace(req.Host)
		if host == "" && req.URL != nil {
			host = strings.TrimSpace(req.URL.Host)
		}
		if host == "" {
			return "", errors.New("self-hosted API URL is missing")
		}
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		raw = scheme + "://" + host
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("self-hosted API URL must be an absolute http(s) URL")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if parsed.Scheme == "" || parsed.Host == "" || (scheme != "http" && scheme != "https") {
		return "", errors.New("self-hosted API URL must be an absolute http(s) URL")
	}
	if isHostedRuntimeURL(raw) {
		return "", errors.New("self-hosted API URL must not point at hosted Voxire URLs")
	}
	return raw, nil
}

func isHostedRuntimeURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	return host == "voxire.com" || host == "rtylr.com" ||
		strings.HasSuffix(host, ".voxire.com") || strings.HasSuffix(host, ".rtylr.com")
}

func runtimeLicenseFrom(report license.StatusReport, selfHosted bool) runtimeLicenseSummary {
	summary := runtimeLicenseSummary{
		Valid:  report.Valid,
		Status: report.Status,
		Reason: report.Reason,
	}
	if !selfHosted {
		return summary
	}
	summary.Plan = report.Plan
	summary.ExpiresAt = report.ExpiresAt
	summary.GraceUntil = report.GraceUntil
	return summary
}
