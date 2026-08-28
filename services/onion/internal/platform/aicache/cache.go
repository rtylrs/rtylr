package aicache

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	aicachesqlc "github.com/voxire/onion/internal/platform/aicache/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
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

func persistAnalyticsEvent(event Event) {
	if !persistentAnalyticsEnabled.Load() {
		return
	}

	record := analyticsEventRecord{
		ID:    analyticsEventID(event),
		Event: event,
	}
	analyticsPersistOnce.Do(func() {
		go analyticsPersistenceWorker()
	})
	select {
	case analyticsPersistQueue <- record:
	default:
		// The in-process collector is authoritative for the request. A full
		// durable queue must never add latency to an AI response.
		logger.Global().Warn("AI cache analytics persistence queue is full; event retained in process only")
	}
}

func analyticsPersistenceWorker() {
	for record := range analyticsPersistQueue {
		persistAnalyticsEventToDB(record)
		persistAnalyticsEventToSharedCache(record)
	}
}

func persistAnalyticsEventToSharedCache(record analyticsEventRecord) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	store := analyticsEventsStore()
	score := float64(record.Event.At.UnixMilli())
	store.SortedSetAdd(ctx, analyticsEventsID, record, score, cacher.WithExpiry(ctx, analyticsEventTTL))
	cutoff := time.Now().UTC().Add(-analyticsEventTTL).UnixMilli()
	store.SortedSetRemoveRangeUpToScore(ctx, analyticsEventsID, float64(cutoff))
	store.SortedSetRemoveRangeByRank(ctx, analyticsEventsID, 0, -int64(maxAnalyticsEvents+1))
}

func persistAnalyticsEventToDB(record analyticsEventRecord) {
	db := dbclient.Pool()
	if db == nil {
		return
	}

	event := normalizeEvent(record.Event)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	queries := aicachesqlc.New(db)
	err := queries.InsertAICacheEvent(ctx, aicachesqlc.InsertAICacheEventParams{
		EventID:                record.ID,
		OccurredAt:             event.At.UTC(),
		Outcome:                event.Outcome,
		Mode:                   event.Mode,
		Provider:               nullSQLString(event.Provider),
		Reason:                 nullSQLString(event.Reason),
		CacheKeyPreview:        nullSQLString(event.CacheKeyPreview),
		BusinessID:             nullSQLInt64(event.BusinessID),
		UserID:                 nullSQLInt64(event.UserID),
		PromptChars:            nullSQLInt32(event.PromptChars),
		ResponseChars:          nullSQLInt32(event.ResponseChars),
		ProviderTokens:         nullSQLInt32(event.ProviderTokens),
		EstimatedSavingsTokens: nullSQLInt32(event.EstimatedSavings),
		LatencyMs:              nullSQLInt64(event.LatencyMs),
	})
	if err != nil {
		logger.Global().Debug("could not persist AI cache analytics event", logger.Error(err))
	}
}

