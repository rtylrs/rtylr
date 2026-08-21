package infrastructure

import (
	"errors"
	"testing"
)

func TestIsAllowedResourceTable_AllowsApprovedPrefixesAndCoreTables(t *testing.T) {
	for _, table := range []string{"finance_invoices", "hr_employees", "flow_tasks", "users", "businesses"} {
		if !isAllowedResourceTable(table) {
			t.Errorf("expected %q to be allowlisted", table)
		}
	}
}

func TestValidateResourceTableRejectsUnallowlistedTables(t *testing.T) {
	for _, table := range []string{"mysql.user", "_perf_indexes", "sessions", "random_secret_table"} {
		if err := validateResourceTable(table); !errors.Is(err, ErrResourceNotAllowed) {
			t.Errorf("expected %q to be rejected with ErrResourceNotAllowed, got %v", table, err)
		}
	}
}
