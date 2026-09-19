package planner

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/ai/classifier/model"
)

// Actions a field decision can carry.
const (
	ActionFill  = "fill"
	ActionCheck = "check"
	ActionClick = "click"
	ActionSkip  = "skip"
)

// DefaultThreshold is the minimum probability a decision needs before it
// leaves review.
const DefaultThreshold = 0.5

// Field describes one form element to fill.
type Field struct {
	ID    string
	Label string
	Role  string
	Value string
}

// Scorer is the decision layer: one probability per option.
type Scorer interface {
	Score(ctx context.Context, ex model.Example) (probs []float64, best int, err error)
}

type batchScorer interface {
	LogitsBatch(ctx context.Context, exs []model.Example) ([][]float32, error)
}

// Request is one planning job: a form, its fields, and the document
// entities extracted upstream.
type Request struct {
	Form        string
	Fields      []Field
	Entities    []Entity
	Threshold   float64
	AllowSubmit bool
}

// Decision is the plan for a single field.
type Decision struct {
	FieldIndex  int
	FieldID     string
	FieldLabel  string
	Action      string
	Value       string
	OptionIndex int
	Probability float64
	NeedsReview bool
	AlreadySet  bool
}

// Plan is an ordered, reviewable set of field decisions: fills, then
// checks, then any approved submit click, then skips.
type Plan struct {
	Decisions []Decision
	Submit    *Decision
	Entities  int
}

var fixedActions = [3]string{ActionCheck, ActionClick, ActionSkip}

// Build scores every field against the candidate entities and returns an
// ordered plan without mutating anything.
func Build(ctx context.Context, sc Scorer, req Request) (Plan, error) {
	if err := ctx.Err(); err != nil {
		return Plan{}, err
	}
	if sc == nil {
		return Plan{}, fmt.Errorf("planner: a scorer is required")
	}
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	if threshold > 1 {
		threshold = 1
	}

	plan := Plan{Entities: len(req.Entities)}
	if len(req.Entities) == 0 {
		for i, f := range req.Fields {
			plan.Decisions = append(plan.Decisions, skipDecision(i, f, true))
		}
		return plan, nil
	}

	options := make([][]string, len(req.Fields))
	examples := make([]model.Example, len(req.Fields))
	for i, f := range req.Fields {
		options[i] = optionsFor(req.Entities)
		examples[i] = model.Example{Context: elementContext(req.Form, f), Options: options[i]}
	}
	rows, err := scoreAll(ctx, sc, examples)
	if err != nil {
		return Plan{}, err
	}
	if len(rows) != len(req.Fields) {
		return Plan{}, fmt.Errorf("planner: scorer returned %d rows for %d fields", len(rows), len(req.Fields))
	}

	decisions := make([]Decision, len(req.Fields))
	for i, f := range req.Fields {
		if err := ctx.Err(); err != nil {
			return Plan{}, err
		}
		probs := rows[i]
		if len(probs) == 0 {
			decisions[i] = skipDecision(i, f, true)
			continue
		}
		best := 0
		for j, p := range probs {
			if p > probs[best] {
				best = j
			}
		}
		decisions[i] = decide(i, f, req.Entities, probs, best, threshold)
	}

	submitFieldIndex := selectSubmit(decisions, req)
	plan.Decisions = orderDecisions(decisions)
	if submitFieldIndex >= 0 {
		for i := range plan.Decisions {
			if plan.Decisions[i].FieldIndex == submitFieldIndex {
				submit := plan.Decisions[i]
				plan.Submit = &submit
				break
			}
		}
	}
	return plan, nil
}

func decide(index int, f Field, entities []Entity, probs []float64, best int, threshold float64) Decision {
	if len(probs) == 0 {
		return skipDecision(index, f, true)
	}
	d := Decision{
		FieldIndex:  index,
		FieldID:     f.ID,
		FieldLabel:  f.Label,
		OptionIndex: best,
		Probability: probs[best],
	}
	n := len(entities)
	switch {
	case best < n:
		d.Action = ActionFill
		d.Value = entities[best].Value
		if d.Value != "" && strings.EqualFold(strings.TrimSpace(d.Value), strings.TrimSpace(f.Value)) {
			d.Action = ActionSkip
			d.AlreadySet = true
			return d
		}
	case best == n:
		d.Action = ActionCheck
	case best == n+1:
		d.Action = ActionClick
	default:
		d.Action = ActionSkip
	}
	if d.Probability < threshold {
		d.Action = ActionSkip
		d.NeedsReview = true
	}
	return d
}

