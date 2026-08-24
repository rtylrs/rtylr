package api_test

import (
	"testing"

	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
)

// Labour cost is admin-only, and the reason is not just tidiness: the aggregate
// is derived from HR pay rates, so a member being able to read it would make
// compensation inferable through Flow. That is the property these two assert.

func TestProjectLabourCostV2_MemberIsDenied(t *testing.T) {
	f := newRoleFixture(t)

	_, err := f.projAPI.GetProjectLabourCost(ctxAs(roleMemberID),
		&flowprojectv2.GetProjectLabourCostRequest{Id: f.projID})

	assertPermissionDenied(t, err, "member GetProjectLabourCost")
}

func TestProjectLabourCostV2_AdminGetsTheAggregate(t *testing.T) {
	f := newRoleFixture(t)

	res, err := f.projAPI.GetProjectLabourCost(ctxAs(roleManagerID),
		&flowprojectv2.GetProjectLabourCostRequest{Id: f.projID})
	if err != nil {
		t.Fatalf("admin GetProjectLabourCost: %v", err)
	}

	cost := res.GetCost()
	if cost == nil {
		t.Fatal("admin GetProjectLabourCost: cost was nil, so the mapping never ran")
	}
	if cost.GetProjectId() != f.projID {
		t.Fatalf("cost is for the wrong project: got %d want %d", cost.GetProjectId(), f.projID)
	}
	// The project has no logged time, so every total should be zero rather than
	// absent. A zero here is a real answer, not a missing one.
	if cost.GetCostMinorUnits() != 0 || cost.GetCostedMinutes() != 0 || cost.GetUncostedMinutes() != 0 {
		t.Fatalf("expected zero totals on a project with no logged time, got %+v", cost)
	}
}
