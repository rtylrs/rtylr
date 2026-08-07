package api

import (
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/platform/hubprofile"
)

func TestProfileToGetMeResponseIncludesConsoleAdmin(t *testing.T) {
	response := profileToGetMeResponse(hubprofile.Profile{
		ID:             1,
		IsConsoleAdmin: true,
	})

	if !response.GetIsConsoleAdmin() {
		t.Fatal("expected console-admin flag to be included in GetMe response")
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal GetMe response: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal GetMe response: %v", err)
	}
	if isConsoleAdmin, ok := decoded["is_console_admin"].(bool); !ok || !isConsoleAdmin {
		t.Fatalf("expected wire response is_console_admin=true, got %#v", decoded["is_console_admin"])
	}
}
