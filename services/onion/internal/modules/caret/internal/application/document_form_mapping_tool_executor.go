package application

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/voxire/onion/internal/ai/classifier/planner"
)

// DocumentFormMappingToolName maps values from a document onto form fields.
const DocumentFormMappingToolName = "documents.map_form_fields"

const (
	documentFormMappingMaxDocumentBytes = 8000
	documentFormMappingMaxFields        = 32
	documentFormMappingMaxEntities      = 64
	documentFormMappingMaxFieldOptions  = 1200
)

// DocumentFormMappingToolExecutor plans document-to-form field mapping with
// a classifier scorer or the lexical fallback.
type DocumentFormMappingToolExecutor struct {
	scorer    planner.Scorer
	modelInfo map[string]any
}

// NewDocumentFormMappingToolExecutor builds the executor; a nil scorer
// selects the lexical fallback.
func NewDocumentFormMappingToolExecutor(scorer planner.Scorer, modelInfo map[string]any) *DocumentFormMappingToolExecutor {
	if scorer == nil {
		scorer = planner.LexicalScorer{}
	}
	if modelInfo == nil {
		modelInfo = map[string]any{"kind": "lexical"}
	}
	return &DocumentFormMappingToolExecutor{scorer: scorer, modelInfo: modelInfo}
}

// ExecuteCaretTool implements ToolExecutor.
func (e *DocumentFormMappingToolExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != DocumentFormMappingToolName {
		return nil, false, nil
	}
	req, err := parseDocumentFormMappingInput(input)
	if err != nil {
		return nil, true, err
	}
	entities := req.Entities
	if len(entities) == 0 && !req.EntitiesProvided {
		entities = planner.ExtractEntities(req.Document, documentFormMappingMaxEntities)
	}
	if len(req.Fields)*(len(entities)+3) > documentFormMappingMaxFieldOptions {
		return nil, true, fmt.Errorf("form is too large for the classifier: %d fields with %d values exceeds the %d field-value budget; split the form or trim the document", len(req.Fields), len(entities), documentFormMappingMaxFieldOptions)
	}
	plan, err := planner.Build(ctx, e.scorer, planner.Request{
		Form:        req.Form,
		Fields:      req.Fields,
		Entities:    entities,
		Threshold:   req.Threshold,
		AllowSubmit: req.AllowSubmit,
	})
	if err != nil {
		return nil, true, err
	}
	return e.renderOutput(plan, len(entities), len(req.Document) > 0), true, nil
}

type documentFormMappingInput struct {
	Document         string
	Entities         []planner.Entity
	EntitiesProvided bool
	Form             string
	Fields           []planner.Field
	Threshold        float64
	AllowSubmit      bool
}

func parseDocumentFormMappingInput(input map[string]any) (documentFormMappingInput, error) {
	var req documentFormMappingInput
	req.Document = strings.TrimSpace(stringValue(input["document"]))
	if len(req.Document) > documentFormMappingMaxDocumentBytes {
		return req, fmt.Errorf("document exceeds %d bytes; send a trimmed excerpt or pre-extracted entities", documentFormMappingMaxDocumentBytes)
	}
	req.Form = planner.NormalizeForm(stringValue(input["form"]))
	if raw, ok := input["threshold"]; ok {
		t, ok := numberValue(raw)
		if !ok || math.IsNaN(t) || math.IsInf(t, 0) || t < 0 || t > 1 {
			return req, fmt.Errorf("threshold must be a number between 0 and 1")
		}
		req.Threshold = t
	}
	req.AllowSubmit = boolValue(input["allow_submit"])

	req.EntitiesProvided = input["entities"] != nil
	entities, err := parseEntities(input["entities"])
	if err != nil {
		return req, err
	}
	req.Entities = entities

	fields, err := parseFields(input["fields"])
	if err != nil {
		return req, err
	}
	req.Fields = fields

	if req.Document == "" && !req.EntitiesProvided {
		return req, fmt.Errorf("provide document text or pre-extracted entities")
	}
	if len(req.Fields) == 0 {
		return req, fmt.Errorf("provide at least one form field with a label")
	}
	return req, nil
}

