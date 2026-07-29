// Command license is the self-host license runmode of onion. Hosted deploys
// can run it as a lightweight status service; self-hosted deploys point it at
// LICENSE_KEY or LICENSE_FILE plus LICENSE_PUBLIC_KEY.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	platformconfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
	serverhttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/server/probe"
	"github.com/voxire/onion/internal/platform/server/profiling"
)

type licenseRunMode struct {
	cfg *platformconfig.AppConfig
	// db is the optional self_host_licenses lookup pool. When nil (DB not
	// configured/reachable) verification falls back to signature/expiry only.
	db *sql.DB
}

type verifyRequest struct {
	LicenseKey string `json:"license_key"`
	InstanceID string `json:"instance_id,omitempty"`
}

func loadDotEnv() {
	platformconfig.LoadDotEnv()
}

func (m licenseRunMode) UsesDatabaseHealthCheck() bool {
	// The DB is an optional enhancement (revocation/expiry overlay) — verify
	// still works on signature/expiry alone if it's down. Only assert a DB
	// health dependency when a pool was actually established.
	return m.db != nil
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

	report := license.VerifyKey(m.cfg, req.LicenseKey)
	report = m.applyDBRevocation(r.Context(), req.LicenseKey, report)
	writeJSON(w, http.StatusOK, report)
}

// applyDBRevocation overlays self_host_licenses revocation/expiry onto a
// signature-valid report so revoking a key in voxtooling takes effect. It is
// resilient: a signature-invalid report, an absent DB, an unknown key, or a DB
// error all leave the signature-based result untouched.
func (m licenseRunMode) applyDBRevocation(ctx context.Context, rawKey string, report license.StatusReport) license.StatusReport {
	if !report.Valid || m.db == nil {
		return report
	}

	record, err := license.LookupByKeyHash(ctx, m.db, license.HashKey(rawKey))
	if err != nil {
		if !errors.Is(err, license.ErrLicenseNotFound) {
			logger.Ctx(ctx).Warn("license DB lookup failed; falling back to signature result",
				logger.Error(err),
			)
		}
		return report
	}

	return license.ApplyRevocation(report, record, time.Now().UTC())
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

// dbConfigured reports whether enough DB connection info is present to attempt
// a connection. Mirrors dbclient's DATABASE_URL-or-DB_* resolution.
func dbConfigured(cfg *platformconfig.AppConfig) bool {
	if cfg != nil {
		return cfg.DatabaseURL != "" || (cfg.DBHost != "" && cfg.DBName != "" && cfg.DBUser != "")
	}
	return os.Getenv("DATABASE_URL") != "" ||
		(os.Getenv("DB_HOST") != "" && os.Getenv("DB_NAME") != "" && os.Getenv("DB_USER") != "")
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Global().Warn("failed to encode JSON response", logger.Error(err))
	}
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

	port := platformconfig.ListenPort("8082")

	logger.Ctx(ctx).Info("license runmode starting",
		logger.String("deployment_mode", cfg.DeploymentMode),
		logger.String("port", port),
	)

	// Connect the self_host_licenses pool so verification can honor revocation
	// and DB-side expiry. Optional and fail-soft: if the DB isn't configured or
	// is unreachable, verification falls back to signature/expiry only.
	var db *sql.DB
	if dbConfigured(cfg) {
		if connErr := dbclient.Connect(ctx); connErr != nil {
			logger.Ctx(ctx).Warn("license DB unavailable; revocation checks disabled",
				logger.Error(connErr),
			)
		} else {
			db = dbclient.Pool()
			defer dbclient.Close()
		}
	}
	if err := probe.Start(ctx, probe.Config{
		ServiceName: "license",
		Port:        probe.PortFromEnv("8888"),
		Stop:        cancel,
		ReadyChecks: []probe.Check{
			func(ctx context.Context) error {
				if db == nil {
					return nil
				}
				return db.PingContext(ctx)
			},
		},
	}); err != nil {
		logger.Global().Warn("probe server failed to start", logger.Error(err))
	}
	if err := profiling.Start(ctx, profiling.PortFromEnv("6060")); err != nil {
		logger.Global().Warn("pprof server failed to start", logger.Error(err))
	}

	if err := serverhttp.Start(ctx, licenseRunMode{cfg: cfg, db: db}, port); err != nil {
		logger.Global().Fatal("license server failed", logger.Error(err))
	}
}
