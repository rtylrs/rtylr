package application

import (
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
	if large["truncated"] != true || large["notice"] == nil {
		t.Fatalf("large output was not bounded: %#v", large)
	}
}
