package aicache

import (
	"strings"
	"testing"
	"time"
)

func TestExactKeyIsStableAndRedactsPrompt(t *testing.T) {
	parts := ExactKeyParts{
		BusinessID: 1,
		UserID:     2,
		Mode:       "support",
		Provider:   "deepseek",
		Prompt:     "secret customer revenue prompt",
	}

	a := ExactKey(parts)
	b := ExactKey(parts)

	if a != b {
		t.Fatalf("ExactKey not stable: %q != %q", a, b)
	}
	if strings.Contains(a, "secret") || strings.Contains(a, "customer") {
		t.Fatalf("ExactKey leaked prompt text: %q", a)
	}
	if len(a) != 64 {
		t.Fatalf("ExactKey length = %d, want 64", len(a))
	}
}

func TestExactKeyIncludesTenantScope(t *testing.T) {
	base := ExactKeyParts{
		BusinessID: 1,
		UserID:     2,
		Mode:       "ask",
		Provider:   "deepseek",
		Prompt:     "same prompt",
	}

	otherTenant := base
	otherTenant.BusinessID = 9

	if ExactKey(base) == ExactKey(otherTenant) {
		t.Fatal("expected tenant change to produce a different exact key")
	}
}

func TestExactKeyIncludesStableRequestContext(t *testing.T) {
	base := ExactKeyParts{
		BusinessID:    1,
		UserID:        2,
		Mode:          "ask",
		Provider:      "deepseek",
		RequestDate:   "2026-06-19",
		SystemContext: "User is in the Dash app.",
		Prompt:        "show my top customers",
	}

	otherContext := base
	otherContext.SystemContext = "User is in the POS app."
	if ExactKey(base) == ExactKey(otherContext) {
		t.Fatal("expected app context change to produce a different exact key")
	}

	otherDate := base
	otherDate.RequestDate = "2026-06-20"
	if ExactKey(base) == ExactKey(otherDate) {
		t.Fatal("expected request date change to produce a different exact key")
	}
}

func TestExactKeyIncludesDataFreshness(t *testing.T) {
	base := ExactKeyParts{
		BusinessID:    1,
		UserID:        2,
		Mode:          "ask",
		Provider:      "deepseek",
		RequestDate:   "2026-06-20",
		SystemContext: "User is in inventory.",
		DataFreshness: `{"module":"erp","version":"products:1"}`,
		Prompt:        "which products are low?",
	}

	other := base
	other.DataFreshness = `{"module":"erp","version":"products:2"}`

	if ExactKey(base) == ExactKey(other) {
		t.Fatal("expected data freshness change to produce a different exact key")
	}
}

func TestExactKeyNormalizesWhitespace(t *testing.T) {
	base := ExactKeyParts{
		BusinessID:    1,
		UserID:        2,
		Mode:          "ask",
		Provider:      "deepseek",
		RequestDate:   "2026-06-19",
		SystemContext: "User is in the\nDash app.",
		Prompt:        " show   my\n top customers ",
	}
	same := base
	same.SystemContext = "User is in the Dash app."
	same.Prompt = "show my top customers"

	if ExactKey(base) != ExactKey(same) {
		t.Fatal("expected whitespace-only changes to preserve the exact key")
	}
}

func TestExactPolicySkipsUnsafeModes(t *testing.T) {
	t.Setenv("AI_RESPONSE_CACHE_ENABLED", "")

	for _, mode := range []string{ModeAgent, ModeAnalyst} {
		policy := ExactPolicyForMode(mode)
		if policy.Enabled {
			t.Fatalf("expected %s mode to skip exact cache", mode)
		}
		if policy.Reason == "" {
			t.Fatalf("expected %s mode to explain skip", mode)
		}
	}
}

