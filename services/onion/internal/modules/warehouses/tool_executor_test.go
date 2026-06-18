package warehouses

import (
	"encoding/json"
	"testing"
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
