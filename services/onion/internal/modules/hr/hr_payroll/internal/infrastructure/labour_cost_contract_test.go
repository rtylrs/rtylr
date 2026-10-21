package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestProjectLabourCostQueryKeepsCompensationBehindHRBoundary(t *testing.T) {
	raw, err := os.ReadFile("../../sql/queries/labour_cost.sql")
	if err != nil {
		t.Fatalf("read labour cost query: %v", err)
	}
	query := strings.ToLower(string(raw))

	for _, required := range []string{
		"emp.business_id = ?",
		"e.business_id = ?",
		"e.project_id = ?",
		"e.is_billable = true",
		"emp.hourly_rate is null",
		"emp.hourly_rate is not null",
	} {
		if !strings.Contains(query, required) {
			t.Errorf("labour cost query missing %q", required)
		}
	}
	if strings.Contains(query, "owner_id") {
		t.Fatal("labour cost query must not use owner_id for tenant scope")
	}
	if strings.Contains(query, "hourly_rate") && strings.Contains(query, "select emp.hourly_rate") {
		t.Fatal("labour cost query must not return hourly_rate")
	}
}
