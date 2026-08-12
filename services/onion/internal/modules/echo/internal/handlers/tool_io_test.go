package handlers

import (
	"strings"
	"testing"
)

func TestNormalizeToolInputCanonicalizesAliasesAndWriteWrappers(t *testing.T) {
	got := normalizeToolInput("crm.create_customer", map[string]any{
		"firstName": "Ada",
		"fields":    map[string]any{"lastName": "Lovelace"},
	})
	if got["first_name"] != "Ada" || got["last_name"] != "Lovelace" {
		t.Fatalf("normalized input = %#v", got)
	}
}

func TestSanitizeToolOutputRedactsSecretsAndBoundsLargePayloads(t *testing.T) {
	got := sanitizeToolOutput(map[string]any{
		"customer": map[string]any{"name": "Ada", "access_token": "secret"},
	})
	if got["customer"].(map[string]any)["access_token"] != "[redacted]" {
		t.Fatalf("secret was not redacted: %#v", got)
	}
	large := sanitizeToolOutput(map[string]any{"rows": strings.Repeat("x", maxToolOutputBytes+100)})
	if large["truncated"] != true || large["notice"] == nil {
		t.Fatalf("large output was not bounded: %#v", large)
	}
}
