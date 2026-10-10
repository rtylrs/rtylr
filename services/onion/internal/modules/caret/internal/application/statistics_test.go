package application

import (
	"context"
	"testing"
)

func TestStatisticsSummarisesASeries(t *testing.T) {
	e := &AssistantToolExecutor{}
	output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.statistics", map[string]any{"values": []any{100.0, 110.0, 121.0, 110.0}})
	if err != nil {
		t.Fatal(err)
	}
	if output["mean"] != 110.25 || output["median"] != 110.0 || output["mode"] != 110.0 || output["count"] != 4 {
		t.Fatalf("stats = %#v", output)
	}
	if output["change_first_to_last_percent"] != 10.0 {
		t.Fatalf("change = %v", output["change_first_to_last_percent"])
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "utilities.statistics", map[string]any{"values": "none"}); err == nil {
		t.Fatal("empty list accepted")
	}
}
