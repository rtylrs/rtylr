package http

import (
	"net/http"
)

// securityResponseHeaders is the static set of hardening headers applied to
// every response. Kept as a table so the set is easy to audit and extend.
//
// Cross-origin choices are deliberate: this server is an API + CDN consumed
// cross-site by *.rtylr.com / *.voxire.com, so Cross-Origin-Resource-Policy is
// "cross-origin" (assets must stay embeddable) and COOP allows popups (so the
// WorkOS OAuth flow is never broken). COEP and a same-site CORP are
// intentionally omitted — they would break legitimate cross-origin embedding.
var securityResponseHeaders = map[string]string{
	"X-Content-Type-Options": "nosniff",
	"X-Frame-Options":        "DENY",
	// X-XSS-Protection is disabled (0) per current OWASP guidance — the legacy
	// auditor it enabled is removed from modern browsers and can introduce
	// information-leak side channels in old ones.
	"X-XSS-Protection":                  "0",
	"Strict-Transport-Security":         "max-age=31536000; includeSubDomains; preload",
	"Referrer-Policy":                   "strict-origin-when-cross-origin",
	"Permissions-Policy":                "accelerometer=(), autoplay=(), browsing-topics=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()",
	"X-Permitted-Cross-Domain-Policies": "none",
	"Cross-Origin-Opener-Policy":        "same-origin-allow-popups",
	"Cross-Origin-Resource-Policy":      "cross-origin",
	"Content-Security-Policy":           "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'",
}

func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			for key, value := range securityResponseHeaders {
				h.Set(key, value)
			}

			h.Del("Server")
			h.Del("X-Powered-By")

			next.ServeHTTP(w, r)
		})
	}
}
