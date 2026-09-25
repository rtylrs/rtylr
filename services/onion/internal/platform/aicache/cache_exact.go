package aicache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

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
	if err != nil || ttl < minExactTTL || ttl > maxExactTTL {
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
		DataFreshness string `json:"data_freshness"`
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
		DataFreshness: normalizeText(parts.DataFreshness),
		Prompt:        normalizeText(parts.Prompt),
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		// Marshalling a struct of plain strings and ints cannot realistically
		// fail; fall back to a deterministic textual encoding so callers still
		// get a stable cache key.
		logger.Global().Warn("failed to marshal exact cache key payload", logger.Error(err))
		bytes = []byte(fmt.Sprintf("%#v", payload))
	}
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
	if strings.TrimSpace(key) == "" {
		return
	}
	if ttl < minExactTTL || ttl > maxExactTTL {
		ttl = defaultExactTTL
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	exactStore().Put(ctx, key, entry, cacher.WithExpiry(ctx, ttl))
}

func Record(event Event) {
	event = normalizeEvent(event)

	collector.mu.Lock()
	collector.record(event)
	collector.mu.Unlock()

	persistAnalyticsEvent(event)
}

func Snapshot() AnalyticsSnapshot {
	if snapshot, ok := persistentSnapshot(); ok {
		return snapshot
	}

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
		AnalyticsSource:   analyticsSourceMem,
		AnalyticsDurable:  false,
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
	persistentAnalyticsEnabled.Store(false)
}

func normalizeEvent(event Event) Event {
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	event.Mode = normalizeMode(event.Mode)
	event.Outcome = strings.TrimSpace(event.Outcome)
	if event.Outcome == "" {
		event.Outcome = OutcomeError
	}
	if event.PromptChars < 0 {
		event.PromptChars = 0
	}
	if event.ResponseChars < 0 {
		event.ResponseChars = 0
	}
	if event.ProviderTokens < 0 {
		event.ProviderTokens = 0
	}
	if event.EstimatedSavings < 0 {
		event.EstimatedSavings = 0
	}
	if event.LatencyMs < 0 {
		event.LatencyMs = 0
	}
	return event
}

func (c *collectorState) record(event Event) {
	c.outcomes[event.Outcome]++
	if event.Mode != "" {
		c.modes[event.Mode]++
	}

	switch event.Outcome {
	case OutcomeHit:
		c.totals.Lookups++
		c.totals.Hits++
		c.totals.TokensSaved += int64(event.EstimatedSavings)
	case OutcomeMiss:
		c.totals.Lookups++
		c.totals.Misses++
	case OutcomeSkip:
		c.totals.Lookups++
		c.totals.Skips++
	case OutcomeWrite:
		c.totals.Writes++
		c.totals.ProviderTokens += int64(event.ProviderTokens)
		c.totals.ResponseBytes += int64(event.ResponseChars)
	case OutcomeError:
		c.totals.Errors++
	}

	if event.LatencyMs > 0 {
		c.totals.TotalLatencyMs += event.LatencyMs
		c.totals.LatencySamples++
	}

	c.recent = append([]Event{event}, c.recent...)
	if len(c.recent) > maxRecentEvents {
		c.recent = c.recent[:maxRecentEvents]
	}
}
