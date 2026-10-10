package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// scriptedCompleter answers each sub-agent from a script keyed by a word in
// its task, one reply per step.
type scriptedCompleter struct {
	mu      sync.Mutex
	scripts map[string][]string
	steps   map[string]int
	prompts []string
}

func (c *scriptedCompleter) CompleteText(_ context.Context, prompt string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prompts = append(c.prompts, prompt)
	for key, replies := range c.scripts {
		if strings.Contains(prompt, "<task>"+key) {
			step := c.steps[key]
			c.steps[key]++
			if step >= len(replies) {
				return "", errors.New("script exhausted")
			}
			return replies[step], nil
		}
	}
	return "", errors.New("no script")
}

func TestSubAgentsResearchInParallelAndReport(t *testing.T) {
	var searches atomic.Int32
	web := publicWebSearcherFunc(func(_ context.Context, query string) (map[string]any, error) {
		searches.Add(1)
		return map[string]any{"sources": []map[string]string{{"title": query, "url": "https://example.com/" + strings.ReplaceAll(query, " ", "-")}}}, nil
	})
	assistant := NewAssistantToolExecutor(nil, web)
	ai := &scriptedCompleter{steps: map[string]int{}, scripts: map[string][]string{
		"Alpha": {`<tool>{"name":"web.search","input":{"query":"alpha pricing"}}</tool>`, `<report>Alpha costs $10 (https://example.com/alpha-pricing).</report>`},
		"Beta":  {`<tool>{"name":"web.search","input":{"query":"beta pricing"}}</tool><tool>{"name":"utilities.calculate","input":{"expression":"12*12"}}</tool>`, `<report>Beta costs $144 a year.</report>`},
	}}
	agents := &SubAgentExecutor{AI: ai}
	agents.Executors = func() []ToolExecutor { return []ToolExecutor{assistant, agents} }
	var progress []string
	var mu sync.Mutex
	ctx := WithToolProgress(context.Background(), func(m string) { mu.Lock(); progress = append(progress, m); mu.Unlock() })

	output, executed, err := agents.ExecuteCaretTool(ctx, AgentsToolName, map[string]any{"tasks": []any{
		map[string]any{"task": "Alpha: find the price"}, "Beta: find the price",
	}})
	if err != nil || !executed {
		t.Fatalf("agents.run: %v %v", executed, err)
	}
	reports := output["reports"].([]map[string]any)
	if reports[0]["status"] != "done" || !strings.Contains(reports[0]["report"].(string), "$10") || reports[1]["tool_calls"] != 2 {
		t.Fatalf("reports = %#v", reports)
	}
	if sources := output["sources"].([]map[string]string); len(sources) != 2 {
		t.Fatalf("sources = %#v", sources)
	}
	if searches.Load() != 2 {
		t.Fatalf("searches = %d", searches.Load())
	}
	joined := strings.Join(progress, "\n")
	if !strings.Contains(joined, `Agent 1: searching "alpha pricing"`) || !strings.Contains(joined, "Agent 2 finished") {
		t.Fatalf("progress = %q", joined)
	}
	if !strings.Contains(ai.prompts[len(ai.prompts)-1], "returned:") {
		t.Fatal("the follow-up step did not see the tool results")
	}
}

func TestSubAgentsCannotWriteOrNest(t *testing.T) {
	ai := &scriptedCompleter{steps: map[string]int{}, scripts: map[string][]string{
		"Sneaky": {`<tool>{"name":"crm.create_customer","input":{"name":"X"}}</tool><tool>{"name":"agents.run","input":{"tasks":["again"]}}</tool>`, `<report>Could not act.</report>`},
	}}
	agents := &SubAgentExecutor{AI: ai}
	agents.Executors = func() []ToolExecutor { return []ToolExecutor{agents} }
	output, _, err := agents.ExecuteCaretTool(context.Background(), AgentsToolName, map[string]any{"tasks": []any{"Sneaky task"}})
	if err != nil {
		t.Fatal(err)
	}
	second := ai.prompts[1]
	if strings.Count(second, "sub-agents can only use read-only tools") != 2 {
		t.Fatalf("writes and nesting must be refused:\n%s", second)
	}
	if output["reports"].([]map[string]any)[0]["status"] != "done" {
		t.Fatalf("output = %#v", output)
	}
	if _, _, err := agents.ExecuteCaretTool(context.WithValue(context.Background(), subAgentKey{}, true), AgentsToolName, map[string]any{"tasks": []any{"x"}}); err == nil {
		t.Fatal("a sub-agent started sub-agents")
	}
}

func TestSubAgentsMustReportWhenStepsRunOut(t *testing.T) {
	loop := make([]string, maxSubAgentSteps+1)
	for i := range loop {
		loop[i] = `<tool>{"name":"utilities.calculate","input":{"expression":"` + strings.Repeat("1+", i) + `1"}}</tool>`
	}
	loop[maxSubAgentSteps] = "Best effort: nothing conclusive."
	ai := &scriptedCompleter{steps: map[string]int{}, scripts: map[string][]string{"Loop": loop}}
	agents := &SubAgentExecutor{AI: ai}
	assistant := NewAssistantToolExecutor(nil, nil)
	agents.Executors = func() []ToolExecutor { return []ToolExecutor{assistant} }
	output, _, err := agents.ExecuteCaretTool(context.Background(), AgentsToolName, map[string]any{"tasks": []any{"Loop forever"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ai.prompts[len(ai.prompts)-1], "No more tool calls are possible") {
		t.Fatal("the last step must ask for the report")
	}
	report := output["reports"].([]map[string]any)[0]
	if report["report"] != "Best effort: nothing conclusive." || report["steps"] != maxSubAgentSteps+1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestParseSubAgentTasksBoundsTheFanOut(t *testing.T) {
	if _, err := parseSubAgentTasks(map[string]any{}); err == nil {
		t.Fatal("no tasks accepted")
	}
	if _, err := parseSubAgentTasks(map[string]any{"tasks": []any{"a", "b", "c", "d", "e"}}); err == nil {
		t.Fatal("five tasks accepted")
	}
	tasks, err := parseSubAgentTasks(map[string]any{"context": "We sell coffee in Beirut", "tasks": []any{map[string]any{"goal": "find suppliers", "context": "budget $500"}}})
	if err != nil || tasks[0].Task != "find suppliers" || !strings.Contains(tasks[0].Context, "coffee") || !strings.Contains(tasks[0].Context, "$500") {
		t.Fatalf("tasks = %#v, %v", tasks, err)
	}
}
