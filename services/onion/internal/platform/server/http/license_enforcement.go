package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
)

// licenseStatusReporter is the subset of license.CachedValidator the
// enforcement middleware needs. It exists so the middleware can be tested with
// a stub and so the dependency is explicit.
type licenseStatusReporter interface {
	Current(ctx context.Context) license.StatusReport
}

// licenseUtilityRoutes are anonymous/utility endpoints that must always load
// even when a self-host license is invalid, so the admin can see the notice and
// log in. These are matched in addition to whatever the access policy already
// treats as anonymous, so the gate never locks an operator out of recovery.
var licenseUtilityRoutes = map[string]struct{}{
	"/health":                 {},
	"/version":                {},
	"/status":                 {},
	"/license/state":          {},
	"/runtime-config":         {},
	"/.well-known/rtylr.json": {},
}

// LicenseEnforcementMiddleware blocks protected API routes when this instance is
// self-hosted and its license is invalid (expired/missing/revoked/unreachable
// past grace). It returns 402 Payment Required with a small JSON body.
//
// In hosted mode (license.IsSelfHosted == false) it is a strict no-op: the
// validator is never consulted and every request passes through unchanged, so
// the hosted request path is untouched.
//
// Anonymous routes (per the access policy) and the utility routes above always
// pass through even when invalid, so login and the admin notice can still load.
// Place this AFTER auth in the chain so it only guards real API calls.
func LicenseEnforcementMiddleware(validator licenseStatusReporter, accessPolicy *RouteAccessPolicy) MiddlewareFunc {
	selfHosted := license.IsSelfHosted(config.GlobalConfig)

	return func(next http.Handler) http.Handler {
		// Hosted mode: zero behavior change, no validator, no per-request cost.
		if !selfHosted || validator == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if _, ok := licenseUtilityRoutes[path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			if accessPolicy != nil && accessPolicy.IsAllowedAnonymously(path) {
				next.ServeHTTP(w, r)
				return
			}

			report := validator.Current(r.Context())
			if report.Valid {
				next.ServeHTTP(w, r)
				return
			}

			writeLicenseInvalidResponse(w, report.Status)
		})
	}
}

func writeLicenseInvalidResponse(w http.ResponseWriter, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error":  "license_invalid",
		"status": status,
	}); err != nil {
		logger.Global().Warn("failed to encode license invalid response", logger.Error(err))
	}
}
