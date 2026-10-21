package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func readQuery(tb testing.TB, path string) string {
	tb.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func namedBlock(tb testing.TB, source, name string) string {
	tb.Helper()
	marker := "-- name: " + name + " "
	start := strings.Index(source, marker)
	if start < 0 {
		tb.Fatalf("query %s not found; rename the test or restore the query", name)
	}
	rest := source[start+len(marker):]
	if next := strings.Index(rest, "-- name: "); next >= 0 {
		return rest[:next]
	}
	return rest
}

func TestPerTaskRollupsRespectModuleOwnership(t *testing.T) {
	t.Run("actual hours is a local projection", func(t *testing.T) {
		body := namedBlock(t, readQuery(t, "../../sql/queries/helpers.sql"), "RecalcActualHours")
		if strings.Contains(body, "hr_time_entries") || strings.Contains(body, "flow_time_logs") {
			t.Fatalf("Flow projection query reads another owned table directly:\n%s", body)
		}
		if !strings.Contains(body, "SET actual_hours = ?") {
			t.Fatalf("Flow projection query does not persist its computed value:\n%s", body)
		}
	})

	t.Run("Flow time logs exclude linked HR rows", func(t *testing.T) {
		body := namedBlock(t, readQuery(t, "../../sql/queries/helpers.sql"), "SumTaskTimeLogMinutes")
		if !strings.Contains(body, "flow_time_logs") || !strings.Contains(body, "time_entry_id IS NULL") {
			t.Fatalf("Flow's local time-log aggregate is incomplete:\n%s", body)
		}
	})

	t.Run("Flow time by member stays in Flow tables", func(t *testing.T) {
		body := namedBlock(t, readQuery(t, "../../sql/queries/features.sql"), "GetInsightsTimeByMember")
		for _, table := range []string{"flow_time_logs", "flow_tasks"} {
			if !strings.Contains(body, table) {
				t.Errorf("GetInsightsTimeByMember does not include Flow-owned %s:\n%s", table, body)
			}
		}
		for _, foreignTable := range []string{"hr_time_entries", "users"} {
			if strings.Contains(body, foreignTable) {
				t.Errorf("GetInsightsTimeByMember directly reads foreign-owned %s:\n%s", foreignTable, body)
			}
		}
		if !strings.Contains(body, "time_entry_id IS NULL") {
			t.Errorf("GetInsightsTimeByMember does not exclude Flow logs linked to a time entry:\n%s", body)
		}
	})

	t.Run("HR time aggregates its own Flow entries", func(t *testing.T) {
		body := namedBlock(t, readQuery(t, "../../../../hr/hr_time/sql/queries/flow_workday_entries.sql"), "SumFlowWorkdayEntryMinutesByMemberAndTaskIDs")
		if !strings.Contains(body, "FROM hr_time_entries") || !strings.Contains(body, "source = 'flow'") || !strings.Contains(body, "business_id = ?") {
			t.Fatalf("HR aggregate is not restricted to HR-owned Flow entries and the requested business:\n%s", body)
		}
		if strings.Contains(body, "JOIN flow_") || strings.Contains(body, "JOIN users") {
			t.Fatalf("HR aggregate directly reads a foreign-owned table:\n%s", body)
		}
	})
}
