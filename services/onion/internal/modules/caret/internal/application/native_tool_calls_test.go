package application

import (
	"reflect"
	"testing"
)

// The markup DeepSeek wrote in a live chat instead of a <tool> marker.
const nativeLayoutCall = "<｜DSML｜function_calls>\n<｜DSML｜invoke name=\"portal.layout\">\n\n</｜DSML｜invoke>\n</｜DSML｜function_calls>"

func TestParseNativeToolCallsReadsDeepSeekMarkup(t *testing.T) {
	text := "<｜DSML｜function_calls>\n" +
		"<｜DSML｜invoke name=\"portal.customize\">\n" +
		"<｜DSML｜parameter name=\"area\" string=\"true\">rail</｜DSML｜parameter>\n" +
		"<｜DSML｜parameter name=\"moves\" string=\"false\">[{\"id\":\"profile\",\"group\":\"top\",\"position\":\"first\"}]</｜DSML｜parameter>\n" +
		"</｜DSML｜invoke>\n" +
		"<｜DSML｜invoke name=\"portal.layout\">\n</｜DSML｜invoke>\n" +
		"</｜DSML｜function_calls>"
	calls, end, ok := ParseNativeToolCalls(text+" after", false)
	if !ok || end != len(text) {
		t.Fatalf("ok = %v, end = %d, want %d", ok, end, len(text))
	}
	want := []NativeToolCall{
		{Name: "portal.customize", Input: map[string]any{
			"area":  "rail",
			"moves": []any{map[string]any{"id": "profile", "group": "top", "position": "first"}},
		}},
		{Name: "portal.layout", Input: map[string]any{}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v", calls)
	}
}

func TestParseNativeToolCallsWaitsForTheWholeBlock(t *testing.T) {
	partial := nativeLayoutCall[:len(nativeLayoutCall)-10]
	if _, _, ok := ParseNativeToolCalls(partial, false); ok {
		t.Fatal("parsed a block that is still arriving")
	}
	// The reply ended before the block closed: the finished invoke still runs.
	calls, end, ok := ParseNativeToolCalls(partial, true)
	if !ok || end != len(partial) || len(calls) != 1 || calls[0].Name != "portal.layout" {
		t.Fatalf("calls = %v, end = %d, ok = %v", calls, end, ok)
	}
}

func TestParseNativeToolCallsAcceptsPlainBars(t *testing.T) {
	calls, _, ok := ParseNativeToolCalls(`<|DSML|tool_calls><|DSML|invoke name="portal.restore_default"><|DSML|parameter name="area" string="true">all</|DSML|parameter></|DSML|invoke></|DSML|tool_calls>`, false)
	if !ok || len(calls) != 1 || calls[0].Name != "portal.restore_default" || calls[0].Input["area"] != "all" {
		t.Fatalf("calls = %v, ok = %v", calls, ok)
	}
}

func TestStripNativeToolCallsKeepsTheProse(t *testing.T) {
	got := StripNativeToolCalls("Checking your layout.\n\n" + nativeLayoutCall + "\nDone.")
	if got != "Checking your layout.\n\n\nDone." {
		t.Fatalf("got %q", got)
	}
}

func TestParseToolMarkersReadsNativeCalls(t *testing.T) {
	calls, rest := parseToolMarkers("One moment. " + nativeLayoutCall)
	if len(calls) != 1 || calls[0].Name != "portal.layout" || rest != "One moment. " {
		t.Fatalf("calls = %v, rest = %q", calls, rest)
	}
}
