package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// The client's roster comes from /hub/team/members and is keyed on users.id,
// while a work day is keyed on hr_employees.id. TimeReports skips any day
// without user_id, so a missing field renders every person as "Not started" and
// 0:00 even when the hours are in the response (#1072). The mapper was already
// correct; what was missing was a value to map.
func TestTeamWorkDayCarriesTheUserWhenResolved(t *testing.T) {
	out := toContractWorkDayV2(domain.WorkDay{
		ID:            408,
		EmployeeID:    9745,
		UserID:        9325,
		LoggedMinutes: 120,
	})

	if out.UserId == nil {
		t.Fatal("user_id is absent, so the client cannot attach this day to anyone")
	}
	if *out.UserId != 9325 {
		t.Errorf("user_id = %d, want 9325", *out.UserId)
	}
	if out.EmployeeId != 9745 {
		t.Errorf("employee_id = %d, want 9745; the two id spaces must both survive", out.EmployeeId)
	}
}

// A caller reading their own day already knows their identity, and an employee
// with no user row has nobody to name. Both leave the field unset rather than
// sending a user id of zero, which the client would treat as a real id.
func TestWorkDayOmitsAnUnresolvedUser(t *testing.T) {
	out := toContractWorkDayV2(domain.WorkDay{ID: 408, EmployeeID: 9745})

	if out.UserId != nil {
		t.Errorf("user_id = %d, want unset", *out.UserId)
	}
}
