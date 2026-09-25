package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

// Flow keeps hours in two tables. Work-day lines are hr_time_entries rows
// carrying a task_id; the task detail time panel writes flow_time_logs. Both
// per-task rollups summed the second alone, so every work-day line counted as
// zero against its own task (tryrtylr/rtylr#903).
//
// These assert on the query text rather than on a database, matching
// child_query_contract_test.go in the invoices module. The behaviour worth
// guarding is which tables a rollup reads, and that is visible in the SQL: a
// future edit that drops one store back out fails here rather than silently
// returning to zeros.
//
// What this deliberately cannot check is arithmetic. That needs a database and
// fixtures, and is worth adding when one is available in this suite.

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

func TestPerTaskRollupsReadBothHourStores(t *testing.T) {
	cases := []struct {
		name  string
		file  string
		query string
	}{
		{"actual_hours", "../../sql/queries/helpers.sql", "RecalcActualHours"},
		{"insights time by member", "../../sql/queries/features.sql", "GetInsightsTimeByMember"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := namedBlock(t, readQuery(t, tc.file), tc.query)

			for _, table := range []string{"hr_time_entries", "flow_time_logs"} {
				if !strings.Contains(body, table) {
					t.Errorf("%s does not read %s, so hours recorded there count as zero:\n%s",
						tc.query, table, body)
				}
			}

			// Without this the same hour, entered in both stores, is counted twice.
			if !strings.Contains(body, "time_entry_id IS NULL") {
				t.Errorf("%s does not exclude flow_time_logs rows linked to a time entry, so an overlapping hour is double counted:\n%s",
					tc.query, body)
			}

			// hr_time_entries.employee_id is a users.id only for rows Flow wrote.
			// hr_time writes entries whose employee_id is an hr_employees.id, and
			// hr_employees.user_id is nullable, so an unscoped read misattributes.
			if !strings.Contains(body, "source = 'flow'") {
				t.Errorf("%s reads hr_time_entries without restricting to source = 'flow', so HR-authored rows are attributed by the wrong id space:\n%s",
					tc.query, body)
			}
		})
	}
}
