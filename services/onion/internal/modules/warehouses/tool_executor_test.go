package warehouses

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
)

func TestParseCreateWarehouseToolInput(t *testing.T) {
	parsed, err := parseCreateWarehouseToolInput(map[string]any{
		"warehouse_name": " Main Warehouse ",
		"address":        " 123 Port Road ",
		"city":           " Beirut ",
		"country":        " Lebanon ",
		"manager":        " Maya ",
		"capacity":       json.Number("250"),
		"default":        "yes",
		"active":         "true",
	})
	if err != nil {
		t.Fatalf("parseCreateWarehouseToolInput: %v", err)
	}
	if parsed.Name != "Main Warehouse" {
		t.Fatalf("Name = %q", parsed.Name)
	}
	if parsed.Address != "123 Port Road" {
		t.Fatalf("Address = %q", parsed.Address)
	}
	if parsed.City != "Beirut" {
		t.Fatalf("City = %q", parsed.City)
	}
	if parsed.Country != "Lebanon" {
		t.Fatalf("Country = %q", parsed.Country)
	}
	if parsed.ManagerName != "Maya" {
		t.Fatalf("ManagerName = %q", parsed.ManagerName)
	}
	if parsed.Capacity != 250 {
		t.Fatalf("Capacity = %d", parsed.Capacity)
	}
	if !parsed.IsDefault {
		t.Fatal("expected IsDefault to be true")
	}
	if !parsed.IsActive {
		t.Fatal("expected IsActive to be true")
	}
}

func TestParseCreateWarehouseToolInputDefaultsActive(t *testing.T) {
	parsed, err := parseCreateWarehouseToolInput(map[string]any{
		"name": "Branch 2",
	})
	if err != nil {
		t.Fatalf("parseCreateWarehouseToolInput: %v", err)
	}
	if parsed.IsDefault {
		t.Fatal("expected IsDefault to default false")
	}
	if !parsed.IsActive {
		t.Fatal("expected IsActive to default true")
	}
}

func TestParseCreateWarehouseToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateWarehouseToolInput(map[string]any{"capacity": 10})
	if err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestParseCreateWarehouseToolInputRejectsNegativeCapacity(t *testing.T) {
	_, err := parseCreateWarehouseToolInput(map[string]any{
		"name":     "Overflow",
		"capacity": -1,
	})
	if err == nil {
		t.Fatal("expected negative capacity error")
	}
}

func TestParseCreateWarehouseToolInputRejectsFractionalCapacity(t *testing.T) {
	_, err := parseCreateWarehouseToolInput(map[string]any{
		"name":     "Overflow",
		"capacity": 1.5,
	})
	if err == nil {
		t.Fatal("expected fractional capacity error")
	}
}

func TestParseCreateWarehouseToolInputRejectsInvalidBool(t *testing.T) {
	_, err := parseCreateWarehouseToolInput(map[string]any{
		"name":      "Overflow",
		"is_active": "maybe",
	})
	if err == nil {
		t.Fatal("expected invalid bool error")
	}
}

func TestParseGetWarehouseToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"warehouse_id": int64(3)},
		{"warehouse": "#3"},
		{"id": "warehouse-3"},
		{"id": "wh_3"},
	}

	for _, input := range cases {
		parsed, err := parseGetWarehouseToolInput(input)
		if err != nil {
			t.Fatalf("parseGetWarehouseToolInput(%v) returned error: %v", input, err)
		}
		if parsed.WarehouseID != 3 {
			t.Fatalf("WarehouseID = %d, want 3 for %v", parsed.WarehouseID, input)
		}
	}
}

func TestParseGetWarehouseToolInputRejectsMissingWarehouse(t *testing.T) {
	_, err := parseGetWarehouseToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "warehouse_id") {
		t.Fatalf("error = %v, want warehouse_id validation", err)
	}
}

func TestWarehouseToolPayloadFormatsFields(t *testing.T) {
	updatedAt := time.Date(2026, 6, 18, 11, 45, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := warehouseToolPayload(domain.Warehouse{
		ID:          3,
		Name:        "Main Warehouse",
		City:        "Beirut",
		Country:     "Lebanon",
		ManagerName: "Maya",
		Capacity:    250,
		IsDefault:   true,
		IsActive:    true,
		UpdatedAt:   updatedAt,
	})

	if payload["id"] != int64(3) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["name"] != "Main Warehouse" {
		t.Fatalf("name = %v", payload["name"])
	}
	if payload["is_default"] != true {
		t.Fatalf("is_default = %v", payload["is_default"])
	}
	if payload["updated_at"] != "2026-06-18T08:45:00Z" {
		t.Fatalf("updated_at = %v", payload["updated_at"])
	}
}