func persistentSnapshot() (AnalyticsSnapshot, bool) {
	if !persistentAnalyticsEnabled.Load() {
		return AnalyticsSnapshot{}, false
	}

	if snapshot, ok := persistentDatabaseSnapshot(); ok {
		return snapshot, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	records := analyticsEventsStore().SortedSetGetAll(ctx, analyticsEventsID)
	if len(records) == 0 {
		return AnalyticsSnapshot{}, false
	}

	cutoff := time.Now().UTC().Add(-analyticsEventTTL)
	events := make([]Event, 0, len(records))
	for _, record := range records {
		event := record.Event
		if event.At.IsZero() || event.Outcome == "" {
			continue
		}
		if event.At.Before(cutoff) {
			continue
		}
		events = append(events, normalizeEvent(event))
	}
	if len(events) == 0 {
		return AnalyticsSnapshot{}, false
	}

	return withAnalyticsSource(snapshotFromEvents(events), analyticsSourceKV, true), true
}

func persistentDatabaseSnapshot() (AnalyticsSnapshot, bool) {
	db := dbclient.Pool()
	if db == nil {
		return AnalyticsSnapshot{}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()

	rows, err := aicachesqlc.New(db).ListRecentAICacheEvents(ctx, aicachesqlc.ListRecentAICacheEventsParams{
		OccurredAt: time.Now().UTC().Add(-analyticsEventTTL),
		Limit:      int32(maxAnalyticsEvents),
	})
	if err != nil {
		logger.Global().Debug("could not load AI cache analytics events", logger.Error(err))
		return AnalyticsSnapshot{}, false
	}

	events := make([]Event, 0, maxAnalyticsEvents)
	for _, row := range rows {
		event, scanErr := eventFromAICacheRow(row)
		if scanErr != nil {
			logger.Global().Debug("could not scan AI cache analytics event", logger.Error(scanErr))
			continue
		}
		events = append(events, normalizeEvent(event))
	}
	if len(events) == 0 {
		return AnalyticsSnapshot{}, false
	}

	return withAnalyticsSource(snapshotFromEvents(events), analyticsSourceDB, true), true
}

func eventFromAICacheRow(row aicachesqlc.ListRecentAICacheEventsRow) (Event, error) {
	event := Event{
		At:               row.OccurredAt,
		Outcome:          row.Outcome,
		Mode:             row.Mode,
		Provider:         row.Provider.String,
		Reason:           row.Reason.String,
		CacheKeyPreview:  row.CacheKeyPreview.String,
		BusinessID:       row.BusinessID.Int64,
		UserID:           row.UserID.Int64,
		PromptChars:      int(row.PromptChars.Int32),
		ResponseChars:    int(row.ResponseChars.Int32),
		ProviderTokens:   int(row.ProviderTokens.Int32),
		EstimatedSavings: int(row.EstimatedSavingsTokens.Int32),
		LatencyMs:        row.LatencyMs.Int64,
	}
	return event, nil
}

func nullSQLString(value string) sql.NullString {
	if strings.TrimSpace(value) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullSQLInt32(value int) sql.NullInt32 {
	if value == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(value), Valid: true} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func nullSQLInt64(value int64) sql.NullInt64 {
	if value == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: value, Valid: true}
}

func snapshotFromEvents(events []Event) AnalyticsSnapshot {
	if len(events) == 0 {
		return AnalyticsSnapshot{
			GeneratedAt:       time.Now().UTC(),
			CacheVersion:      CacheVersion,
			ExactCacheEnabled: ExactCacheEnabled(),
			AnalyticsSource:   analyticsSourceMem,
			Policies:          PolicySnapshot(),
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].At.After(events[j].At)
	})

	state := newCollector()
	state.started = events[len(events)-1].At
	for i := len(events) - 1; i >= 0; i-- {
		state.record(normalizeEvent(events[i]))
	}

	recent := make([]Event, len(state.recent))
	copy(recent, state.recent)

	hitMiss := state.totals.Hits + state.totals.Misses
	var hitRate float64
	if hitMiss > 0 {
		hitRate = float64(state.totals.Hits) / float64(hitMiss)
	}

	var avgLatency int64
	if state.totals.LatencySamples > 0 {
		avgLatency = state.totals.TotalLatencyMs / state.totals.LatencySamples
	}

	return AnalyticsSnapshot{
		GeneratedAt:       time.Now().UTC(),
		WindowStartedAt:   state.started,
		CacheVersion:      CacheVersion,
		ExactCacheEnabled: ExactCacheEnabled(),
		AnalyticsSource:   analyticsSourceMem,
		AnalyticsDurable:  false,
		HitRate:           hitRate,
		AverageLatencyMs:  avgLatency,
		Totals:            state.totals,
		Outcomes:          breakdowns(state.outcomes),
		Modes:             breakdowns(state.modes),
		Policies:          PolicySnapshot(),
		Recent:            recent,
	}
}

func withAnalyticsSource(snapshot AnalyticsSnapshot, source string, durable bool) AnalyticsSnapshot {
	snapshot.AnalyticsSource = source
	snapshot.AnalyticsDurable = durable
	return snapshot
}

func analyticsEventID(event Event) string {
	var entropy [8]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return fmt.Sprintf("%020d-%s-%s", event.At.UnixNano(), event.Outcome, event.CacheKeyPreview)
	}
	return fmt.Sprintf("%020d-%s", event.At.UnixNano(), hex.EncodeToString(entropy[:]))
}

func exactStore() cacher.Cache[ExactResponse] {
	return cacher.OfWithNamespace[ExactResponse](cacheNamespace)
}

func analyticsEventsStore() cacher.Cache[analyticsEventRecord] {
	return cacher.OfWithNamespace[analyticsEventRecord](analyticsNamespace)
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
