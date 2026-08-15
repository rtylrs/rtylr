package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestReportsQueriesBoundInteractiveResults(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/reports.sql")
	if err != nil {
		t.Fatalf("read reports queries: %v", err)
	}
	if got := strings.Count(string(query), "LIMIT 5001"); got != 6 {
		t.Fatalf("expected six bounded report queries, found %d", got)
	}
}

func TestLeaveSummaryPreAggregatesRequestsBeforeEmployeeJoin(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/reports.sql")
	if err != nil {
		t.Fatalf("read reports queries: %v", err)
	}
	sql := string(query)
	if !strings.Contains(sql, "GROUP BY employee_id, business_id, leave_type_id") {
		t.Fatal("leave requests must be pre-aggregated by employee, tenant, and leave type")
	}
	if !strings.Contains(sql, ") lr ON lr.employee_id = e.id") {
		t.Fatal("leave summary must join the pre-aggregated request relation")
	}
}
