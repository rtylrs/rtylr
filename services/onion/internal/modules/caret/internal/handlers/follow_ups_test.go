package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func streamEvents(t *testing.T, body string) []StreamEvent {
	t.Helper()
	var events []StreamEvent
	for _, line := range strings.Split(body, "\n") {
		payload, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}
		var event StreamEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	return events
}

func TestFollowUpsBecomeTheirOwnEventNotText(t *testing.T) {
	rec := httptest.NewRecorder()
	var full strings.Builder
	renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent))
	reply := `Your revenue grew 12% this month.<followups>["Which products drove it?","Compare with last year","which products drove it?"]</followups>`
	for _, chunk := range []string{reply[:20], reply[20:45], reply[45:]} {
		if err := renderer.AppendText(chunk); err != nil {
			t.Fatal(err)
		}
	}
	if err := renderer.Finish(); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(full.String(), "followups") || strings.Contains(full.String(), "Which products") {
		t.Fatalf("follow-ups leaked into the reply text: %q", full.String())
	}
	var followUps []string
	var order []StreamEventType
	for _, event := range streamEvents(t, rec.Body.String()) {
		order = append(order, event.Type)
		if event.Type == EventFollowUps {
			followUps = event.Items
		}
	}
	if strings.Join(followUps, "|") != "Which products drove it?|Compare with last year" {
		t.Fatalf("follow-ups = %v (deduped, in order)", followUps)
	}
	if len(order) < 3 || order[len(order)-3] != EventFollowUps || order[len(order)-1] != EventDone {
		t.Fatalf("follow-ups must come just before message_end: %v", order)
	}
}

func TestMalformedFollowUpsAreDroppedQuietly(t *testing.T) {
	rec := httptest.NewRecorder()
	var full strings.Builder
	renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent))
	_ = renderer.AppendText(`Done.<followups>not json</followups> <followups>["unfinished`)
	_ = renderer.Finish()
	if strings.Contains(full.String(), "followups") || strings.Contains(full.String(), "unfinished") || strings.Contains(rec.Body.String(), `"follow_ups"`) {
		t.Fatalf("malformed markers must not show: text=%q", full.String())
	}
}
