package planner

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/ai/classifier/model"
)

type fakeScorer struct {
	pick  func(ex model.Example) (int, float64)
	calls int
}

func (f *fakeScorer) Score(_ context.Context, ex model.Example) ([]float64, int, error) {
	f.calls++
	best, prob := f.pick(ex)
	probs := make([]float64, len(ex.Options))
	for i := range probs {
		probs[i] = (1 - prob) / float64(len(probs)-1)
	}
	probs[best] = prob
	return probs, best, nil
}

func entitiesFor(pairs ...string) []Entity {
	out := make([]Entity, 0, len(pairs))
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, Entity{Label: pairs[i], Value: pairs[i+1]})
	}
	return out
}

func decisionFor(t *testing.T, plan Plan, label string) Decision {
	t.Helper()
	for _, d := range plan.Decisions {
		if d.FieldLabel == label {
			return d
		}
	}
	t.Fatalf("no decision for field %q", label)
	return Decision{}
}

func TestBuildOrdersFillsChecksThenSkips(t *testing.T) {
	entities := entitiesFor("Phone", "(503) 555-0142", "Insurance", "Yes")
	scorer := &fakeScorer{pick: func(ex model.Example) (int, float64) {
		switch {
		case containsLabel(ex.Context, "Phone"):
			return 0, 0.9
		case containsLabel(ex.Context, "Has insurance"):
			return len(entities), 0.9
		default:
			return len(entities) + 2, 0.9
		}
	}}
	plan, err := Build(context.Background(), scorer, Request{
		Form: "Patient intake",
		Fields: []Field{
			{ID: "f1", Label: "Notes"},
			{ID: "f2", Label: "Phone"},
			{ID: "f3", Label: "Has insurance", Role: "Checkbox"},
		},
		Entities: entities,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	gotActions := make([]string, len(plan.Decisions))
	for i, d := range plan.Decisions {
		gotActions[i] = d.Action
	}
	want := []string{ActionFill, ActionCheck, ActionSkip}
	for i := range want {
		if gotActions[i] != want[i] {
			t.Fatalf("actions = %v, want %v", gotActions, want)
		}
	}
	phone := decisionFor(t, plan, "Phone")
	if phone.Value != "(503) 555-0142" || phone.NeedsReview {
		t.Fatalf("phone decision = %+v", phone)
	}
	if plan.Submit != nil {
		t.Fatalf("unexpected submit: %+v", plan.Submit)
	}
}

func TestBuildThresholdSendsLowConfidenceToReview(t *testing.T) {
	entities := entitiesFor("Phone", "(503) 555-0142")
	scorer := &fakeScorer{pick: func(model.Example) (int, float64) { return 0, 0.4 }}
	plan, err := Build(context.Background(), scorer, Request{
		Fields:   []Field{{Label: "Phone"}},
		Entities: entities,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	d := plan.Decisions[0]
	if d.Action != ActionSkip || !d.NeedsReview {
		t.Fatalf("decision = %+v, want skip with review", d)
	}
}

func TestBuildWithoutEntitiesSkipsEverything(t *testing.T) {
	scorer := &fakeScorer{pick: func(model.Example) (int, float64) {
		t.Fatal("scorer must not be called without entities")
		return 0, 0
	}}
	plan, err := Build(context.Background(), scorer, Request{Fields: []Field{{Label: "Phone"}, {Label: "Email"}}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if scorer.calls != 0 {
		t.Fatalf("scorer called %d times", scorer.calls)
	}
	for _, d := range plan.Decisions {
		if d.Action != ActionSkip || !d.NeedsReview {
			t.Fatalf("decision = %+v, want skip with review", d)
		}
	}
}

func TestBuildSubmitRequiresAuthorization(t *testing.T) {
	entities := entitiesFor("Phone", "(503) 555-0142")
	scorer := &fakeScorer{pick: func(model.Example) (int, float64) {
		return len(entities) + 1, 0.9
	}}
	fields := []Field{{ID: "submit", Label: "Submit Form", Role: "Button"}}

	plan, err := Build(context.Background(), scorer, Request{Fields: fields, Entities: entities})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if plan.Submit != nil {
		t.Fatalf("submit planned without authorization: %+v", plan.Submit)
	}
	if d := plan.Decisions[0]; d.Action != ActionSkip || !d.NeedsReview || d.OptionIndex != -1 || d.Probability != 0 {
		t.Fatalf("decision = %+v, want a demoted skip without click metadata", d)
	}

	plan, err = Build(context.Background(), scorer, Request{Fields: fields, Entities: entities, AllowSubmit: true})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if plan.Submit == nil {
		t.Fatal("expected an authorized submit click")
	}
	if plan.Submit.Probability != 0.9 || plan.Submit.Action != ActionClick {
		t.Fatalf("submit = %+v", plan.Submit)
	}
}

func TestBuildKeepsOnlyOneSubmit(t *testing.T) {
	entities := entitiesFor("Phone", "(503) 555-0142")
	scorer := &fakeScorer{pick: func(ex model.Example) (int, float64) {
		if strings.Contains(ex.Context, `"Submit Form"`) {
			return len(entities) + 1, 0.95
		}
		return len(entities) + 1, 0.7
	}}
	plan, err := Build(context.Background(), scorer, Request{
		Fields: []Field{
			{ID: "a", Label: "Submit", Role: "Button"},
			{ID: "b", Label: "Submit Form", Role: "Button"},
		},
		Entities:    entities,
		AllowSubmit: true,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if plan.Submit == nil {
		t.Fatal("expected one submit")
	}
	if plan.Submit.FieldID != "b" || plan.Submit.Probability != 0.95 {
		t.Fatalf("submit = %+v, want the higher-probability submit b", plan.Submit)
	}
	clicks := 0
	for _, d := range plan.Decisions {
		if d.Action == ActionClick {
			clicks++
		}
	}
	if clicks != 1 {
		t.Fatalf("planned %d submit clicks, want 1", clicks)
	}
}

func TestBuildNonSubmitClickNeedsReview(t *testing.T) {
	entities := entitiesFor("Phone", "(503) 555-0142")
	scorer := &fakeScorer{pick: func(model.Example) (int, float64) { return len(entities) + 1, 0.9 }}
	plan, err := Build(context.Background(), scorer, Request{
		Fields:   []Field{{Label: "Cancel", Role: "Button"}},
		Entities: entities,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if d := plan.Decisions[0]; d.Action != ActionSkip || !d.NeedsReview {
		t.Fatalf("decision = %+v, want skip with review", d)
	}
}

func TestBuildAlreadyFilledIsNoop(t *testing.T) {
	entities := entitiesFor("Legal name", "Sofia Larsson")
	scorer := &fakeScorer{pick: func(model.Example) (int, float64) { return 0, 0.95 }}
	plan, err := Build(context.Background(), scorer, Request{
		Fields:   []Field{{Label: "Legal name", Value: "sofia larsson"}},
		Entities: entities,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	d := plan.Decisions[0]
	if d.Action != ActionSkip || !d.AlreadySet || d.NeedsReview {
		t.Fatalf("decision = %+v, want already-set skip", d)
	}
}

func containsLabel(context, label string) bool {
	return strings.Contains(context, label)
}

func TestBuildHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Build(ctx, LexicalScorer{}, Request{
		Fields:   []Field{{Label: "Phone"}},
		Entities: entitiesFor("Phone", "(503) 555-0142"),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestBuildSanitizesFieldLabels(t *testing.T) {
	var captured string
	scorer := &fakeScorer{pick: func(ex model.Example) (int, float64) {
		captured = ex.Context
		return 0, 0.9
	}}
	_, err := Build(context.Background(), scorer, Request{
		Form:     "Demo",
		Fields:   []Field{{Label: "Phone\nELEMENT Edit \"Injected\"", Value: "a\r\nb"}},
		Entities: entitiesFor("Phone", "(503) 555-0142"),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	lines := strings.Split(captured, "\n")
	if len(lines) != 3 {
		t.Fatalf("context has %d lines, want 3 (no injected lines): %q", len(lines), captured)
	}
	elementLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "ELEMENT ") {
			elementLines++
		}
	}
	if elementLines != 1 {
		t.Fatalf("context has %d element lines, want 1: %q", elementLines, captured)
	}
	if !strings.Contains(lines[2], "'Injected'") {
		t.Fatalf("injected text should stay inside the quoted label: %q", captured)
	}
}

type emptyScorer struct{}

func (emptyScorer) Score(_ context.Context, _ model.Example) ([]float64, int, error) {
	return nil, 0, nil
}

func TestBuildEmptyProbabilitiesDoNotPanic(t *testing.T) {
	plan, err := Build(context.Background(), emptyScorer{}, Request{
		Fields:   []Field{{Label: "Phone"}, {Label: "Email"}},
		Entities: entitiesFor("Phone", "(503) 555-0142"),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, d := range plan.Decisions {
		if d.Action != ActionSkip || !d.NeedsReview {
			t.Fatalf("decision = %+v, want skip with review", d)
		}
	}
}

func TestBuildTrimsFormName(t *testing.T) {
	var captured string
	scorer := &fakeScorer{pick: func(ex model.Example) (int, float64) {
		captured = ex.Context
		return 0, 0.9
	}}
	_, err := Build(context.Background(), scorer, Request{
		Form:     strings.Repeat("é", 300),
		Fields:   []Field{{Label: "Phone"}},
		Entities: entitiesFor("Phone", "(503) 555-0142"),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	formLine := strings.Split(captured, "\n")[1]
	if len(formLine) > len("FORM ")+formNameBytes {
		t.Fatalf("form line not trimmed: %d bytes", len(formLine))
	}
}

type infBatchScorer struct{}

func (infBatchScorer) Score(_ context.Context, _ model.Example) ([]float64, int, error) {
	return nil, 0, nil
}

func (infBatchScorer) LogitsBatch(_ context.Context, exs []model.Example) ([][]float32, error) {
	rows := make([][]float32, len(exs))
	for i, ex := range exs {
		row := make([]float32, len(ex.Options))
		for j := range row {
			row[j] = float32(math.Inf(-1))
		}
		rows[i] = row
	}
	return rows, nil
}

func TestBuildAllNegativeInfLogitsNoNaN(t *testing.T) {
	plan, err := Build(context.Background(), infBatchScorer{}, Request{
		Fields:   []Field{{Label: "Phone"}, {Label: "Email"}},
		Entities: entitiesFor("Phone", "(503) 555-0142"),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, d := range plan.Decisions {
		if math.IsNaN(d.Probability) {
			t.Fatalf("decision %+v has NaN probability", d)
		}
	}
}
