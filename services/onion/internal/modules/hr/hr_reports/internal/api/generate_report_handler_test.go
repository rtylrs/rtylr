package api

import "testing"

func TestSanitizeCSVCell(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "formula", in: "=HYPERLINK(\"https://evil\")", want: "'=HYPERLINK(\"https://evil\")"},
		{name: "plus", in: "+1", want: "'+1"},
		{name: "minus", in: "-1", want: "'-1"},
		{name: "at", in: "@cmd", want: "'@cmd"},
		{name: "tab", in: "\tvalue", want: "'\tvalue"},
		{name: "safe", in: "Alice", want: "Alice"},
		{name: "empty", in: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeCSVCell(tt.in); got != tt.want {
				t.Fatalf("sanitizeCSVCell(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
