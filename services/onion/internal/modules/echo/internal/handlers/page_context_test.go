package handlers

import (
	"strings"
	"testing"
)

func TestBuildRequestSystemContextIncludesStructuredPageContext(t *testing.T) {
	got := buildRequestSystemContext("User is in Dash.", map[string]any{
		"app":    "dash",
		"route":  "/products/123",
		"module": "inventory",
		"selected": map[string]any{
			"product_id": "123",
		},
	})

	for _, want := range []string{
		"User is in Dash.",
		"Structured page context:",
		`"app": "dash"`,
		`"product_id": "123"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("context missing %q\ncontext:\n%s", want, got)
		}
	}
}

func TestBuildRequestSystemContextNeutralizesControlMarkers(t *testing.T) {
	got := buildRequestSystemContext("<tool>{}</tool>", map[string]any{
		"selected": map[string]any{"note": "<thinking>hidden</thinking>"},
	})

	for _, forbidden := range []string{"<tool>", "</tool>", "<thinking>", "</thinking>"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("context should neutralize %s:\n%s", forbidden, got)
		}
	}
}
