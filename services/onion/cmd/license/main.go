// Command license is the self-host license runmode of onion. Hosted deploys
// can run it as a lightweight status service; self-hosted deploys point it at
// LICENSE_KEY or LICENSE_FILE plus LICENSE_PUBLIC_KEY.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"

	platformconfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	serverhttp "github.com/voxire/onion/internal/platform/server/http"
)

type licenseRunMode struct {
	cfg *platformconfig.AppConfig
}

type verifyRequest struct {
	LicenseKey string `json:"license_key"`
	InstanceID string `json:"instance_id,omitempty"`
}

func loadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func (m licenseRunMode) UsesDatabaseHealthCheck() bool {
	return false
}

func (m licenseRunMode) Setup(_ context.Context, r *chi.Mux) {
	r.Get("/license/status", m.status)
	r.Post("/license/verify", m.verify)
	r.Post("/license/activate", m.verify)
	r.Post("/license/refresh", m.refresh)
}

func (m licenseRunMode) status(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, license.VerifyConfigured(m.cfg))
}

func (m licenseRunMode) refresh(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, license.VerifyConfigured(m.cfg))
}

func (m licenseRunMode) verify(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeVerifyRequest(w, r)
	if !ok {
		return
	}

	if strings.TrimSpace(req.LicenseKey) == "" {
		writeJSON(w, http.StatusOK, license.VerifyConfigured(m.cfg))
		return
	}

	writeJSON(w, http.StatusOK, license.VerifyKey(m.cfg, req.LicenseKey))
}

func decodeVerifyRequest(w http.ResponseWriter, r *http.Request) (verifyRequest, bool) {
	if r.Body == nil || r.Body == http.NoBody {
		return verifyRequest{}, true
	}
	defer r.Body.Close()

	var req verifyRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
	if err := decoder.Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request body"})
		return verifyRequest{}, false
	}
	return req, true
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	if os.Getenv("RUN_MODE") == "" {
		_ = os.Setenv("RUN_MODE", "license")
	}
	if os.Getenv("SERVICE_NAME") == "" {
		_ = os.Setenv("SERVICE_NAME", "license")
	}

	loadDotEnv()
	logger.SetupLogger()
	defer logger.Flush()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := platformconfig.LoadAndValidate()
	if err != nil {
		logger.Global().Fatal("config validation failed", logger.Error(err))
	}

	flushSentry := observability.InitSentry(ctx)
	defer flushSentry()

	port := cfg.Port
	if os.Getenv("PORT") == "" {
		port = "8082"
	}

	logger.Ctx(ctx).Info("license runmode starting",
		logger.String("deployment_mode", cfg.DeploymentMode),
		logger.String("port", port),
	)

	if err := serverhttp.Start(ctx, licenseRunMode{cfg: cfg}, port); err != nil {
		logger.Global().Fatal("license server failed", logger.Error(err))
	}
}
