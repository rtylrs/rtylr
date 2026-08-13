package handlers

import "testing"

func TestExtractToolCallsParsesNestedInputWithoutLeakingValidMarkers(t *testing.T) {
	input := `before <tool>{"name":"crm.create_customer","input":{"profile":{"first_name":"Ada"}}}</tool> after`
	calls, remainder := ExtractToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("calls = %#v", calls)
	}
	profile, ok := calls[0].Input["profile"].(map[string]any)
	if !ok || profile["first_name"] != "Ada" {
		t.Fatalf("calls = %#v", calls)
	}
	if remainder != "before  after" {
		t.Fatalf("remainder = %q", remainder)
	}
}

func TestExtractToolCallsPreservesMalformedMarkerForInspection(t *testing.T) {
	input := `before <tool>{not-json}</tool> after`
	calls, remainder := ExtractToolCalls(input)
	if len(calls) != 0 || remainder != input {
		t.Fatalf("calls=%#v remainder=%q", calls, remainder)
	}
}
