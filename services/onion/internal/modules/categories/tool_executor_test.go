package categories

import "testing"

func TestParseCreateCategoryToolInput(t *testing.T) {
	parsed, err := parseCreateCategoryToolInput(map[string]any{
		"category_name": " Supplements ",
		"description":   " Sports nutrition ",
	})
	if err != nil {
		t.Fatalf("parseCreateCategoryToolInput: %v", err)
	}
	if parsed.Name != "Supplements" {
		t.Fatalf("Name = %q", parsed.Name)
	}
	if parsed.Description != "Sports nutrition" {
		t.Fatalf("Description = %q", parsed.Description)
	}
}

func TestParseCreateCategoryToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateCategoryToolInput(map[string]any{"description": "No name"})
	if err == nil {
		t.Fatal("expected missing name error")
	}
}
