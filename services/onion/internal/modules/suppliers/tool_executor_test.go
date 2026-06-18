package suppliers

import (
	"strings"
	"testing"
)

func TestParseCreateSupplierToolInput(t *testing.T) {
	parsed, err := parseCreateSupplierToolInput(map[string]any{
		"name":          "Cedar Wholesale",
		"email":         "orders@cedar.example",
		"phone":         "+96170000000",
		"city":          "Beirut",
		"country":       "Lebanon",
		"payment_terms": "Net 30",
		"notes":         "Preferred protein supplier",
	})
	if err != nil {
		t.Fatalf("parseCreateSupplierToolInput returned error: %v", err)
	}
	if parsed.Name != "Cedar Wholesale" {
		t.Fatalf("Name = %q, want Cedar Wholesale", parsed.Name)
	}
	if parsed.Email != "orders@cedar.example" {
		t.Fatalf("Email = %q, want orders@cedar.example", parsed.Email)
	}
	if !parsed.IsActive {
		t.Fatalf("IsActive = false, want default true")
	}
}

func TestParseCreateSupplierToolInputAllowsInactive(t *testing.T) {
	parsed, err := parseCreateSupplierToolInput(map[string]any{
		"name":      "Old Supplier",
		"is_active": "inactive",
	})
	if err != nil {
		t.Fatalf("parseCreateSupplierToolInput returned error: %v", err)
	}
	if parsed.IsActive {
		t.Fatalf("IsActive = true, want false")
	}
}

func TestParseCreateSupplierToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateSupplierToolInput(map[string]any{"email": "orders@cedar.example"})
	if err == nil || !strings.Contains(err.Error(), "name") {
		t.Fatalf("error = %v, want name validation", err)
	}
}

func TestParseCreateSupplierToolInputRejectsBadEmail(t *testing.T) {
	_, err := parseCreateSupplierToolInput(map[string]any{
		"name":  "Cedar Wholesale",
		"email": "not an email",
	})
	if err == nil || !strings.Contains(err.Error(), "email") {
		t.Fatalf("error = %v, want email validation", err)
	}
}
