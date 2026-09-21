package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type recordingHandler struct {
	records []slog.Record
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *recordingHandler) Handle(_ context.Context, record slog.Record) error {
	h.records = append(h.records, record.Clone())
	return nil
}
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

func TestDedupHandlerKeepsDistinctAttributedEvents(t *testing.T) {
	inner := &recordingHandler{}
	dedup := &dedupHandler{inner: inner, window: time.Hour}

	first := slog.NewRecord(time.Now(), slog.LevelWarn, "Invalid token", 0)
	first.AddAttrs(slog.String("remote_ip", "192.0.2.1"))
	second := slog.NewRecord(time.Now(), slog.LevelWarn, "Invalid token", 0)
	second.AddAttrs(slog.String("remote_ip", "192.0.2.2"))
	repeat := slog.NewRecord(time.Now(), slog.LevelWarn, "Invalid token", 0)
	repeat.AddAttrs(slog.String("remote_ip", "192.0.2.2"))

	for _, record := range []slog.Record{first, second, repeat} {
		if err := dedup.Handle(context.Background(), record); err != nil {
			t.Fatalf("Handle: %v", err)
		}
	}
	if len(inner.records) != 2 {
		t.Fatalf("forwarded %d records, want 2", len(inner.records))
	}
}

func TestDedupHandlerCleansStaleEntries(t *testing.T) {
	inner := &recordingHandler{}
	dedup := &dedupHandler{inner: inner, window: time.Millisecond}
	first := slog.NewRecord(time.Now(), slog.LevelWarn, "first", 0)
	if err := dedup.Handle(context.Background(), first); err != nil {
		t.Fatalf("first Handle: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	second := slog.NewRecord(time.Now(), slog.LevelWarn, "second", 0)
	if err := dedup.Handle(context.Background(), second); err != nil {
		t.Fatalf("second Handle: %v", err)
	}
	if len(dedup.seen) != 1 {
		t.Fatalf("dedup map retained %d entries, want 1", len(dedup.seen))
	}
}

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

func TestFlushWithoutBetterStackHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Flush panicked: %v", r)
		}
	}()
	Flush()
}

func TestFlushWaitsForPendingBetterStackLogs(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	h := &betterStackHandler{
		level:    slog.LevelWarn,
		delivery: &betterStackDelivery{},
		send: func(context.Context, string, string, time.Duration, []byte) error {
			close(started)
			<-release
			return nil
		},
	}

	record := slog.NewRecord(time.Now(), slog.LevelWarn, "pending", 0)
	if err := h.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	<-started

	flushed := make(chan struct{})
	go func() {
		h.Flush()
		close(flushed)
	}()
	select {
	case <-flushed:
		t.Fatal("Flush returned before pending send completed")
	case <-time.After(10 * time.Millisecond):
	}

	close(release)
	select {
	case <-flushed:
	case <-time.After(time.Second):
		t.Fatal("Flush did not return after pending send completed")
	}
}

// Fatal must deliver the record it just queued to Better Stack before exiting;
// os.Exit skips the deferred Flush in runServer.
func TestFatalFlushesBetterStackBeforeExiting(t *testing.T) {
	originalHandler := activeBetterStackHandler
	originalDefault := slog.Default()
	originalExit := exitProcess
	t.Cleanup(func() {
		activeBetterStackHandler = originalHandler
		slog.SetDefault(originalDefault)
		exitProcess = originalExit
	})

	started := make(chan struct{})
	release := make(chan struct{})
	delivered := make(chan struct{})
	h := &betterStackHandler{
		level:    slog.LevelWarn,
		delivery: &betterStackDelivery{},
		send: func(context.Context, string, string, time.Duration, []byte) error {
			close(started)
			<-release
			close(delivered)
			return nil
		},
	}
	activeBetterStackHandler = h
	slog.SetDefault(slog.New(&multiHandler{handlers: []slog.Handler{h}}))

	exitCode := make(chan int, 1)
	exitProcess = func(code int) { exitCode <- code }

	go Global().Fatal("boot failed", String("reason", "missing env"))

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("fatal record was never sent to Better Stack")
	}

	select {
	case code := <-exitCode:
		t.Fatalf("Fatal exited with code %d before the pending Better Stack POST finished", code)
	case <-time.After(10 * time.Millisecond):
	}

	close(release)
	select {
	case <-delivered:
	case <-time.After(time.Second):
		t.Fatal("Better Stack POST did not finish")
	}
	select {
	case code := <-exitCode:
		if code != 1 {
			t.Fatalf("Fatal exit code = %d, want 1", code)
		}
	case <-time.After(time.Second):
		t.Fatal("Fatal did not exit after flushing")
	}
}

func TestBetterStackFlushIsReusable(t *testing.T) {
	delivered := make(chan struct{}, 2)
	h := &betterStackHandler{
		level:    slog.LevelWarn,
		delivery: &betterStackDelivery{},
		send: func(context.Context, string, string, time.Duration, []byte) error {
			delivered <- struct{}{}
			return nil
		},
	}
	record := func(msg string) slog.Record {
		return slog.NewRecord(time.Now(), slog.LevelWarn, msg, 0)
	}
	assertDelivered := func(stage string) {
		t.Helper()
		select {
		case <-delivered:
		case <-time.After(time.Second):
			t.Fatalf("%s: record was dropped because Flush disabled delivery permanently", stage)
		}
	}

	// runServer()'s deferred Flush runs on every crash-restart cycle; the next
	// cycle's warn/error records must still reach Better Stack.
	h.Flush()
	if err := h.Handle(context.Background(), record("after first flush")); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertDelivered("after first flush")

	h.Flush()
	if err := h.Handle(context.Background(), record("after second flush")); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertDelivered("after second flush")
}

func TestBetterStackHandlerPreservesGroupedAttributes(t *testing.T) {
	h := &betterStackHandler{
		level:    slog.LevelWarn,
		delivery: &betterStackDelivery{},
		send:     sendBetterStack,
	}

	grouped, ok := h.WithGroup("request").WithAttrs([]slog.Attr{slog.String("path", "/health")}).(*betterStackHandler)
	if !ok {
		t.Fatal("WithAttrs returned unexpected handler type")
	}
	if len(grouped.attrs) != 1 {
		t.Fatalf("got %d handler attrs, want one group", len(grouped.attrs))
	}
	if grouped.attrs[0].Key != "request" || grouped.attrs[0].Value.Kind() != slog.KindGroup {
		t.Fatalf("got %#v, want request group", grouped.attrs[0])
	}
	groupAttrs := grouped.attrs[0].Value.Group()
	if len(groupAttrs) != 1 || groupAttrs[0].Key != "path" {
		t.Fatalf("got grouped attrs %#v, want path attr", groupAttrs)
	}
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
