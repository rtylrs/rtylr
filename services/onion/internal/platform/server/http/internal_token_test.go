package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

// newInternalTokenRequest builds a request carrying the given Internal-Token
// header. An empty name omits the header entirely, as a real client would.
func newInternalTokenRequest(name, token string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/license/activate", nil)
	if name != "" {
		req.Header.Set(internalTokenHeader, token)
	}
	return req
}

func TestHasValidInternalTokenFailsClosed(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.AppConfig
		global     *config.AppConfig
		env        string
		headerName string
		header     string
		want       bool
	}{
		{
			name:       "no token configured anywhere",
			cfg:        &config.AppConfig{},
			global:     &config.AppConfig{},
			headerName: "Internal-Token",
			header:     "anything",
			want:       false,
		},
		{
			name:       "header supplied but no token configured",
			cfg:        &config.AppConfig{},
			global:     &config.AppConfig{},
			headerName: "",
			header:     "",
			want:       false,
		},
		{
			name:       "token configured but header absent",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: "",
			want:       false,
		},
		{
			name:       "empty header with token configured",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "",
			want:       false,
		},
		{
			name:       "whitespace-only configured token",
			cfg:        &config.AppConfig{InternalToken: "   "},
			headerName: internalTokenHeader,
			header:     "   ",
			want:       false,
		},
		{
			name:       "malformed token",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "not-the-secret",
			want:       false,
		},
		{
			name:       "empty supplied token",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "",
			want:       false,
		},
		{
			name:       "prefix of the correct token",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "sec",
			want:       false,
		},
		{
			name:       "correct token with trailing data",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "secret-and-more",
			want:       false,
		},
		{
			name:       "correct token differing in case",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "SECRET",
			want:       false,
		},
		{
			name:       "token from environment only",
			global:     nil,
			env:        "env-secret",
			headerName: internalTokenHeader,
			header:     "env-secret",
			want:       true,
		},
		{
			name:       "no configured token and empty environment",
			global:     nil,
			env:        "",
			headerName: internalTokenHeader,
			header:     "env-secret",
			want:       false,
		},
		{
			name:       "blank global config does not fall through to environment",
			global:     &config.AppConfig{},
			env:        "env-secret",
			headerName: internalTokenHeader,
			header:     "env-secret",
			want:       false,
		},
		{
			name:       "exact match",
			cfg:        &config.AppConfig{InternalToken: "secret"},
			headerName: internalTokenHeader,
			header:     "secret",
			want:       true,
		},
		{
			name:       "surrounding whitespace on configured token is tolerated",
			cfg:        &config.AppConfig{InternalToken: "  secret  "},
			headerName: internalTokenHeader,
			header:     "secret",
			want:       true,
		},
		{
			name:       "explicit config takes precedence over global config",
			cfg:        &config.AppConfig{InternalToken: "cfg-secret"},
			global:     &config.AppConfig{InternalToken: "global-secret"},
			headerName: internalTokenHeader,
			header:     "global-secret",
			want:       false,
		},
		{
			name:       "explicit config takes precedence and accepts its own token",
			cfg:        &config.AppConfig{InternalToken: "cfg-secret"},
			global:     &config.AppConfig{InternalToken: "global-secret"},
			headerName: internalTokenHeader,
			header:     "cfg-secret",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("INTERNAL_TOKEN", tt.env)

			old := config.GlobalConfig
			config.GlobalConfig = tt.global
			t.Cleanup(func() { config.GlobalConfig = old })

			if got := hasValidInternalTokenForConfig(newInternalTokenRequest(tt.headerName, tt.header), tt.cfg); got != tt.want {
				t.Errorf("hasValidInternalTokenForConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasValidInternalTokenFailsClosedOnNilRequest(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{InternalToken: "secret"}
	t.Cleanup(func() { config.GlobalConfig = old })

	if hasValidInternalTokenForConfig(nil, &config.AppConfig{InternalToken: "secret"}) {
		t.Error("nil request must not authenticate")
	}
	if hasValidInternalToken(nil) {
		t.Error("nil request must not authenticate via hasValidInternalToken")
	}
}

func TestHasValidInternalTokenFailsClosedWithNilConfig(t *testing.T) {
	old := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{InternalToken: "secret"}
	t.Cleanup(func() { config.GlobalConfig = old })

	if !hasValidInternalToken(newInternalTokenRequest(internalTokenHeader, "secret")) {
		t.Error("a configured token must still authenticate with a nil config")
	}
	if hasValidInternalToken(newInternalTokenRequest(internalTokenHeader, "wrong")) {
		t.Error("a wrong token must not authenticate with a nil config")
	}
}
