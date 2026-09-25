package aicache

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	CacheVersion = "exact-v5"

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
	cacheNamespace     = "ai_response_exact"
	analyticsNamespace = "ai_response_analytics"
	analyticsEventsID  = "events"
	defaultExactTTL    = 30 * time.Minute
	minExactTTL        = time.Minute
	maxExactTTL        = 24 * time.Hour
	analyticsEventTTL  = 30 * 24 * time.Hour
	maxAnalyticsEvents = 5000
	maxRecentEvents    = 100
	analyticsSourceDB  = "mysql"
	analyticsSourceKV  = "shared_cache"
	analyticsSourceMem = "process"
	defaultModeReason  = "exact cache is enabled for tenant-scoped replay"
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
	DataFreshness string
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

type analyticsEventRecord struct {
	ID    string `json:"id"`
	Event Event  `json:"event"`
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
	AnalyticsSource   string         `json:"analytics_source"`
	AnalyticsDurable  bool           `json:"analytics_durable"`
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

var (
	collector                  = newCollector()
	persistentAnalyticsEnabled = newPersistentAnalyticsFlag()
	analyticsPersistQueue      = make(chan analyticsEventRecord, 512)
	analyticsPersistOnce       sync.Once
)

func newPersistentAnalyticsFlag() *atomic.Bool {
	enabled := &atomic.Bool{}
	enabled.Store(true)
	return enabled
}

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
