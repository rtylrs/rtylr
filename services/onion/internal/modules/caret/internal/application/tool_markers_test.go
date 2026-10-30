package application

import (
	"encoding/json"
	"testing"
)

const nativeLayoutCall = "<｜DSML｜function_calls>\n<｜DSML｜invoke name=\"portal.layout\">\n\n</｜DSML｜invoke>\n</｜DSML｜function_calls>"

func TestParseToolMarkersReadsNativeCalls(t *testing.T) {
	calls, rest := parseToolMarkers("One moment. " + nativeLayoutCall)
	if len(calls) != 1 || calls[0].Name != "portal.layout" || rest != "One moment. " {
		t.Fatalf("calls = %v, rest = %q", calls, rest)
	}
}

// Native markup inside a <tool> input is what the person wants saved, not a
// call to make.
func TestParseToolMarkersKeepsNativeMarkupInsideAnInput(t *testing.T) {
	note := "Example: " + nativeLayoutCall
	calls, _ := parseToolMarkers(`<tool>{"name":"memory.remember","input":{"value":` + jsonString(note) + `}}</tool>`)
	if len(calls) != 1 || calls[0].Name != "memory.remember" || calls[0].Input["value"] != note {
		t.Fatalf("calls = %#v", calls)
	}
}

func jsonString(s string) string {
	raw, _ := json.Marshal(s)
	return string(raw)
}
