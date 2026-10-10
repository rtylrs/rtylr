package application

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNormalizeToolInputCanonicalizesAliasesAndWriteWrappers(t *testing.T) {
	got := NormalizeToolInput("crm.create_customer", map[string]any{
		"firstName": "Ada",
		"fields":    map[string]any{"lastName": "Lovelace"},
	})
	if got["first_name"] != "Ada" || got["last_name"] != "Lovelace" {
		t.Fatalf("normalized input = %#v", got)
	}
}

func TestSanitizeToolOutputRedactsSecretsAndBoundsLargePayloads(t *testing.T) {
	got := SanitizeToolOutput(map[string]any{
		"customer": map[string]any{"name": "Ada", "access_token": "secret"},
	})
	customer, ok := got["customer"].(map[string]any)
	if !ok || customer["access_token"] != "[redacted]" {
		t.Fatalf("secret was not redacted: %#v", got)
	}
	large := SanitizeToolOutput(map[string]any{"rows": strings.Repeat("x", MaxToolOutputBytes+100)})
	if large["truncated"] != true || !strings.HasPrefix(large["rows"].(string), "xxxx") {
		t.Fatalf("large output was not shortened: %#v", large)
	}
	encoded, err := json.Marshal(large)
	if err != nil || len(encoded) > MaxToolOutputBytes {
		t.Fatalf("shortened output is %d bytes", len(encoded))
	}
}

func TestSanitizeToolOutputKeepsTheStartOfLongListsAndPages(t *testing.T) {
	rows := make([]map[string]any, 500)
	for i := range rows {
		rows[i] = map[string]any{"name": fmt.Sprintf("customer %d", i), "notes": strings.Repeat("ب", 200)}
	}
	got := SanitizeToolOutput(map[string]any{"customers": rows, "total": 500})
	list, ok := got["customers"].([]any)
	if !ok || len(list) < 6 || got["total"] != 500.0 || got["truncated"] != true {
		t.Fatalf("list output = %#v", got)
	}
	if first := list[0].(map[string]any); first["name"] != "customer 0" {
		t.Fatalf("first row = %#v", first)
	}
	if last, _ := list[len(list)-1].(string); !strings.Contains(last, "more not shown") {
		t.Fatalf("missing the omitted-count marker: %#v", list[len(list)-1])
	}
}
