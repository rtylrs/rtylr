package api

import "testing"

// The allow case cannot be asserted through scope(): it calls
// tenant.ResolveBusinessID, which fails closed on a nil pool, so a permitted
// caller and a rejected one are both PermissionDenied in a unit test. The
// decision itself is callerMayActFor, which is what these cases cover. The
// deny-path tests in worktime_scope_test.go still go through the API, because
// there the observable result is the one being asserted.
func TestCallerMayActFor(t *testing.T) {
	const caller int64 = 1001

	cases := []struct {
		name       string
		employeeID int64
		want       bool
	}{
		{"unset id means the caller themselves", 0, true},
		{"negative id is treated as unset, not as another employee", -1, true},
		{"the caller's own id, stated explicitly", caller, true},
		{"another employee's id", 9999, false},
		{"another employee's id one below the caller's", caller - 1, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := callerMayActFor(caller, tc.employeeID); got != tc.want {
				t.Fatalf("callerMayActFor(%d, %d) = %v, want %v", caller, tc.employeeID, got, tc.want)
			}
		})
	}
}
