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

func TestEveryWriteToolRequiresApproval(t *testing.T) {
	for _, tool := range CoreTools {
		if tool.Scope == "write" && !tool.RequiresApproval && !tool.Personal {
			t.Errorf("write tool %q must require approval", tool.Name)
		}
		// Skipping approval is only for rearranging the person's own Portal.
		if tool.Personal && tool.Module != "portal" {
			t.Errorf("tool %q changes more than the person's own Portal view, so it can't skip approval", tool.Name)
		}
	}
}
