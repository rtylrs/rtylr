package http

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/voxire/onion/internal/platform/config"
)

func hasValidInternalToken(r *http.Request) bool {
	expected := internalToken()
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
