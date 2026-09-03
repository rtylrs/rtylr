package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// A work day belongs to one employee, and CreateWorkDayLine used to resolve the
// day with a business-wide scope, so any member of the business could file time
// onto a colleague's day. callerOwnsWorkDay is where that decision now lives;
// the table below is the decision, without a database in the way.
func TestCallerOwnsWorkDay(t *testing.T) {
	const caller = int64(9325)

	cases := []struct {
		name             string
		callerUserID     int64
		dayEmployeeID    int64
		callerEmployeeID int64
		want             bool
	}{
		{"own day, resolved through hr_employees", caller, 9745, 9745, true},
		{"own day, keyed by user id where HR never onboarded", caller, caller, 0, true},
		{"another employee's day", caller, 9741, 9745, false},
		{"another employee's day with no employee row of their own", caller, 9741, 0, false},
		{"unauthenticated caller", 0, 9745, 9745, false},
		{"day with no employee", caller, 0, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := callerOwnsWorkDay(tc.callerUserID, tc.dayEmployeeID, tc.callerEmployeeID)
			if got != tc.want {
				t.Fatalf("callerOwnsWorkDay(%d, %d, %d) = %v, want %v",
					tc.callerUserID, tc.dayEmployeeID, tc.callerEmployeeID, got, tc.want)
			}
		})
	}
}

// The caller identity arrives as WorkDayLineInput.LoggedBy. A zero there means
// no authenticated user reached the service, which must not resolve to somebody
// else's day.
func TestAssertCallerOwnsWorkDay_RejectsUnauthenticatedBeforeTouchingTheRepository(t *testing.T) {
	svc := &Service{}
	err := svc.assertCallerOwnsWorkDay(t.Context(), workDayLineInputWithLoggedBy(0), 9741)
	if !errors.Is(err, ErrWorkDayNotFound) {
		t.Fatalf("want ErrWorkDayNotFound, got %v", err)
	}
}

// The caller's own day short-circuits before the repository is consulted, which
// a nil repo proves: reaching storage here would panic.
func TestAssertCallerOwnsWorkDay_AllowsOwnDayWithoutTouchingTheRepository(t *testing.T) {
	svc := &Service{}
	if err := svc.assertCallerOwnsWorkDay(t.Context(), workDayLineInputWithLoggedBy(9325), 9325); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func workDayLineInputWithLoggedBy(userID int64) domain.WorkDayLineInput {
	return domain.WorkDayLineInput{
		WorkDayID:  409,
		OwnerID:    9003,
		BusinessID: 9003,
		ProjectID:  998833,
		LoggedBy:   userID,
	}
}
