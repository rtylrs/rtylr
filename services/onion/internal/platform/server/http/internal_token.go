package http

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
)

const internalTokenHeader = "Internal-Token"

func hasValidInternalToken(r *http.Request) bool {
	return hasValidInternalTokenForConfig(r, nil)
}

// hasValidInternalTokenForConfig reports whether r carries the configured
// internal-service token. It is total and fails closed: true is returned only
// for an exact, constant-time match against a non-blank configured token.
// Every other outcome — no token configured, blank token, nil request, missing
// or blank header, prefix, or any other mismatch — denies.
//
// Callers must not gate this behind a "was a token configured" test of their
// own. A token-less server is a server that serves no internal requests, not a
// server that serves them to anyone.
func hasValidInternalTokenForConfig(r *http.Request, cfg *config.AppConfig) bool {
	if r == nil {
		return false
	}

	expected := internalToken()
	if cfg != nil {
		if configured := strings.TrimSpace(cfg.InternalToken); configured != "" {
			expected = configured
		}
	}
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return false
	}

	actual := r.Header.Get(internalTokenHeader)
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func internalToken() string {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.InternalToken
	}
	return os.Getenv("INTERNAL_TOKEN")
}
