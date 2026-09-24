package http

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
)

func hasValidInternalToken(r *http.Request) bool {
	return hasValidInternalTokenForConfig(r, nil)
}

func hasValidInternalTokenForConfig(r *http.Request, cfg *config.AppConfig) bool {
	expected := internalToken()
	if cfg != nil {
		if configured := strings.TrimSpace(cfg.InternalToken); configured != "" {
			expected = configured
		}
	}
	if expected == "" {
		return false
	}

	actual := r.Header.Get("Internal-Token")
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func internalToken() string {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.InternalToken
	}
	return os.Getenv("INTERNAL_TOKEN")
}
