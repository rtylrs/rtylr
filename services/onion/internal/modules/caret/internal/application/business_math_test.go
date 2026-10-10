package application

import (
	"context"
	"testing"
)

func TestBusinessMath(t *testing.T) {
	e := &AssistantToolExecutor{}
	run := func(input map[string]any) map[string]any {
		t.Helper()
		output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.business_math", input)
		if err != nil {
			t.Fatalf("%v: %v", input, err)
		}
		return output
	}
	if out := run(map[string]any{"operation": "margin", "cost": 6.0, "price": 10.0}); out["margin_percent"] != 40.0 || out["markup_percent"] != 66.67 {
		t.Fatalf("margin = %#v", out)
	}
	if out := run(map[string]any{"operation": "price for margin", "cost": 6.0, "margin_percent": 40.0}); out["price"] != 10.0 {
		t.Fatalf("price_for_margin = %#v", out)
	}
	if out := run(map[string]any{"operation": "vat", "amount": 111.0, "rate_percent": 11.0, "inclusive": true}); out["net"] != 100.0 || out["tax"] != 11.0 {
		t.Fatalf("vat inclusive = %#v", out)
	}
	if out := run(map[string]any{"operation": "loan_payment", "principal": 20000.0, "annual_rate_percent": 7.0, "months": 36.0}); out["monthly_payment"] != 617.54 {
		t.Fatalf("loan = %#v", out)
	}
	if out := run(map[string]any{"operation": "break_even", "fixed_costs": 5000.0, "price": 25.0, "variable_cost": 10.0}); out["break_even_units"] != 334.0 {
		t.Fatalf("break_even = %#v", out)
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "utilities.business_math", map[string]any{"operation": "astrology"}); err == nil {
		t.Fatal("unknown operation accepted")
	}
}