func parseEntities(raw any) ([]planner.Entity, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("entities must be an array of \"Label: value\" strings or {label,value} objects")
	}
	if len(items) > documentFormMappingMaxEntities {
		return nil, fmt.Errorf("entities exceed the %d item limit", documentFormMappingMaxEntities)
	}
	out := make([]planner.Entity, 0, len(items))
	for i, item := range items {
		switch v := item.(type) {
		case string:
			label, value, found := strings.Cut(strings.TrimSpace(v), ":")
			if !found {
				return nil, fmt.Errorf("entities[%d] must look like \"Label: value\"", i)
			}
			if planner.IsURLNoise(strings.TrimSpace(label), strings.TrimSpace(value)) {
				continue
			}
			entity, ok := planner.NormalizeEntity(planner.Entity{Label: label, Value: value})
			if !ok {
				return nil, fmt.Errorf("entities[%d] needs a usable label and value", i)
			}
			out = append(out, entity)
		case map[string]any:
			label := strings.TrimSpace(stringValue(v["label"]))
			value := strings.TrimSpace(stringValue(v["value"]))
			if planner.IsURLNoise(label, value) {
				continue
			}
			entity, ok := planner.NormalizeEntity(planner.Entity{Label: label, Value: value})
			if !ok {
				return nil, fmt.Errorf("entities[%d] needs a usable label and value", i)
			}
			out = append(out, entity)
		default:
			return nil, fmt.Errorf("entities[%d] must be a string or object", i)
		}
	}
	return out, nil
}

func parseFields(raw any) ([]planner.Field, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("fields must be an array of {id,label,role,value} objects")
	}
	if len(items) > documentFormMappingMaxFields {
		return nil, fmt.Errorf("fields exceed the %d item limit", documentFormMappingMaxFields)
	}
	out := make([]planner.Field, 0, len(items))
	for i, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("fields[%d] must be an object", i)
		}
		field, ok := planner.NormalizeField(planner.Field{
			ID:    stringValue(obj["id"]),
			Label: stringValue(obj["label"]),
			Role:  stringValue(obj["role"]),
			Value: stringValue(obj["value"]),
		})
		if !ok {
			return nil, fmt.Errorf("fields[%d] needs a non-empty label", i)
		}
		out = append(out, field)
	}
	return out, nil
}

func (e *DocumentFormMappingToolExecutor) renderOutput(plan planner.Plan, entityCount int, extracted bool) map[string]any {
	counts := map[string]int{}
	fields := make([]map[string]any, 0, len(plan.Decisions))
	for _, d := range plan.Decisions {
		counts[d.Action]++
		entry := map[string]any{
			"label":        d.FieldLabel,
			"action":       d.Action,
			"option_index": d.OptionIndex,
			"probability":  math.Round(d.Probability*10000) / 10000,
		}
		if d.FieldID != "" {
			entry["id"] = d.FieldID
		}
		if d.Action == planner.ActionFill {
			entry["value"] = d.Value
		}
		if d.NeedsReview {
			entry["needs_review"] = true
		}
		if d.AlreadySet {
			entry["already_set"] = true
		}
		fields = append(fields, entry)
	}

	kind, _ := e.modelInfo["kind"].(string)
	out := map[string]any{
		"summary":                planSummary(kind, plan, counts, entityCount, extracted),
		"scorer":                 kind,
		"model":                  e.modelInfo,
		"field_count":            len(plan.Decisions),
		"fill_count":             counts[planner.ActionFill],
		"check_count":            counts[planner.ActionCheck],
		"skip_count":             counts[planner.ActionSkip],
		"needs_review_count":     countsNeedsReview(plan),
		"extracted_entity_count": entityCount,
		"apply_order":            "fills, then checks, then any approved submit click; skips are not applied",
		"fields":                 fields,
	}
	if entityCount == 0 {
		out["notice"] = "No \"Label: value\" pairs were found in the document; every field needs manual input."
	}
	if plan.Submit != nil {
		out["submit"] = map[string]any{
			"label":             plan.Submit.FieldLabel,
			"action":            planner.ActionClick,
			"probability":       math.Round(plan.Submit.Probability*10000) / 10000,
			"requires_approval": true,
		}
	}
	return out
}

func countsNeedsReview(plan planner.Plan) int {
	n := 0
	for _, d := range plan.Decisions {
		if d.NeedsReview {
			n++
		}
	}
	return n
}

func planSummary(kind string, plan planner.Plan, counts map[string]int, entityCount int, extracted bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Planned %d of %d fields", counts[planner.ActionFill]+counts[planner.ActionCheck], len(plan.Decisions))
	if extracted {
		fmt.Fprintf(&b, " from %d extracted values", entityCount)
	} else {
		fmt.Fprintf(&b, " from %d provided values", entityCount)
	}
	fmt.Fprintf(&b, " (%d fills, %d checks, %d skips", counts[planner.ActionFill], counts[planner.ActionCheck], counts[planner.ActionSkip])
	if review := countsNeedsReview(plan); review > 0 {
		fmt.Fprintf(&b, ", %d need review", review)
	}
	b.WriteString(")")
	if plan.Submit != nil {
		b.WriteString("; one submit click is ready for approval")
	}
	if kind != "" {
		fmt.Fprintf(&b, ". Scorer: %s.", kind)
	}
	return b.String()
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func numberValue(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func boolValue(v any) bool {
	b, _ := v.(bool)
	return b
}