func TestExactPolicyAllowsAskAndSupport(t *testing.T) {
	t.Setenv("AI_RESPONSE_CACHE_ENABLED", "")

	for _, mode := range []string{ModeAsk, ModeSupport} {
		policy := ExactPolicyForMode(mode)
		if !policy.Enabled {
			t.Fatalf("expected %s mode to allow exact cache: %s", mode, policy.Reason)
		}
		if policy.TTLSec <= 0 {
			t.Fatalf("expected %s mode to expose positive TTL", mode)
		}
	}
}

func TestSnapshotAggregatesEvents(t *testing.T) {
	ResetForTest()

	Record(Event{Outcome: OutcomeMiss, Mode: ModeAsk})
	Record(Event{Outcome: OutcomeWrite, Mode: ModeAsk, ProviderTokens: 42, ResponseChars: 100, LatencyMs: 80})
	Record(Event{Outcome: OutcomeHit, Mode: ModeAsk, EstimatedSavings: 42, LatencyMs: 5})
	Record(Event{Outcome: OutcomeSkip, Mode: ModeAgent, Reason: "agent mode"})

	snap := Snapshot()
	if snap.Totals.Lookups != 3 {
		t.Fatalf("lookups = %d, want 3", snap.Totals.Lookups)
	}
	if snap.Totals.Hits != 1 || snap.Totals.Misses != 1 || snap.Totals.Skips != 1 {
		t.Fatalf("unexpected totals: %+v", snap.Totals)
	}
	if snap.Totals.Writes != 1 {
		t.Fatalf("writes = %d, want 1", snap.Totals.Writes)
	}
	if snap.Totals.ProviderTokens != 42 || snap.Totals.TokensSaved != 42 {
		t.Fatalf("unexpected token accounting: %+v", snap.Totals)
	}
	if snap.HitRate != 0.5 {
		t.Fatalf("hit rate = %f, want 0.5", snap.HitRate)
	}
	if len(snap.Recent) != 4 {
		t.Fatalf("recent events = %d, want 4", len(snap.Recent))
	}
}

func TestSnapshotFromEventsAggregatesPersistedEvents(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{At: now.Add(-1 * time.Minute), Outcome: OutcomeHit, Mode: ModeAsk, EstimatedSavings: 42, LatencyMs: 5},
		{At: now.Add(-3 * time.Minute), Outcome: OutcomeMiss, Mode: ModeAsk, LatencyMs: 15},
		{At: now.Add(-2 * time.Minute), Outcome: OutcomeWrite, Mode: ModeAsk, ProviderTokens: 42, ResponseChars: 100, LatencyMs: 40},
		{At: now.Add(-4 * time.Minute), Outcome: OutcomeSkip, Mode: ModeAgent, Reason: "agent mode"},
	}

	snap := snapshotFromEvents(events)

	if !snap.WindowStartedAt.Equal(now.Add(-4 * time.Minute)) {
		t.Fatalf("window start = %s, want %s", snap.WindowStartedAt, now.Add(-4*time.Minute))
	}
	if snap.Totals.Lookups != 3 || snap.Totals.Hits != 1 || snap.Totals.Misses != 1 || snap.Totals.Skips != 1 {
		t.Fatalf("unexpected lookup totals: %+v", snap.Totals)
	}
	if snap.Totals.Writes != 1 || snap.Totals.ProviderTokens != 42 || snap.Totals.ResponseBytes != 100 {
		t.Fatalf("unexpected write totals: %+v", snap.Totals)
	}
	if snap.Totals.TokensSaved != 42 || snap.HitRate != 0.5 || snap.AverageLatencyMs != 20 {
		t.Fatalf("unexpected derived metrics: hit_rate=%f avg_latency=%d totals=%+v", snap.HitRate, snap.AverageLatencyMs, snap.Totals)
	}
	if len(snap.Recent) != 4 || snap.Recent[0].Outcome != OutcomeHit || snap.Recent[3].Outcome != OutcomeSkip {
		t.Fatalf("recent events not sorted newest first: %+v", snap.Recent)
	}
}
