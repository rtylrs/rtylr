package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestFieldConstructorsRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		f    Field
		key  string
	}{
		{"any", Any("k", 1), "k"},
		{"bool", Bool("b", true), "b"},
		{"float64", Float64("f", 3.14), "f"},
		{"int", Int("i", 42), "i"},
		{"string", String("s", "hi"), "s"},
		{"error", Error(errors.New("boom")), "error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.f.Key != tc.key {
				t.Errorf("expected key %q, got %q", tc.key, tc.f.Key)
			}
		})
	}
}

func TestCtxAndWithLogger(t *testing.T) {
	var buf bytes.Buffer
	custom := slog.New(slog.NewTextHandler(&buf, nil))
	ctx := WithLogger(context.Background(), custom)
	l := Ctx(ctx)
	l.Info("hello", String("k", "v"))

	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Errorf("expected msg in output, got %q", out)
	}
	if !strings.Contains(out, "k=v") {
		t.Errorf("expected attribute in output, got %q", out)
	}
}

func TestCtxFallsBackToGlobalWhenContextEmpty(t *testing.T) {
	l := Ctx(context.Background())
	if l == nil || l.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestGlobalReturnsLogger(t *testing.T) {
	l := Global()
	if l == nil || l.Logger == nil {
		t.Fatal("expected non-nil global logger")
	}
}

func TestPretty(t *testing.T) {
	got := Pretty(map[string]int{"a": 1, "b": 2})
	if !strings.Contains(got, `"a": 1`) {
		t.Errorf("expected 'a: 1' in pretty output: %s", got)
	}

	// channel cannot be marshalled — should fall back to a friendly message.
	bad := Pretty(make(chan int))
	if !strings.Contains(bad, "error") {
		t.Errorf("expected error message for unmarshallable input, got %q", bad)
	}
}

func TestConvertFieldsHandlesAllTypes(t *testing.T) {
	fields := []Field{
		String("s", "x"),
		Int("i", 7),
		Float64("f", 1.5),
		Bool("b", true),
		Any("any", 1),
		Error(errors.New("oops")),
	}
	attrs := convertFields(fields)
	if len(attrs) != len(fields) {
		t.Fatalf("expected %d attrs, got %d", len(fields), len(attrs))
	}
}

func TestFlushIsNoOp(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Flush panicked: %v", r)
		}
	}()
	Flush()
}

func TestBetterStackEndpoint(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty uses library default", input: "", want: ""},
		{name: "host gets https scheme", input: "s123.eu-nbg-2.betterstackdata.com", want: "https://s123.eu-nbg-2.betterstackdata.com"},
		{name: "full endpoint is preserved", input: " https://logs.example.test/ ", want: "https://logs.example.test/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := betterStackEndpoint(tt.input); got != tt.want {
				t.Fatalf("betterStackEndpoint(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
