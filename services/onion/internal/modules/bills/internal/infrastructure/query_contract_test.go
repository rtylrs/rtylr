package infrastructure

import (
	"strings"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

const billQueries = "../../sql/queries/bills.sql"

func TestBillListQueriesHaveServerBounds(t *testing.T) {
	testutil.RequireBoundedListQueries(t, billQueries)
}

// TestBillListQueriesUseStableOrdering pins the total order every bill list
// shares. due_date is not unique, so ListBillsPaginated's LIMIT ? OFFSET ?
// could otherwise repeat or skip rows between pages; the id tiebreak is what
// makes paging sound.
//
// This asserts the property per query rather than counting how many statements
// mention the clause. The count form passed when a new list query arrived
// without the ordering and failed when one arrived with it, which is how #892
// turned main red while leaving the real regression uncaught.
func TestBillListQueriesUseStableOrdering(t *testing.T) {
	const stableOrder = "ORDER BY due_date ASC, id ASC"

	for _, q := range testutil.ParseSQLQueries(t, billQueries) {
		if !q.IsMany() || q.IsAggregate() {
			continue
		}
		if !strings.Contains(q.Text, "FROM finance_bills") {
			continue
		}
		if !strings.Contains(q.Text, stableOrder) {
			t.Errorf("%s selects bill rows without %q, so its paging is not stable", q.Name, stableOrder)
		}
	}
}
