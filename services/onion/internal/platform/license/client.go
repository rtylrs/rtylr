package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	platformconfig "github.com/voxire/onion/internal/platform/config"
)

// StatusUnreachable means the license authority could not be reached and no
// cached-valid result remains within the grace window.
const StatusUnreachable = "unreachable"

// refreshInterval is how often a healthy self-host backend re-checks the
// authority. Between checks the last result is served from cache.
const refreshInterval = time.Hour

const maxOfflineGrace = 24 * time.Hour

// Client talks to the hosted license authority (license.voxire.com).
type Client struct {
	authorityURL  string
	internalToken string
	httpClient    *http.Client
}

// NewClientWithInternalToken configures the credential used by the dedicated
// license authority. It is kept separate from NewClient so callers that only
// need a public verifier retain the existing constructor contract.
func NewClientWithInternalToken(authorityURL, internalToken string) *Client {
	authorityURL = strings.TrimRight(strings.TrimSpace(authorityURL), "/")
	if authorityURL == "" {
		authorityURL = platformconfig.DefaultLicenseAuthorityURL
	}
	return &Client{
		authorityURL:  authorityURL,
		internalToken: strings.TrimSpace(internalToken),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Verify POSTs the license key to {authority}/license/verify and returns the
// authority's StatusReport. A non-nil error means the authority was unreachable
// or returned an unusable response — callers apply grace logic on error.
func (c *Client) Verify(ctx context.Context, key string) (StatusReport, error) {
	body, err := json.Marshal(map[string]string{"license_key": strings.TrimSpace(key)})
	if err != nil {
		return StatusReport{}, fmt.Errorf("encode license verify request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.authorityURL+"/license/verify", bytes.NewReader(body))
	if err != nil {
		return StatusReport{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.internalToken != "" {
		req.Header.Set("Internal-Token", c.internalToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return StatusReport{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return StatusReport{}, fmt.Errorf("license authority returned %d", resp.StatusCode)
	}

	var report StatusReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return StatusReport{}, fmt.Errorf("decode authority response: %w", err)
	}
	return report, nil
}

// CachedValidator resolves the effective license status for a self-host backend.
// It periodically re-checks the authority and, when the authority is briefly
// unreachable, keeps serving the last valid result for the configured grace
// window. In hosted mode it is a no-op that always reports valid/hosted.
type CachedValidator struct {
	cfg    *platformconfig.AppConfig
	client *Client
	key    string
	grace  time.Duration

	mu                     sync.Mutex
	last                   StatusReport
	lastSuccess            time.Time // last time the authority returned a valid report
	lastCheck              time.Time
	lastAttemptUnreachable bool // whether the most recent authority attempt failed
	haveResult             bool
	clock                  *ClockGuard
}

func NewCachedValidator(cfg *platformconfig.AppConfig) *CachedValidator {
	grace := gracePeriod(cfg)
	if grace <= 0 {
		grace = 24 * time.Hour
	}
	if grace > maxOfflineGrace {
		grace = maxOfflineGrace
	}
	key, _ := configuredKey(cfg)

	// The configured authority, not the hardcoded default. #1079 restored
	// reading LICENSE_AUTHORITY_URL into the config and this is the only
	// construction path for the client, so without this the setting reached
	// nothing and an air-gapped self-host install could not point at its own
	// authority.
	//
	// Safe to honour here rather than pinning: Current and Refresh both return
	// hostedReport unless the deployment is self-hosted, so a hosted install
	// never contacts an authority at all and this value is only ever used by an
	// operator configuring their own. Authenticating the authority's reply is a
	// separate and still-open concern, tracked in #1091; pinning the URL was
	// never a substitute for it, since an operator who controls the host can
	// redirect the hosted name just as easily.
	//
	// Config defaults this to DefaultLicenseAuthorityURL and NewClient falls
	// back to the same value for an empty string, so a deployment that sets
	// nothing is unchanged.
	authority := ""
	internalToken := ""
	if cfg != nil {
		authority = cfg.LicenseAuthorityURL
		internalToken = cfg.InternalToken
	}

	return &CachedValidator{
		cfg:    cfg,
		client: NewClientWithInternalToken(authority, internalToken),
		key:    strings.TrimSpace(key),
		grace:  grace,
		clock:  NewClockGuard(clockStateFile(cfg)),
	}
}

// Current returns the effective license status, refreshing from the authority
// at most once per refreshInterval. Hosted mode always returns a valid report.
func (v *CachedValidator) Current(ctx context.Context) StatusReport {
	now := v.now()
	if !IsSelfHosted(v.cfg) {
		return hostedReport(v.cfg, now)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.haveResult && now.Sub(v.lastCheck) < refreshInterval {
		return v.effectiveLocked(now)
	}
	return v.refreshLocked(ctx, now)
}

// Refresh forces an authority re-check (e.g. at boot or on a schedule).
func (v *CachedValidator) Refresh(ctx context.Context) StatusReport {
	now := v.now()
	if !IsSelfHosted(v.cfg) {
		return hostedReport(v.cfg, now)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.refreshLocked(ctx, now)
}

func clockStateFile(cfg *platformconfig.AppConfig) string {
	if cfg != nil {
		return cfg.LicenseClockStateFile
	}
	return "/var/lib/rtylr/license-clock.json"
}

func (v *CachedValidator) now() time.Time {
	if v.clock != nil {
		return v.clock.Now()
	}
	return time.Now().UTC()
}

func (v *CachedValidator) refreshLocked(ctx context.Context, now time.Time) StatusReport {
	mode := deploymentMode(v.cfg)
	if v.key == "" {
		v.last = StatusReport{Mode: mode, Valid: false, Status: StatusMissing, Reason: ErrMissingLicense.Error(), CheckedAt: now}
		v.lastCheck = now
		v.lastAttemptUnreachable = false
		v.haveResult = true
		return v.last
	}

	report, err := v.client.Verify(ctx, v.key)
	v.lastCheck = now
	if err == nil {
		report.Mode = mode
		report.CheckedAt = now
		v.last = report
		v.lastAttemptUnreachable = false
		v.haveResult = true
		if report.Valid {
			v.lastSuccess = now
		}
		return report
	}

	// Authority unreachable — serve the last valid result within grace.
	v.lastAttemptUnreachable = true
	v.haveResult = true
	return v.effectiveLocked(now)
}

// effectiveLocked returns the latest authority answer, or applies grace when
// the most recent authority attempt was unreachable. Caller holds v.mu.
func (v *CachedValidator) effectiveLocked(now time.Time) StatusReport {
	mode := deploymentMode(v.cfg)
	if v.haveResult && !v.lastAttemptUnreachable {
		cached := v.last
		cached.Mode = mode
		cached.CheckedAt = now
		return cached
	}
	if v.last.Valid && !v.lastSuccess.IsZero() && now.Before(v.lastSuccess.Add(v.grace)) {
		degraded := v.last
		degraded.Mode = mode
		degraded.CheckedAt = now
		if degraded.Reason == "" {
			degraded.Reason = "license authority unreachable; within grace window"
		}
		return degraded
	}
	return StatusReport{
		Mode:      mode,
		Valid:     false,
		Status:    StatusUnreachable,
		Reason:    "license authority unreachable and grace window elapsed",
		CheckedAt: now,
	}
}
