package infrastructure_test

import (
	"os"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

const reportQueries = "../../sql/queries/reports.sql"

func TestReportsQueriesBoundInteractiveResults(t *testing.T) {
	testutil.RequireBoundedListQueries(t, reportQueries)
}

// TestReportsQueriesUseTheInteractiveCap pins the cap itself. Boundedness alone
// would accept any literal, and these are interactive reports: the 5001 row
// bound is one over the page size the UI draws, so the caller can tell "there
// is more" without a second count query.
//
// Asserted per query rather than by counting matches, so adding a report with
// the wrong cap fails, and adding one with the right cap does not.
func TestReportsQueriesUseTheInteractiveCap(t *testing.T) {
	const interactiveLimit = "LIMIT 5001"

	for _, q := range testutil.ParseSQLQueries(t, reportQueries) {
		if !q.IsMany() || q.IsPaginated() || q.Exempt() || q.IsAggregate() {
			continue
		}
		if q.HasLimit() && !strings.Contains(q.Text, interactiveLimit) {
			t.Errorf("%s bounds its result but not with %q, so the UI cannot detect overflow", q.Name, interactiveLimit)
		}
	}
}

func TestLeaveSummaryPreAggregatesRequestsBeforeEmployeeJoin(t *testing.T) {
	query, err := os.ReadFile(reportQueries)
	if err != nil {
		t.Fatalf("read reports queries: %v", err)
	}
	sql := string(query)
	if !strings.Contains(sql, "GROUP BY req.employee_id, req.business_id, req.leave_type_id") && !strings.Contains(sql, "GROUP BY employee_id, business_id, leave_type_id") {
		t.Fatal("leave requests must be pre-aggregated by employee, tenant, and leave type")
	}
	if !strings.Contains(sql, ") lr ON lr.employee_id = e.id") {
		t.Fatal("leave summary must join the pre-aggregated request relation")
	}
}
