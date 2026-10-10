package application

import (
	"context"
	"testing"
)

func TestRandomPickDrawsDistinctWinners(t *testing.T) {
	e := &AssistantToolExecutor{}
	from := []any{"a", "b", "c", "d", "e"}
	output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.random", map[string]any{"from": from, "count": 3.0})
	if err != nil {
		t.Fatal(err)
	}
	picked := output["picked"].([]string)
	seen := map[string]bool{}
	for _, p := range picked {
		seen[p] = true
	}
	if len(picked) != 3 || len(seen) != 3 {
		t.Fatalf("picked = %v", picked)
	}
	output, _, err = e.ExecuteCaretTool(context.Background(), "utilities.random", map[string]any{"min": 5.0, "max": 5.0})
	if err != nil || output["number"] != int64(5) {
		t.Fatalf("range = %#v %v", output, err)
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "utilities.random", map[string]any{"from": from, "count": 9.0}); err == nil {
		t.Fatal("more winners than entrants accepted")
	}
}
