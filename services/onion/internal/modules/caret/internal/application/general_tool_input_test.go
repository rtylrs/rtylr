package application

import "testing"

func TestGeneralToolInputHelpersAcceptAliasesAndStrings(t *testing.T) {
	input := map[string]any{"location": " Beirut ", "amount": "1,250.5", "values": "1, 2;3 4", "list": []any{1.0, "2", "x"}, "flag": "true"}
	if got := inputText(input, "city", "location"); got != "Beirut" {
		t.Fatalf("inputText = %q", got)
	}
	if got, ok := inputNumber(input, "amount"); !ok || got != 1250.5 {
		t.Fatalf("inputNumber = %v, %v", got, ok)
	}
	if got := inputNumbers(input, "values"); len(got) != 4 || got[3] != 4 {
		t.Fatalf("inputNumbers(string) = %v", got)
	}
	if got := inputNumbers(input, "list"); len(got) != 2 || got[1] != 2 {
		t.Fatalf("inputNumbers(array) = %v", got)
	}
	if !inputBool(input, "flag") || inputBool(input, "missing") {
		t.Fatal("inputBool misread")
	}
}
