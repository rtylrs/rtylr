package application

import (
	"strings"
	"testing"
)

func TestInputHintsMatchWhatToolsNeed(t *testing.T) {
	for name, want := range map[string]string{
		"crm.get_customer":       `"id"`,
		"insights.get_dashboard": `{}`,
		"insights.get_cash_flow": `"start_date"`,
		"hr.get_leave_balances":  `"employee_id"`,
	} {
		tool, ok := ToolByName(name)
		if !ok || !strings.Contains(tool.InputHint, want) {
			t.Errorf("%s hint = %q, want it to contain %s", name, tool.InputHint, want)
		}
	}
	for _, name := range []string{"web.search", ToolSearchName, "utilities.calculate"} {
		if tool, _ := ToolByName(name); strings.Contains(tool.OutputHint, "tenant-scoped") {
			t.Errorf("%s still claims a tenant-scoped result", name)
		}
	}
}
