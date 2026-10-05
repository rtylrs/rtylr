package testutil

import "testing"

// A second tenant, so a cross-tenant assertion costs one line rather than a
// fixture rewrite.
//
// A tenant-scoped read tested against one tenant passes whether or not the
// business_id filter exists: seed tenant A, read as tenant A, get A's rows back.
// The only test that catches a missing or wrong filter seeds a second tenant and
// asserts its rows are absent. These exist so that second tenant does not have to
// be invented per package.
//
// Deliberately far outside the fixture id range (1001-3702 in internal/testutil/fixtures)
// so a row seeded for the other tenant can never collide with a fixture row.
const (
	OtherOwnerID    int64 = 991001
	OtherBusinessID int64 = 992001
)

// AssertTenantIsolation fails unless got contains every id in want and none of
// the ids in forbidden.
//
// Both halves are required, and that is the point of the helper. A test that
// only asserts the other tenant's rows are absent passes when the query returns
// nothing at all: a broken filter, a wrong join, or an empty seed all produce
// the same green result as correct isolation. Requiring the caller's own rows to
// be present too is what makes the absence meaningful.
func AssertTenantIsolation[T comparable](tb testing.TB, got, want, forbidden []T) {
	tb.Helper()

	if len(want) == 0 {
		tb.Fatal("AssertTenantIsolation: want is empty, so absence would prove nothing")
	}

	seen := make(map[T]struct{}, len(got))
	for _, g := range got {
		seen[g] = struct{}{}
	}

	for _, w := range want {
		if _, ok := seen[w]; !ok {
			tb.Errorf("own tenant row %v missing from the result", w)
		}
	}

	for _, f := range forbidden {
		if _, ok := seen[f]; ok {
			tb.Errorf("cross-tenant leak: row %v belongs to another tenant", f)
		}
	}
}
