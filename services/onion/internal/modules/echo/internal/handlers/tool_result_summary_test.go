package handlers

import (
	"strings"
	"testing"
)

func TestSummarizeToolResultUsesCollectionCountAndFirstItem(t *testing.T) {
	got := summarizeToolResult("crm.list_customers", nil, map[string]any{
		"customers": []any{
			map[string]any{"first_name": "Maya", "last_name": "Khoury"},
			map[string]any{"first_name": "Rami", "last_name": "Haddad"},
		},
	}, "success", "")

	if got != "Found 2 customers. First: Maya Khoury." {
		t.Fatalf("unexpected summary: %q", got)
	}
}

func TestSummarizeToolResultKeepsErrorsBusinessReadable(t *testing.T) {
	got := summarizeToolResult("store.get_product", nil, nil, "error", "product not found")

	if !strings.Contains(got, "Could not check product") || !strings.Contains(got, "product not found") {
		t.Fatalf("unexpected error summary: %q", got)
	}
}
