package application

import (
	"context"
	"strings"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/snips"
)

type recordingAnalyticsCompletion struct {
	prompt string
}

func (c *recordingAnalyticsCompletion) CompleteTextWithOptions(_ context.Context, prompt string, _ snips.CompletionOptions) (string, error) {
	c.prompt = prompt
	return `{"view":"ai_pos_orders","aggregations":[{"func":"sum","col":"total","as":"revenue"}]}`, nil
}

func TestAnalyticsQueryExecutorKeepsRowsOutOfProviderPrompt(t *testing.T) {
	completion := &recordingAnalyticsCompletion{}
	executor := NewAnalyticsQueryToolExecutor(completion, &snapshotAnalystService{})
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{
		BusinessID: 11,
		UserID:     22,
	})

	output, executed, err := executor.ExecuteCaretTool(ctx, "analytics.query", map[string]any{
		"question": "How much revenue today?",
	})
	if err != nil {
		t.Fatalf("ExecuteCaretTool: %v", err)
	}
	if !executed || output == nil {
		t.Fatalf("expected analytics query execution, output=%#v executed=%v", output, executed)
	}
	if !strings.Contains(completion.prompt, "ai_pos_orders") {
		t.Fatalf("schema-only query planning prompt omitted the allowlisted schema: %s", completion.prompt)
	}
	for _, forbidden := range []string{`"revenue":100`, "Ada Secret", "customer-data"} {
		if strings.Contains(completion.prompt, forbidden) {
			t.Fatalf("provider query-planning prompt received result data %q: %s", forbidden, completion.prompt)
		}
	}
	if rows, ok := output["rows"].([]map[string]any); !ok || len(rows) != 1 || rows[0]["revenue"] != float64(100) {
		t.Fatalf("server result did not retain analytics rows: %#v", output["rows"])
	}
}
