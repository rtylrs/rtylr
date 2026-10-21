package application

import (
	"context"
	"sync"
	"testing"
)

func TestBusinessFindSearchesEveryAppAndFiltersMatches(t *testing.T) {
	asked := map[string]any{}
	var mu sync.Mutex
	executor := toolExecutorFunc(func(_ context.Context, name string, input map[string]any) (map[string]any, bool, error) {
		mu.Lock()
		asked[name] = input["search"]
		mu.Unlock()
		switch name {
		case "crm.list_customers":
			return map[string]any{"customers": []any{map[string]any{"id": "c1", "name": "Rana Haddad", "email": "rana@x.co"}, map[string]any{"id": "c2", "name": "Omar"}}}, true, nil
		case "hr.list_employees":
			// This tool ignores the filter; business.find filters for it.
			return map[string]any{"employees": []map[string]any{{"id": 5, "first_name": "Rana", "last_name": "K", "role": "Barista"}, {"id": 6, "first_name": "Sami"}}}, true, nil
		}
		return map[string]any{"items": []any{}}, true, nil
	})
	find := &BusinessFindExecutor{Executors: func() []ToolExecutor { return []ToolExecutor{executor} }}
	output, _, err := find.ExecuteCaretTool(context.Background(), BusinessFindToolName, map[string]any{"query": "rana"})
	if err != nil {
		t.Fatal(err)
	}
	matches := output["matches"].([]map[string]any)
	if len(matches) != 2 || matches[0]["kind"] != "customer" || matches[0]["details_tool"] != "crm.get_customer" || matches[1]["name"] != "Rana K" || matches[1]["role"] != "Barista" {
		t.Fatalf("matches = %#v", matches)
	}
	if asked["crm.list_customers"] != "rana" || len(asked) != len(findSources) {
		t.Fatalf("asked = %#v", asked)
	}
	if _, _, err := find.ExecuteCaretTool(context.Background(), BusinessFindToolName, map[string]any{"query": "r"}); err == nil {
		t.Fatal("a one-letter query was accepted")
	}
}

type toolExecutorFunc func(context.Context, string, map[string]any) (map[string]any, bool, error)

func (f toolExecutorFunc) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	return f(ctx, name, input)
}
