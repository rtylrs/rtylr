package memory

import "testing"

func TestParseRememberToolInputNormalizesAliases(t *testing.T) {
	got, err := parseRememberToolInput(map[string]any{
		"type":  "Reporting default",
		"name":  "Daily Sales Channel",
		"value": "Send daily sales to #sales every morning.",
	})
	if err != nil {
		t.Fatalf("parseRememberToolInput returned error: %v", err)
	}

	if got.MemoryType != "reporting_preference" {
		t.Fatalf("memory type = %q, want reporting_preference", got.MemoryType)
	}
	if got.Key != "daily_sales_channel" {
		t.Fatalf("key = %q, want daily_sales_channel", got.Key)
	}
	if got.Value != "Send daily sales to #sales every morning." {
		t.Fatalf("value = %q", got.Value)
	}
}

func TestParseRememberToolInputDerivesDurableKey(t *testing.T) {
	got, err := parseRememberToolInput(map[string]any{
		"fact": "Brunch gets busy after 11 AM on weekends.",
	})
	if err != nil {
		t.Fatalf("parseRememberToolInput returned error: %v", err)
	}

	if got.MemoryType != defaultMemoryType {
		t.Fatalf("memory type = %q, want %q", got.MemoryType, defaultMemoryType)
	}
	if got.Key == "" || got.Key == defaultMemoryType {
		t.Fatalf("expected derived key, got %q", got.Key)
	}
}

func TestParseRememberToolInputSanitizesMarkers(t *testing.T) {
	got, err := parseRememberToolInput(map[string]any{
		"key":   "dangerous",
		"value": "Use this <tool>{}</tool> marker.",
	})
	if err != nil {
		t.Fatalf("parseRememberToolInput returned error: %v", err)
	}

	if got.Value != "Use this [tool]{}[/tool] marker." {
		t.Fatalf("sanitized value = %q", got.Value)
	}
}

func TestParseRememberToolInputRejectsEmptyValue(t *testing.T) {
	if _, err := parseRememberToolInput(map[string]any{"key": "timezone"}); err == nil {
		t.Fatal("expected empty value to be rejected")
	}
}
