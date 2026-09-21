package invalidate

import (
	"testing"

	"github.com/voxire/onion/internal/platform/cache/keys"
)

func TestPatternsForRejectsInvalidNumericID(t *testing.T) {
	got := patternsFor(keys.InvalidateUser, "not-a-number")
	if got != nil {
		t.Fatalf("expected nil patterns for invalid id, got %v", got)
	}
}

func TestPatternsForRejectsZeroNumericID(t *testing.T) {
	got := patternsFor(keys.InvalidateBusiness, "0")
	if got != nil {
		t.Fatalf("expected nil patterns for zero id, got %v", got)
	}
}

func TestPatternsForUserScope(t *testing.T) {
	got := patternsFor(keys.InvalidateUser, "42")
	want := []string{"cache:user:42:*"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("patternsFor(user,42) = %v, want %v", got, want)
	}
}

func TestPatternsForTeamUsesStringBusinessID(t *testing.T) {
	got := patternsFor(keys.InvalidateTeam, "7")
	want := []string{"cache:business:7:team*"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("patternsFor(team,7) = %v, want %v", got, want)
	}
}

func TestPatternsForTeamRejectsEmptyID(t *testing.T) {
	if got := patternsFor(keys.InvalidateTeam, ""); got != nil {
		t.Fatalf("expected nil patterns for empty team id, got %v", got)
	}
}

func TestPatternsForSubscriptionDelimitsOwnerID(t *testing.T) {
	got := patternsFor(keys.InvalidateSubscription, "4")
	want := []string{"cache:subscription:4", "cache:subscription:4:*"}
	if len(got) != len(want) {
		t.Fatalf("patternsFor(subscription,4) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("patternsFor(subscription,4) = %v, want %v", got, want)
		}
	}
}
