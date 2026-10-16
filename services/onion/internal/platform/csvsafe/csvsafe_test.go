package csvsafe

import "testing"

func TestCell(t *testing.T) {
	cases := map[string]string{
		"":                     "",
		"Jane Doe":             "Jane Doe",
		"=HYPERLINK(\"x\")":    "'=HYPERLINK(\"x\")",
		"+1 555":               "'+1 555",
		"-2+3":                 "'-2+3",
		"@SUM(A1)":             "'@SUM(A1)",
		"\tcmd":                "'\tcmd",
		"\rcmd":                "'\rcmd",
		"a=b":                  "a=b",
		"2026-10-16T00:00:00Z": "2026-10-16T00:00:00Z",
	}
	for in, want := range cases {
		if got := Cell(in); got != want {
			t.Errorf("Cell(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRowDoesNotMutateInput(t *testing.T) {
	in := []string{"=1", "ok"}
	out := Row(in...)
	if out[0] != "'=1" || out[1] != "ok" || in[0] != "=1" {
		t.Fatalf("unexpected result %v / %v", out, in)
	}
}