func selectSubmit(decisions []Decision, req Request) int {
	bestIndex := -1
	for i := range decisions {
		d := &decisions[i]
		if d.Action != ActionClick {
			continue
		}
		if !isSubmitField(req.Fields[d.FieldIndex]) || !req.AllowSubmit || d.Probability < thresholdOr(req.Threshold) {
			demoteClick(d)
			continue
		}
		if bestIndex == -1 || d.Probability > decisions[bestIndex].Probability {
			bestIndex = i
		}
	}
	if bestIndex == -1 {
		return -1
	}
	submitFieldIndex := decisions[bestIndex].FieldIndex
	for i := range decisions {
		d := &decisions[i]
		if d.Action == ActionClick && i != bestIndex {
			demoteClick(d)
		}
	}
	return submitFieldIndex
}

// demoteClick turns a disallowed or superseded click into a review skip and
// clears the stale option metadata so clients cannot read "skip (p=0.9)".
func demoteClick(d *Decision) {
	d.Action = ActionSkip
	d.OptionIndex = -1
	d.Probability = 0
	d.NeedsReview = true
}

func thresholdOr(threshold float64) float64 {
	if threshold <= 0 {
		return DefaultThreshold
	}
	if threshold > 1 {
		return 1
	}
	return threshold
}

func isSubmitField(f Field) bool {
	role := strings.ToLower(strings.TrimSpace(f.Role))
	if role != "" && role != "button" && role != "axbutton" {
		return false
	}
	label := normalizeLabel(f.Label)
	return label == "submit" || label == "submit form"
}

func normalizeLabel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, ":")
	return strings.Join(strings.Fields(s), " ")
}

func orderDecisions(decisions []Decision) []Decision {
	rank := map[string]int{ActionFill: 0, ActionCheck: 1, ActionClick: 2, ActionSkip: 3}
	out := make([]Decision, len(decisions))
	copy(out, decisions)
	sort.SliceStable(out, func(i, j int) bool {
		return rank[out[i].Action] < rank[out[j].Action]
	})
	return out
}

func skipDecision(index int, f Field, needsReview bool) Decision {
	return Decision{
		FieldIndex:  index,
		FieldID:     f.ID,
		FieldLabel:  f.Label,
		Action:      ActionSkip,
		OptionIndex: -1,
		NeedsReview: needsReview,
	}
}

func optionsFor(entities []Entity) []string {
	out := make([]string, 0, len(entities)+len(fixedActions))
	for _, e := range entities {
		out = append(out, "fill "+e.Label+": "+e.Value)
	}
	out = append(out, fixedActions[:]...)
	return out
}

func elementContext(form string, f Field) string {
	role := strings.TrimSpace(f.Role)
	if role == "" {
		role = "Edit"
	}
	name := NormalizeForm(form)
	if name == "" {
		name = "Form"
	}
	return "TASK fill the form from the document, then submit\n" +
		"FORM " + cleanField(name) + "\n" +
		"ELEMENT " + cleanField(role) + " \"" + cleanField(f.Label) + "\" value=\"" + cleanField(f.Value) + "\""
}

func cleanField(s string) string {
	s = strings.NewReplacer("\"", "'", "\n", " ", "\r", " ", "\t", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func scoreAll(ctx context.Context, sc Scorer, examples []model.Example) ([][]float64, error) {
	if bs, ok := sc.(batchScorer); ok {
		logits, err := bs.LogitsBatch(ctx, examples)
		if err != nil {
			return nil, err
		}
		rows := make([][]float64, len(logits))
		for i, row := range logits {
			rows[i] = softmax(row)
		}
		return rows, nil
	}
	rows := make([][]float64, len(examples))
	for i, ex := range examples {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		probs, _, err := sc.Score(ctx, ex)
		if err != nil {
			return nil, err
		}
		rows[i] = probs
	}
	return rows, nil
}

func softmax(logits []float32) []float64 {
	if len(logits) == 0 {
		return nil
	}
	maxLogit := logits[0]
	for _, v := range logits[1:] {
		if v > maxLogit {
			maxLogit = v
		}
	}
	out := make([]float64, len(logits))
	var sum float64
	for i, v := range logits {
		d := float64(v - maxLogit)
		if math.IsNaN(d) {
			d = 0
		}
		e := math.Exp(d)
		out[i] = e
		sum += e
	}
	if sum == 0 {
		uniform := 1 / float64(len(out))
		for i := range out {
			out[i] = uniform
		}
		return out
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}
