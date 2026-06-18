package brands

import "testing"

func TestParseCreateBrandToolInput(t *testing.T) {
	parsed, err := parseCreateBrandToolInput(map[string]any{
		"brand_name":  " Optimum Nutrition ",
		"logo":        " https://example.com/on.png ",
		"description": " Protein supplements ",
	})
	if err != nil {
		t.Fatalf("parseCreateBrandToolInput: %v", err)
	}
	if parsed.Name != "Optimum Nutrition" {
		t.Fatalf("Name = %q", parsed.Name)
	}
	if parsed.LogoURL != "https://example.com/on.png" {
		t.Fatalf("LogoURL = %q", parsed.LogoURL)
	}
	if parsed.Description != "Protein supplements" {
		t.Fatalf("Description = %q", parsed.Description)
	}
}

func TestParseCreateBrandToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateBrandToolInput(map[string]any{"description": "No name"})
	if err == nil {
		t.Fatal("expected missing name error")
	}
}
