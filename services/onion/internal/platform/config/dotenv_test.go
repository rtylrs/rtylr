package config

import "testing"

func TestListenPortUsesEnvWhenSet(t *testing.T) {
	t.Setenv("PORT", "4242")
	if got := ListenPort("9999"); got != "4242" {
		t.Fatalf("ListenPort = %q, want 4242", got)
	}
}

func TestListenPortFallsBackWhenUnset(t *testing.T) {
	t.Setenv("PORT", "")
	if got := ListenPort("8080"); got != "8080" {
		t.Fatalf("ListenPort = %q, want 8080", got)
	}
}
