package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// ListCycles built each response cycle from four fields, so project_id, dates
// and status were never populated. Every cycle serialised with project_id "0"
// and no dates, and the Cycles page could not match a cycle to the selected
// project or bucket it into Active, Upcoming or Completed. It rendered nothing
// at all for a project that did have a cycle (#1064).
func TestCycleMapperCarriesEveryDeclaredField(t *testing.T) {
	projectID := int64(998833)
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)

	out := toContractCycleV2(domain.Cycle{
		ID:        407,
		ProjectID: &projectID,
		Name:      "QA Sprint 1",
		Status:    "upcoming",
		StartDate: start,
		EndDate:   end,
		Scope:     3,
		Completed: 1,
	})

	cases := []struct {
		field string
		got   any
		want  any
	}{
		{"id", out.Id, int64(407)},
		{"project_id", out.ProjectId, projectID},
		{"name", out.Name, "QA Sprint 1"},
		{"status", out.Status, "upcoming"},
		{"start_date", out.StartDate, "2026-09-07"},
		{"end_date", out.EndDate, "2026-09-18"},
		{"scope", out.Scope, int32(3)},
		{"completed", out.Completed, int32(1)},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %v, want %v", tc.field, tc.got, tc.want)
		}
	}
}

// ProjectID is a pointer on the domain because workspace-era cycles had none.
// The mapper must not dereference it, and a zero date must not serialise as
// 0001-01-01, which the client would have to special-case to read as absent.
func TestCycleMapperHandlesAnAbsentProjectAndDates(t *testing.T) {
	out := toContractCycleV2(domain.Cycle{ID: 1, Name: "legacy"})

	if out.ProjectId != 0 {
		t.Errorf("project_id = %d, want 0 for a cycle with no project", out.ProjectId)
	}
	if out.StartDate != "" || out.EndDate != "" {
		t.Errorf("dates = %q/%q, want empty strings for a zero time", out.StartDate, out.EndDate)
	}
}
