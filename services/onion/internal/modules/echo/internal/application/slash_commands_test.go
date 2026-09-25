package application

import (
	"strings"
	"testing"
)

func TestParseCaretSlashCommand(t *testing.T) {
	command, args, ok := ParseCaretSlashCommand("  /GOAL launch the new store  ")
	if !ok {
		t.Fatal("expected /goal to be recognized")
	}
	if command.Name != "goal" || command.Mode != ModeAsk {
		t.Fatalf("command = %#v", command)
	}
	if args != "launch the new store" {
		t.Fatalf("args = %q", args)
	}

	if _, _, ok := ParseCaretSlashCommand("/unknown thing"); ok {
		t.Fatal("unknown slash command should remain a normal prompt")
	}
}

func TestNormalizeCaretSlashCommand(t *testing.T) {
	got, command, ok := NormalizeCaretSlashCommand("/goal launch the new store")
	if !ok || command.Name != "goal" {
		t.Fatalf("normalize result = %q, %#v, %v", got, command, ok)
	}
	for _, want := range []string{"The user invoked /goal", "measurable goal", "launch the new store"} {
		if !strings.Contains(got, want) {
			t.Fatalf("normalized prompt missing %q: %s", want, got)
		}
	}

	got, command, ok = NormalizeCaretSlashCommand("/remember brunch gets busy after 11am")
	if !ok || command.Mode != ModeAgent || !strings.Contains(got, "memory.remember") {
		t.Fatalf("remember result = %q, %#v, %v", got, command, ok)
	}
}

func TestNormalizeCaretSlashCommandShowsUsageWhenArgumentsAreMissing(t *testing.T) {
	got, command, ok := NormalizeCaretSlashCommand("/goal")
	if !ok || command.Name != "goal" {
		t.Fatalf("result = %q, %#v, %v", got, command, ok)
	}
	if !strings.Contains(got, "/goal <outcome>") || !strings.Contains(got, "missing input") {
		t.Fatalf("usage prompt = %q", got)
	}
}

func TestCaretSlashCommandHelpIsUseful(t *testing.T) {
	got := CaretSlashCommandHelp()
	for _, want := range []string{"/goal <outcome>", "/brief", "/status", "/remember <fact>", "/help"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help missing %q: %s", want, got)
		}
	}
}
