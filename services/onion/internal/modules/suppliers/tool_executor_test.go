package suppliers

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
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
		"supplier_name": "Old Supplier",
		"active":        "inactive",
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

func TestParseListSuppliersToolInputDefaults(t *testing.T) {
	parsed, err := parseListSuppliersToolInput(map[string]any{})
	if err != nil {
		t.Fatalf("parseListSuppliersToolInput returned error: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d, want 1", parsed.Page)
	}
	if parsed.PageSize != defaultListSuppliersToolPageSize {
		t.Fatalf("PageSize = %d, want %d", parsed.PageSize, defaultListSuppliersToolPageSize)
	}
}

func TestParseListSuppliersToolInputCapsPageSize(t *testing.T) {
	parsed, err := parseListSuppliersToolInput(map[string]any{
		"page":  json.Number("2"),
		"limit": json.Number("500"),
	})
	if err != nil {
		t.Fatalf("parseListSuppliersToolInput returned error: %v", err)
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d, want 2", parsed.Page)
	}
	if parsed.PageSize != maxListSuppliersToolPageSize {
		t.Fatalf("PageSize = %d, want %d", parsed.PageSize, maxListSuppliersToolPageSize)
	}
}

func TestParseListSuppliersToolInputRejectsNegativePage(t *testing.T) {
	_, err := parseListSuppliersToolInput(map[string]any{"page": -1})
	if err == nil || !strings.Contains(err.Error(), "page") {
		t.Fatalf("error = %v, want page validation", err)
	}
}

func TestParseGetSupplierToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"supplier_id": int64(42)},
		{"supplier": "#42"},
		{"id": "supplier-42"},
		{"id": "sup_42"},
	}

	for _, input := range cases {
		parsed, err := parseGetSupplierToolInput(input)
		if err != nil {
			t.Fatalf("parseGetSupplierToolInput(%v) returned error: %v", input, err)
		}
		if parsed.SupplierID != 42 {
			t.Fatalf("SupplierID = %d, want 42 for %v", parsed.SupplierID, input)
		}
	}
}

func TestParseGetSupplierToolInputRejectsMissingSupplier(t *testing.T) {
	_, err := parseGetSupplierToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "supplier_id") {
		t.Fatalf("error = %v, want supplier_id validation", err)
	}
}

func TestSupplierToolPayloadFormatsFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 30, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := supplierToolPayload(domain.Supplier{
		ID:           7,
		Name:         "Cedar Wholesale",
		Email:        "orders@cedar.example",
		Phone:        "+96170000000",
		City:         "Beirut",
		Country:      "Lebanon",
		PaymentTerms: "Net 30",
		IsActive:     true,
		CreatedAt:    createdAt,
	})

	if payload["id"] != int64(7) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["name"] != "Cedar Wholesale" {
		t.Fatalf("name = %v", payload["name"])
	}
	if payload["payment_terms"] != "Net 30" {
		t.Fatalf("payment_terms = %v", payload["payment_terms"])
	}
	if payload["created_at"] != "2026-06-18T07:30:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
