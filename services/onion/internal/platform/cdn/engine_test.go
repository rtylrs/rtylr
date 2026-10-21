package cdn

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// stubEngine writes a shell script that emulates either an old binary
// (exits after usage) or a minimal --serve responder.
func stubEngine(t *testing.T, mode string) string {
	t.Helper()
	// The stub forks dd, od and tr for every frame, so it answers the serve
	// probe orders of magnitude slower than the real binary and overruns the
	// production window whenever the machine is busy with the rest of the
	// suite. Widen the probe for stub-backed tests only; the deadlines each
	// test sets on its own jobs are untouched, so what they assert is unchanged.
	previousProbeTimeout := engineProbeTimeout
	engineProbeTimeout = 30 * time.Second
	t.Cleanup(func() { engineProbeTimeout = previousProbeTimeout })

	dir := t.TempDir()
	path := filepath.Join(dir, "media-engine")
	script := "#!/bin/sh\n"
	switch mode {
	case "oldschool":
		// Any invocation: print usage, exit 1 — like a pre-serve binary.
		script += `echo "Usage: media-engine <subcommand>"
exit 1
`
	case "serve":
		// Reads u32le-framed jobs, replies with a framed response. Uses dd to
		// read exactly N bytes and printf for framing; good enough for tests.
		script += `if [ "$1" != "--serve" ]; then echo "usage"; exit 1; fi
while :; do
  len=$(dd bs=1 count=4 2>/dev/null | od -An -tu4 -N4 | tr -cd '0-9')
  [ -z "$len" ] && exit 0
  [ "$len" = "0" ] && exit 0
  dd bs=1 count="$len" >/dev/null 2>&1
  printf '\000\000\000\000\002\000\000\000ok'
done
`
	case "slow":
		// Answers the first job (the pool's probe) then never answers again:
		// exercises ctx abort + kill.
		script += `if [ "$1" != "--serve" ]; then echo "usage"; exit 1; fi
len=$(dd bs=1 count=4 2>/dev/null | od -An -tu4 -N4 | tr -cd '0-9')
dd bs=1 count="$len" >/dev/null 2>&1
printf '\000\000\000\000\002\000\000\000ok'
while :; do
  len=$(dd bs=1 count=4 2>/dev/null | od -An -tu4 -N4 | tr -cd '0-9')
  [ -z "$len" ] && exit 0
  dd bs=1 count="$len" >/dev/null 2>&1
  sleep 300
done
`
	default:
		t.Fatalf("unknown mode %q", mode)
	}
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEnginePoolFallsBackWhenServeUnsupported(t *testing.T) {
	p := newEnginePool(stubEngine(t, "oldschool"), 2)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, _, err := p.acquire(ctx); !errors.Is(err, errEngineDisabled) {
		t.Fatalf("expected errEngineDisabled, got %v", err)
	}
	// Disabled state must latch: second acquire short-circuits too.
	if _, _, err := p.acquire(ctx); !errors.Is(err, errEngineDisabled) {
		t.Fatalf("latching failed, got %v", err)
	}
}

func TestEnginePoolServesAndRecycles(t *testing.T) {
	p := newEnginePool(stubEngine(t, "serve"), 1)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		conn, release, err := p.acquire(ctx)
		if err != nil {
			t.Fatalf("acquire %d: %v", i, err)
		}
		out, err := conn.run(ctx, []string{"transform", "--input", "x"})
		release()
		if err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		if out != "ok" {
			t.Fatalf("run %d: got %q", i, out)
		}
	}
	if len(p.idle) != 1 {
		t.Fatalf("want 1 pooled conn, got %d", len(p.idle))
	}
}

func TestEngineFrameEncoding(t *testing.T) {
	tokens := []string{"transform", "--input", "/a b/x.webp"}
	var payload []byte
	for i, tok := range tokens {
		if i > 0 {
			payload = append(payload, 0)
		}
		payload = append(payload, tok...)
	}
	frame := make([]byte, 4+len(payload))
	binary.LittleEndian.PutUint32(frame, uint32(len(payload)))
	copy(frame[4:], payload)

	if got := binary.LittleEndian.Uint32(frame); uint32(len(payload)) != got {
		t.Fatalf("length prefix mismatch")
	}
}

func TestRunTransformedSingleShotFallback(t *testing.T) {
	// A real single-shot-capable engine: /bin/true succeeds as any command.
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true executable is unavailable")
	}
	c := &CDN{mediaEngine: truePath}
	c.engine = newEnginePool(truePath, 1)
	if err := c.runTransformed(context.Background(), []string{"anything"}); err != nil {
		t.Fatalf("single-shot fallback failed: %v", err)
	}
}

func TestRunTransformedResidentPath(t *testing.T) {
	c := &CDN{mediaEngine: stubEngine(t, "serve")}
	c.engine = newEnginePool(c.mediaEngine, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.runTransformed(ctx, []string{"transform"}); err != nil {
		t.Fatalf("resident run failed: %v", err)
	}
}

// TestEnginePoolAgainstRealBinary exercises the pool against the actual C++
// media-engine when MEDIA_ENGINE_BIN points at a --serve-capable build:
//
//	MEDIA_ENGINE_BIN=build/media-engine go test ./internal/platform/cdn/ -run Real -v
func TestEnginePoolAgainstRealBinary(t *testing.T) {
	bin := os.Getenv("MEDIA_ENGINE_BIN")
	if bin == "" {
		t.Skip("set MEDIA_ENGINE_BIN to a built media-engine binary")
	}
	img := os.Getenv("MEDIA_ENGINE_TEST_IMAGE")
	if img == "" {
		t.Skip("set MEDIA_ENGINE_TEST_IMAGE to a sample input file")
	}
	out := filepath.Join(t.TempDir(), "out.webp")

	p := newEnginePool(bin, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	args := []string{
		"transform", "--input", img, "--output", out,
		"--chain", "w_64,h_64,c_fill/e_grayscale",
	}

	// First call pays spawn+probe; subsequent calls reuse the resident.
	start := time.Now()
	if err := runVia(ctx, t, p, args); err != nil {
		t.Fatalf("first job: %v", err)
	}
	first := time.Since(start)

	start = time.Now()
	n := 10
	for i := 0; i < n; i++ {
		if err := runVia(ctx, t, p, args); err != nil {
			t.Fatalf("job %d: %v", i, err)
		}
	}
	t.Logf("first job %v, steady-state %v/job", first, time.Since(start)/time.Duration(n))

	if len(p.idle) != 1 {
		t.Fatalf("want exactly 1 pooled conn, got %d", len(p.idle))
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output missing: %v", err)
	}
}

func runVia(ctx context.Context, t *testing.T, p *enginePool, args []string) error {
	t.Helper()
	conn, release, err := p.acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire: %w", err)
	}
	defer release()
	_, err = conn.run(ctx, args)
	return err
}

func TestEngineConnKilledOnExpiry(t *testing.T) {
	p := newEnginePool(stubEngine(t, "slow"), 1)
	conn, release, err := p.acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if _, err := conn.run(ctx, []string{"job"}); err == nil {
		t.Fatalf("expected abort error from non-responding engine")
	}
	release()
	if len(p.idle) != 0 {
		t.Fatalf("aborted conn must not be pooled")
	}
}
