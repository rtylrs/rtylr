package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	platformconfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
	"github.com/voxire/onion/internal/platform/logger"
)

type licenseAuthorityServer struct {
	cfg   *platformconfig.AppConfig
	db    *sql.DB
	clock *license.ClockGuard
}

type licenseVerifyRequest struct {
	LicenseKey string `json:"license_key"`
	InstanceID string `json:"instance_id,omitempty"`
}

// RegisterLicenseAuthorityRoutes registers the self-host license authority endpoints.
// These allow self-hosted backends to verify and activate their license keys against
// this authority service.
func RegisterLicenseAuthorityRoutes(r chi.Router, cfg *platformconfig.AppConfig, db *sql.DB) {
	s := &licenseAuthorityServer{
		cfg:   cfg,
		db:    db,
		clock: license.NewClockGuard(cfg.LicenseClockStateFile),
	}

	r.Get("/license/status", s.status)
	r.Post("/license/verify", s.verify)
	r.Post("/license/activate", s.activate)
	r.Post("/license/refresh", s.refresh)
}

func (s *licenseAuthorityServer) status(w http.ResponseWriter, _ *http.Request) {
	writeLicenseJSON(w, http.StatusOK, license.VerifyConfiguredAt(s.cfg, s.now()))
}

func (s *licenseAuthorityServer) refresh(w http.ResponseWriter, _ *http.Request) {
	writeLicenseJSON(w, http.StatusOK, license.VerifyConfiguredAt(s.cfg, s.now()))
}

func (s *licenseAuthorityServer) verify(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeLicenseVerifyRequest(w, r)
	if !ok {
		return
	}

	if strings.TrimSpace(req.LicenseKey) == "" {
		writeLicenseJSON(w, http.StatusOK, license.VerifyConfiguredAt(s.cfg, s.now()))
		return
	}

	report := license.VerifyKeyAt(s.cfg, req.LicenseKey, s.now())
	report = s.applyDBRevocation(r.Context(), req.LicenseKey, report)
	writeLicenseJSON(w, http.StatusOK, report)
}

func (s *licenseAuthorityServer) activate(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeLicenseVerifyRequest(w, r)
	if !ok {
		return
	}
	if strings.TrimSpace(req.InstanceID) == "" {
		writeLicenseJSON(w, http.StatusBadRequest, map[string]string{"error": license.ErrInstanceIDMissing.Error()})
		return
	}
	rawKey := strings.TrimSpace(req.LicenseKey)
	if rawKey == "" {
		rawKey, _ = licenseKeyFromAuthorityConfig(s.cfg)
	}
	report := license.VerifyKeyAt(s.cfg, rawKey, s.now())
	report = s.applyDBRevocation(r.Context(), rawKey, report)
	if !report.Valid {
		writeLicenseJSON(w, http.StatusPaymentRequired, report)
		return
	}
	if err := license.Activate(r.Context(), s.db, license.HashKey(rawKey), req.InstanceID, report.SeatLimit); err != nil {
		if errors.Is(err, license.ErrSeatLimitExceeded) {
			writeLicenseJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		logger.Ctx(r.Context()).Warn("license activation failed", logger.Error(err))
		writeLicenseJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "license activation unavailable"})
		return
	}
	writeLicenseJSON(w, http.StatusOK, report)
}

func (s *licenseAuthorityServer) now() time.Time {
	if s.clock != nil {
		return s.clock.Now()
	}
	return time.Now().UTC()
}

func licenseKeyFromAuthorityConfig(cfg *platformconfig.AppConfig) (string, error) {
	if cfg == nil {
		return "", license.ErrMissingLicense
	}
	if strings.TrimSpace(cfg.LicenseKey) != "" {
		return strings.TrimSpace(cfg.LicenseKey), nil
	}
	if strings.TrimSpace(cfg.LicenseFile) != "" {
		contents, err := os.ReadFile(cfg.LicenseFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(contents)), nil
	}
	return "", license.ErrMissingLicense
}

func (s *licenseAuthorityServer) applyDBRevocation(ctx context.Context, rawKey string, report license.StatusReport) license.StatusReport {
	if !report.Valid {
		return report
	}
	if s.db == nil {
		return licenseAuthorityDBUnavailable(report, "license authority database is unavailable")
	}

	record, err := license.LookupByKeyHash(ctx, s.db, license.HashKey(rawKey))
	if err != nil {
		if errors.Is(err, license.ErrLicenseNotFound) {
			report.Valid = false
			report.Status = license.StatusInvalid
			report.Reason = "license is not registered with the authority"
			return report
		}
		logger.Ctx(ctx).Warn("license DB lookup failed", logger.Error(err))
		return licenseAuthorityDBUnavailable(report, "license authority lookup failed")
	}

	return license.ApplyRevocation(report, record, time.Now().UTC())
}

func licenseAuthorityDBUnavailable(report license.StatusReport, reason string) license.StatusReport {
	report.Valid = false
	report.Status = license.StatusUnreachable
	report.Reason = reason
	return report
}

func decodeLicenseVerifyRequest(w http.ResponseWriter, r *http.Request) (licenseVerifyRequest, bool) {
	if r.Body == nil || r.Body == http.NoBody {
		return licenseVerifyRequest{}, true
	}
	defer r.Body.Close()

	var req licenseVerifyRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
	if err := decoder.Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeLicenseJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request body"})
		return licenseVerifyRequest{}, false
	}
	return req, true
}

func writeLicenseJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Global().Warn("failed to encode JSON response", logger.Error(err))
	}
}
