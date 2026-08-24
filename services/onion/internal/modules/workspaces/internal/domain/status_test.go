package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// The claim the whole bucket design rests on: the four buckets are a refinement
// of the three sets apps/flow already derives in constantsFlow.ts, so nothing
// the frontend computes today changes when buckets arrive.
//
// These are the literal contents of those constants, copied deliberately rather
// than derived, so this test fails if either side drifts.
var (
	frontendOpen   = []string{"backlog", "todo", "in_progress", "in_review"}
	frontendClosed = []string{"done", "cancelled"}
	frontendActive = []string{"in_progress", "in_review"}
)

func bucketsFor(statuses []string) map[string]bool {
	out := map[string]bool{}
	for _, s := range statuses {
		out[domain.BucketForTaskStatus(s)] = true
	}
	return out
}

func TestBucketsReconstructTheFrontendSets(t *testing.T) {
	// OPEN is exactly backlog plus active.
	open := bucketsFor(frontendOpen)
	if len(open) != 2 || !open[domain.StatusBucketBacklog] || !open[domain.StatusBucketActive] {
		t.Fatalf("OPEN should map onto {backlog, active}, got %v", open)
	}
	// CLOSED is exactly done plus cancelled.
	closed := bucketsFor(frontendClosed)
	if len(closed) != 2 || !closed[domain.StatusBucketDone] || !closed[domain.StatusBucketCancelled] {
		t.Fatalf("CLOSED should map onto {done, cancelled}, got %v", closed)
	}
	// ACTIVE is exactly the active bucket and nothing else. This is the one that
	// pins `todo` to backlog: if todo were active, ACTIVE would silently widen
	// to include it and every "in flight" count in the app would change.
	active := bucketsFor(frontendActive)
	if len(active) != 1 || !active[domain.StatusBucketActive] {
		t.Fatalf("ACTIVE should map onto {active} alone, got %v", active)
	}
	if got := domain.BucketForTaskStatus("todo"); got != domain.StatusBucketBacklog {
		t.Fatalf("todo must bucket as backlog or ACTIVE widens, got %q", got)
	}
	// OPEN and CLOSED must not overlap, or a task counts twice in a rollup.
	for b := range open {
		if closed[b] {
			t.Fatalf("bucket %q is in both OPEN and CLOSED", b)
		}
	}
}

func TestBucketForUnknownStatusIsEmptyNotADefault(t *testing.T) {
	// Returning a default here is how a cancelled task gets counted as open.
	for _, in := range []string{"", "in_qa", "DONE", "Done", "archived"} {
		if got := domain.BucketForTaskStatus(in); got != "" {
			t.Errorf("BucketForTaskStatus(%q) = %q, want empty", in, got)
		}
	}
}

func TestEverySixBuiltInsHaveABucket(t *testing.T) {
	for _, s := range []string{
		domain.TaskStatusBacklog, domain.TaskStatusTodo, domain.TaskStatusInProgress,
		domain.TaskStatusInReview, domain.TaskStatusDone, domain.TaskStatusCancelled,
	} {
		if got := domain.BucketForTaskStatus(s); got == "" {
			t.Errorf("built-in %q has no bucket", s)
		}
	}
}

func TestLegacyStatusForBucketRoundTrips(t *testing.T) {
	// Whatever legacy status a bucket derives must map back to that same bucket,
	// or the string column would contradict the bucket it came from.
	for _, b := range []string{
		domain.StatusBucketBacklog, domain.StatusBucketActive,
		domain.StatusBucketDone, domain.StatusBucketCancelled,
	} {
		legacy := domain.LegacyStatusForBucket(b)
		if legacy == "" {
			t.Fatalf("bucket %q derives no legacy status", b)
		}
		if !domain.IsValidTaskStatus(legacy) {
			t.Fatalf("bucket %q derives %q, which is not a valid task status", b, legacy)
		}
		if back := domain.BucketForTaskStatus(legacy); back != b {
			t.Errorf("bucket %q -> legacy %q -> bucket %q, want round trip", b, legacy, back)
		}
	}
	if got := domain.LegacyStatusForBucket("nonsense"); got != "" {
		t.Errorf("unknown bucket derived %q, want empty", got)
	}
}

func TestNormalizeStatusInput(t *testing.T) {
	tests := []struct {
		name       string
		in         domain.StatusInput
		wantErr    error
		wantLegacy string
		wantColor  string
		wantName   string
	}{
		{
			name:       "derives legacy status from the bucket when absent",
			in:         domain.StatusInput{Name: "Building", Bucket: "active"},
			wantLegacy: domain.TaskStatusInProgress,
			wantColor:  domain.DefaultStatusColor,
			wantName:   "Building",
		},
		{
			name:       "keeps an explicit legacy status",
			in:         domain.StatusInput{Name: "In legal", Bucket: "active", LegacyStatus: "in_review"},
			wantLegacy: "in_review",
			wantColor:  domain.DefaultStatusColor,
			wantName:   "In legal",
		},
		{
			name:      "trims the name",
			in:        domain.StatusInput{Name: "  Shipped  ", Bucket: "done"},
			wantName:  "Shipped",
			wantColor: domain.DefaultStatusColor,
			// done maps to itself
			wantLegacy: domain.TaskStatusDone,
		},
		{
			name:    "rejects a blank name",
			in:      domain.StatusInput{Name: "   ", Bucket: "done"},
			wantErr: domain.ErrStatusNameRequired,
		},
		{
			name:    "rejects an unknown bucket rather than defaulting it",
			in:      domain.StatusInput{Name: "Waiting", Bucket: "waiting"},
			wantErr: domain.ErrInvalidStatusBucket,
		},
		{
			name:    "rejects an unknown legacy status",
			in:      domain.StatusInput{Name: "Building", Bucket: "active", LegacyStatus: "in_qa"},
			wantErr: domain.ErrInvalidLegacyStatus,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NormalizeStatusInput(tc.in)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tc.wantName {
				t.Errorf("name = %q, want %q", got.Name, tc.wantName)
			}
			if got.LegacyStatus != tc.wantLegacy {
				t.Errorf("legacy = %q, want %q", got.LegacyStatus, tc.wantLegacy)
			}
			if got.Color != tc.wantColor {
				t.Errorf("color = %q, want %q", got.Color, tc.wantColor)
			}
		})
	}
}

func TestValidStatusBucket(t *testing.T) {
	for _, ok := range []string{"backlog", "active", "done", "cancelled"} {
		if !domain.ValidStatusBucket(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "Backlog", "todo", "in_progress", "archived", "open"} {
		if domain.ValidStatusBucket(bad) {
			t.Errorf("%q should not be valid", bad)
		}
	}
}
