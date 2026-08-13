package application

import "testing"

func TestToolRegistryIsInternallyConsistent(t *testing.T) {
	if err := ValidateToolRegistry(); err != nil {
		t.Fatalf("ValidateToolRegistry: %v", err)
	}
	if len(coreToolIndex) != len(CoreTools) {
		t.Fatalf("tool index has %d entries for %d definitions", len(coreToolIndex), len(CoreTools))
	}
}

func TestToolLookupNormalizesWhitespace(t *testing.T) {
	tool, ok := ToolByName("  business.snapshot ")
	if !ok || tool.Name != "business.snapshot" {
		t.Fatalf("normalized lookup = %#v, %v", tool, ok)
	}
}
