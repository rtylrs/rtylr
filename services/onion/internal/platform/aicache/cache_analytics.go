package aicache

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	aicachesqlc "github.com/voxire/onion/internal/platform/aicache/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

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
