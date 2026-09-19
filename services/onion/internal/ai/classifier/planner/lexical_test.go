package planner

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/ai/classifier/model"
)

func TestElementLabel(t *testing.T) {
	cases := []struct {
		context string
		want    string
	}{
		{`TASK fill the form from the document, then submit` + "\n" + `FORM Demo` + "\n" + `ELEMENT Edit "Phone number" value=""`, "Phone number"},
		{`ELEMENT Checkbox "Has insurance" value="unchecked"`, "Has insurance"},
		{`ELEMENT Button Unquoted value=""`, "Unquoted"},
		{`no element here`, ""},
	}
	for _, c := range cases {
		if got := elementLabel(c.context); got != c.want {
			t.Errorf("elementLabel(%q) = %q, want %q", c.context, got, c.want)
		}
	}
}

func scoreOne(t *testing.T, label string, options ...string) ([]float64, int) {
	t.Helper()
	scorer := LexicalScorer{}
	probs, best, err := scorer.Score(context.Background(), model.Example{
		Context: `ELEMENT Edit "` + label + `" value=""`,
		Options: options,
	})
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	return probs, best
}

func TestLexicalExactLabelFill(t *testing.T) {
	probs, best := scoreOne(t, "Phone number",
		"fill Phone number: (503) 555-0142", "fill Email: jane@example.com", "check", "click", "skip")
	if best != 0 {
		t.Fatalf("best = %d, want 0 (%v)", best, probs)
	}
	if probs[0] < 0.7 {
		t.Fatalf("probability = %v, want a confident fill", probs[0])
	}
}

func TestLexicalSubstringLabelFill(t *testing.T) {
	probs, best := scoreOne(t, "Emergency contact phone",
		"fill Phone: +1 633 208 1384", "fill Legal name: Sofia Larsson", "check", "click", "skip")
	if best != 0 {
		t.Fatalf("best = %d, want 0 (%v)", best, probs)
	}
}

func TestLexicalBooleanFieldChecks(t *testing.T) {
	probs, best := scoreOne(t, "Has insurance",
		"fill Insurance: yes", "fill Name: Jane Doe", "check", "click", "skip")
	if best != 2 {
		t.Fatalf("best = %d, want check (%v)", best, probs)
	}
}

func TestLexicalSubmitButtonClicks(t *testing.T) {
	probs, best := scoreOne(t, "Submit Form",
		"fill Name: Jane Doe", "fill Phone: (503) 555-0142", "check", "click", "skip")
	if best != 3 {
		t.Fatalf("best = %d, want click (%v)", best, probs)
	}
}

func TestLexicalUnrelatedFallsToSkip(t *testing.T) {
	probs, best := scoreOne(t, "Loyalty notes",
		"fill Vendor: Acme", "fill Invoice: INV-1", "check", "click", "skip")
	if best != 4 {
		t.Fatalf("best = %d (%v), want skip", best, probs)
	}
}

func TestLexicalScoreComposesWithBuild(t *testing.T) {
	plan, err := Build(context.Background(), LexicalScorer{}, Request{
		Form: "Patient intake",
		Fields: []Field{
			{ID: "name", Label: "Legal name"},
			{ID: "phone", Label: "Phone number"},
			{ID: "insurance", Label: "Has insurance", Role: "Checkbox"},
		},
		Entities: ExtractEntities("Legal name: Sofia Larsson\nPhone number: (503) 555-0142\nInsurance: Yes\n", 0),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	byID := map[string]Decision{}
	for _, d := range plan.Decisions {
		byID[d.FieldID] = d
	}
	if byID["name"].Action != ActionFill || byID["name"].Value != "Sofia Larsson" {
		t.Fatalf("name = %+v", byID["name"])
	}
	if byID["phone"].Action != ActionFill || byID["phone"].Value != "(503) 555-0142" {
		t.Fatalf("phone = %+v", byID["phone"])
	}
	if byID["insurance"].Action != ActionCheck {
		t.Fatalf("insurance = %+v", byID["insurance"])
	}
	if strings.TrimSpace(plan.Decisions[0].Action) == "" {
		t.Fatal("empty action")
	}
}
