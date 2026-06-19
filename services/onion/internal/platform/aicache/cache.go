package aicache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const (
	CacheVersion = "exact-v2"

	OutcomeHit   = "hit"
	OutcomeMiss  = "miss"
	OutcomeSkip  = "skip"
	OutcomeWrite = "write"
	OutcomeError = "error"

	ModeAsk     = "ask"
	ModeAgent   = "agent"
	ModeAnalyst = "analyst"
	ModeSupport = "support"
)

const (
	cacheNamespace    = "ai_response_exact"
	defaultExactTTL   = 30 * time.Minute
	maxRecentEvents   = 100
	defaultModeReason = "exact cache is enabled for tenant-scoped replay"
)

type ExactPolicy struct {
	Enabled bool          `json:"enabled"`
	Mode    string        `json:"mode"`
	Reason  string        `json:"reason"`
	TTL     time.Duration `json:"-"`
	TTLSec  int64         `json:"ttl_seconds"`
}

type PolicyStatus struct {
	Mode       string `json:"mode"`
	Enabled    bool   `json:"enabled"`
	Reason     string `json:"reason"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

type ExactKeyParts struct {
	BusinessID    int64
	UserID        int64
	Mode          string
	Provider      string
	Model         string
	RequestDate   string
	SystemContext string
	Prompt        string
}

type ExactResponse struct {
	Response         string    `json:"response"`
	Provider         string    `json:"provider"`
	Mode             string    `json:"mode"`
	PromptChars      int       `json:"prompt_chars"`
	ResponseChars    int       `json:"response_chars"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

func (e ExactResponse) TotalTokens() int {
	total := e.PromptTokens + e.CompletionTokens
	if total > 0 {
		return total
	}
	return estimateTokens(e.PromptChars) + estimateTokens(e.ResponseChars)
}

type Event struct {
	At               time.Time `json:"at"`
	Outcome          string    `json:"outcome"`
	Mode             string    `json:"mode"`
	Provider         string    `json:"provider,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	CacheKeyPreview  string    `json:"cache_key_preview,omitempty"`
	BusinessID       int64     `json:"business_id,omitempty"`
	UserID           int64     `json:"user_id,omitempty"`
	PromptChars      int       `json:"prompt_chars,omitempty"`
	ResponseChars    int       `json:"response_chars,omitempty"`
	ProviderTokens   int       `json:"provider_tokens,omitempty"`
	EstimatedSavings int       `json:"estimated_savings_tokens,omitempty"`
	LatencyMs        int64     `json:"latency_ms,omitempty"`
}

type Totals struct {
	Lookups        int64 `json:"lookups"`
	Hits           int64 `json:"hits"`
	Misses         int64 `json:"misses"`
	Skips          int64 `json:"skips"`
	Writes         int64 `json:"writes"`
	Errors         int64 `json:"errors"`
	ProviderTokens int64 `json:"provider_tokens"`
	TokensSaved    int64 `json:"tokens_saved"`
	ResponseBytes  int64 `json:"response_bytes"`
	TotalLatencyMs int64 `json:"total_latency_ms"`
	LatencySamples int64 `json:"latency_samples"`
}

type Breakdown struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type AnalyticsSnapshot struct {
	GeneratedAt       time.Time      `json:"generated_at"`
	WindowStartedAt   time.Time      `json:"window_started_at"`
	CacheVersion      string         `json:"cache_version"`
	ExactCacheEnabled bool           `json:"exact_cache_enabled"`
	HitRate           float64        `json:"hit_rate"`
	AverageLatencyMs  int64          `json:"average_latency_ms"`
	Totals            Totals         `json:"totals"`
	Outcomes          []Breakdown    `json:"outcomes"`
	Modes             []Breakdown    `json:"modes"`
	Policies          []PolicyStatus `json:"policies"`
	Recent            []Event        `json:"recent"`
}

type collectorState struct {
	mu       sync.Mutex
	started  time.Time
	totals   Totals
	outcomes map[string]int64
	modes    map[string]int64
	recent   []Event
}

var collector = newCollector()

func newCollector() *collectorState {
	return &collectorState{
		started:  time.Now().UTC(),
		outcomes: make(map[string]int64),
		modes:    make(map[string]int64),
		recent:   make([]Event, 0, maxRecentEvents),
	}
}

func ExactPolicyForMode(mode string) ExactPolicy {
	normalized := normalizeMode(mode)
	ttl := ExactTTL()
	if !ExactCacheEnabled() {
		return ExactPolicy{Enabled: false, Mode: normalized, Reason: "exact cache disabled by AI_RESPONSE_CACHE_ENABLED", TTL: ttl, TTLSec: int64(ttl.Seconds())}
	}

	switch normalized {
	case ModeAgent:
		return ExactPolicy{Enabled: false, Mode: normalized, Reason: "agent mode may mutate data or call tools", TTL: ttl, TTLSec: int64(ttl.Seconds())}
	case ModeAnalyst:
		return ExactPolicy{Enabled: false, Mode: normalized, Reason: "analyst mode depends on live business data", TTL: ttl, TTLSec: int64(ttl.Seconds())}
	default:
		return ExactPolicy{Enabled: true, Mode: normalized, Reason: defaultModeReason, TTL: ttl, TTLSec: int64(ttl.Seconds())}
	}
}

func PolicySnapshot() []PolicyStatus {
	modes := []string{ModeAsk, ModeSupport, ModeAnalyst, ModeAgent}
	out := make([]PolicyStatus, 0, len(modes))
	for _, mode := range modes {
		p := ExactPolicyForMode(mode)
		out = append(out, PolicyStatus{
			Mode:       p.Mode,
			Enabled:    p.Enabled,
			Reason:     p.Reason,
			TTLSeconds: p.TTLSec,
		})
	}
	return out
}

func ExactCacheEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AI_RESPONSE_CACHE_ENABLED"))) {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

func ExactTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("AI_RESPONSE_CACHE_TTL"))
	if raw == "" {
		return defaultExactTTL
	}
	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl <= 0 {
		return defaultExactTTL
	}
	return ttl
}

func ExactKey(parts ExactKeyParts) string {
	payload := struct {
		Version       string `json:"version"`
		BusinessID    int64  `json:"business_id"`
		UserID        int64  `json:"user_id"`
		Mode          string `json:"mode"`
		Provider      string `json:"provider"`
		Model         string `json:"model"`
		RequestDate   string `json:"request_date"`
		SystemContext string `json:"system_context"`
		Prompt        string `json:"prompt"`
	}{
		Version:       CacheVersion,
		BusinessID:    parts.BusinessID,
		UserID:        parts.UserID,
		Mode:          normalizeMode(parts.Mode),
		Provider:      normalizeText(parts.Provider),
		Model:         normalizeText(parts.Model),
		RequestDate:   normalizeText(parts.RequestDate),
		SystemContext: normalizeText(parts.SystemContext),
		Prompt:        normalizeText(parts.Prompt),
	}

	bytes, _ := json.Marshal(payload)
	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:])
}

func normalizeText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func KeyPreview(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:12]
}

func GetExact(ctx context.Context, key string) (ExactResponse, bool) {
	return exactStore().Get(ctx, key)
}

func PutExact(ctx context.Context, key string, entry ExactResponse, ttl time.Duration) {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	exactStore().Put(ctx, key, entry, cacher.WithExpiry(ctx, ttl))
}

func Record(event Event) {
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	event.Mode = normalizeMode(event.Mode)
	event.Outcome = strings.TrimSpace(event.Outcome)
	if event.Outcome == "" {
		event.Outcome = OutcomeError
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()

	collector.outcomes[event.Outcome]++
	if event.Mode != "" {
		collector.modes[event.Mode]++
	}

	switch event.Outcome {
	case OutcomeHit:
		collector.totals.Lookups++
		collector.totals.Hits++
		collector.totals.TokensSaved += int64(event.EstimatedSavings)
	case OutcomeMiss:
		collector.totals.Lookups++
		collector.totals.Misses++
	case OutcomeSkip:
		collector.totals.Lookups++
		collector.totals.Skips++
	case OutcomeWrite:
		collector.totals.Writes++
		collector.totals.ProviderTokens += int64(event.ProviderTokens)
		collector.totals.ResponseBytes += int64(event.ResponseChars)
	case OutcomeError:
		collector.totals.Errors++
	}

	if event.LatencyMs > 0 {
		collector.totals.TotalLatencyMs += event.LatencyMs
		collector.totals.LatencySamples++
	}

	collector.recent = append([]Event{event}, collector.recent...)
	if len(collector.recent) > maxRecentEvents {
		collector.recent = collector.recent[:maxRecentEvents]
	}
}

func Snapshot() AnalyticsSnapshot {
	collector.mu.Lock()
	defer collector.mu.Unlock()

	recent := make([]Event, len(collector.recent))
	copy(recent, collector.recent)
	totals := collector.totals
	outcomes := breakdowns(collector.outcomes)
	modes := breakdowns(collector.modes)

	var hitRate float64
	hitMiss := totals.Hits + totals.Misses
	if hitMiss > 0 {
		hitRate = float64(totals.Hits) / float64(hitMiss)
	}

	var avgLatency int64
	if totals.LatencySamples > 0 {
		avgLatency = totals.TotalLatencyMs / totals.LatencySamples
	}

	return AnalyticsSnapshot{
		GeneratedAt:       time.Now().UTC(),
		WindowStartedAt:   collector.started,
		CacheVersion:      CacheVersion,
		ExactCacheEnabled: ExactCacheEnabled(),
		HitRate:           hitRate,
		AverageLatencyMs:  avgLatency,
		Totals:            totals,
		Outcomes:          outcomes,
		Modes:             modes,
		Policies:          PolicySnapshot(),
		Recent:            recent,
	}
}

func ResetForTest() {
	collector = newCollector()
}

func exactStore() cacher.Cache[ExactResponse] {
	return cacher.OfWithNamespace[ExactResponse](cacheNamespace)
}

func normalizeMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ModeAgent:
		return ModeAgent
	case ModeAnalyst:
		return ModeAnalyst
	case ModeSupport:
		return ModeSupport
	default:
		return ModeAsk
	}
}

func breakdowns(counts map[string]int64) []Breakdown {
	out := make([]Breakdown, 0, len(counts))
	for key, count := range counts {
		out = append(out, Breakdown{Key: key, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Key < out[j].Key
		}
		return out[i].Count > out[j].Count
	})
	return out
}

func estimateTokens(chars int) int {
	if chars <= 0 {
		return 0
	}
	return chars / 4
}
