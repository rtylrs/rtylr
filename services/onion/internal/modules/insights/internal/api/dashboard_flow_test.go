package api

import (
	"testing"

	flowinsights "github.com/voxire/onion/internal/modules/flow_insights"
)

// insights.dashboard.v1's WorkloadEntry asks for open_tasks and total_active.
// Flow reports neither, so both are derived here. These cases pin the
// arithmetic, because a dashboard reading "3 open" when the answer is 1 is the
// kind of wrong that looks plausible.
func TestToWorkloadEntries_DerivesOpenAndActive(t *testing.T) {
	got := toWorkloadEntries([]flowinsights.FlowWorkloadMember{{
		UserID:          7,
		TotalTasks:      10,
		CompletedTasks:  4,
		InProgressTasks: 3,
		OverdueTasks:    2,
	}})

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	e := got[0]
	if e.UserId != 7 {
		t.Errorf("UserId = %d, want 7", e.UserId)
	}
	// 10 total - 4 done - 3 in progress = 3 assigned but not started.
	if e.OpenTasks != 3 {
		t.Errorf("OpenTasks = %d, want 3", e.OpenTasks)
	}
	// 10 total - 4 done = 6 still outstanding, in progress included.
	if e.TotalActive != 6 {
		t.Errorf("TotalActive = %d, want 6", e.TotalActive)
	}
	if e.InProgressTasks != 3 || e.CompletedTasks != 4 || e.OverdueTasks != 2 {
		t.Errorf("passthrough fields wrong: %+v", e)
	}
}

// Flow's counts should never make these negative, but this feeds a dashboard,
// and a negative task count would be read as a display bug rather than a data
// one. Clamped, and the clamp is asserted rather than assumed.
func TestToWorkloadEntries_ClampsNegativeDerivations(t *testing.T) {
	got := toWorkloadEntries([]flowinsights.FlowWorkloadMember{{
		UserID:          8,
		TotalTasks:      1,
		CompletedTasks:  5,
		InProgressTasks: 4,
	}})

	if got[0].OpenTasks != 0 {
		t.Errorf("OpenTasks = %d, want 0", got[0].OpenTasks)
	}
	if got[0].TotalActive != 0 {
		t.Errorf("TotalActive = %d, want 0", got[0].TotalActive)
	}
}

// An empty result must be an empty slice, not nil: the generated JSON encoder
// renders a nil slice as null, and a dashboard iterating null is a client crash
// rather than an empty state.
func TestToWorkloadEntries_EmptyInputIsEmptySlice(t *testing.T) {
	got := toWorkloadEntries(nil)
	if got == nil {
		t.Fatal("got nil, want an empty slice")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestToRoadmapProjects_EmptyInputIsEmptySlice(t *testing.T) {
	got := toRoadmapProjects(nil)
	if got == nil {
		t.Fatal("got nil, want an empty slice")
	}
}

// velocity_change is the only computed field on this response: Flow reports both
// weeks, the contract wants the delta.
func TestToInsightsResponse_VelocityChangeIsTheWeekDelta(t *testing.T) {
	got := toInsightsResponse(flowinsights.FlowDelivery{
		TotalTasks:        20,
		CompletedTasks:    12,
		InProgressTasks:   5,
		OverdueTasks:      3,
		AvgCompletionDays: 2.5,
		VelocityThisWeek:  9,
		VelocityLastWeek:  4,
	})

	if got.VelocityChange != 5 {
		t.Errorf("VelocityChange = %d, want 5", got.VelocityChange)
	}
	if got.Velocity != 9 {
		t.Errorf("Velocity = %d, want 9", got.Velocity)
	}
	if got.AvgCycleTimeDays != 2.5 {
		t.Errorf("AvgCycleTimeDays = %v, want 2.5", got.AvgCycleTimeDays)
	}
	if got.TotalTasks != 20 || got.CompletedTasks != 12 || got.InProgressTasks != 5 || got.OverdueTasks != 3 {
		t.Errorf("passthrough fields wrong: %+v", got)
	}
}

// A slowing week must report a negative delta, not a clamped zero. The clamp
// belongs on task counts, not on a rate of change.
func TestToInsightsResponse_VelocityChangeCanBeNegative(t *testing.T) {
	got := toInsightsResponse(flowinsights.FlowDelivery{VelocityThisWeek: 2, VelocityLastWeek: 7})
	if got.VelocityChange != -5 {
		t.Errorf("VelocityChange = %d, want -5", got.VelocityChange)
	}
}
