// Package orgchart defines the platform boundary for reporting relationships.
// Consumers receive only the set of user ids a manager is responsible for; the
// staff records behind it, and every compensation or personal field on them,
// stay inside HR so they cannot leak through Flow.
//
// This exists because docs/sqlc-query-ownership.md forbids importing another
// module's internal/infrastructure package to reuse its SQL, and permits
// sharing "only through an explicit application-facing provider interface".
// Flow needs to know who reports to a caller; it must not learn anything else
// about them.
package orgchart

import "context"

// ManagerResolver reports which users a manager is responsible for.
//
// Implementations resolve every rule source they know about and return the
// union: today the department manager on an HR department, later a
// person-to-person reporting line on the staff record. Consumers must not care
// which rule produced a given id, so adding a source is not a breaking change.
//
// The ids are USER ids, not staff ids. A staff record without a linked user
// account cannot be a Flow subject at all, so it is omitted rather than
// returned as a zero.
type ManagerResolver interface {
	// ManagedUserIDs returns the user ids reporting to managerUserID inside
	// businessID, excluding the manager. An empty slice and a nil error mean
	// the caller manages nobody, which is the common case and not an error.
	ManagedUserIDs(ctx context.Context, managerUserID, businessID int64) ([]int64, error)
}
