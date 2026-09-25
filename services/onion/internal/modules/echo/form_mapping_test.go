package echo

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/echo/internal/application"
)

func TestNewFormMappingToolExecutorLexicalFallback(t *testing.T) {
	exec, err := NewFormMappingToolExecutor("")
	if err != nil {
		t.Fatalf("empty path must not error: %v", err)
	}
	if exec == nil {
		t.Fatal("executor is nil")
	}
}

func TestNewFormMappingToolExecutorLoadsCheckpoint(t *testing.T) {
	exec, err := NewFormMappingToolExecutor("../../ai/classifier/testdata/forms")
	if err != nil {
		t.Fatalf("Load fixture checkpoint: %v", err)
	}
	out, executed, err := exec.ExecuteCaretTool(context.Background(), application.DocumentFormMappingToolName, map[string]any{
		"document": "Phone number: (503) 555-0142",
		"fields":   []any{map[string]any{"id": "phone", "label": "Phone number"}},
	})
	if err != nil {
		t.Fatalf("ExecuteCaretTool: %v", err)
	}
	if !executed {
		t.Fatal("executor did not claim the tool")
	}
	if out["scorer"] != "s1-forms" {
		t.Fatalf("scorer = %v, want s1-forms", out["scorer"])
	}
	if out["field_count"] != 1 {
		t.Fatalf("field_count = %v", out["field_count"])
	}
}

func TestNewFormMappingToolExecutorBadCheckpointFallsBack(t *testing.T) {
	exec, err := NewFormMappingToolExecutor("../../ai/classifier/testdata/does-not-exist")
	if err == nil {
		t.Fatal("expected an error describing the fallback")
	}
	if exec == nil {
		t.Fatal("executor must still be usable on fallback")
	}
}

func TestNewFormMappingToolExecutorSanitizesError(t *testing.T) {
	exec, err := NewFormMappingToolExecutor("/definitely/not/here/cua-s1-forms")
	if err == nil {
		t.Fatal("expected an error describing the fallback")
	}
	out, executed, err := exec.ExecuteCaretTool(context.Background(), application.DocumentFormMappingToolName, map[string]any{
		"document": "Phone: (503) 555-0142",
		"fields":   []any{map[string]any{"label": "Phone"}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	modelInfo, ok := out["model"].(map[string]any)
	if !ok {
		t.Fatalf("model info has type %T", out["model"])
	}
	detail, _ := modelInfo["error"].(string)
	if detail != "not_found" {
		t.Fatalf("error kind = %q, want not_found", detail)
	}
	if strings.ContainsAny(detail, "/\\") {
		t.Fatalf("path leaked into tool output: %q", detail)
	}
}
