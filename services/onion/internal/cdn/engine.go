package cdn

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// engineMaxJobs recycles a resident media-engine after this many jobs so
// leaks in the child cannot accumulate forever.
const engineMaxJobs = 500

// engineJobTimeout bounds one framed request/response exchange with the
// resident engine; on expiry the process is killed and the job falls back
// to a single-shot invocation.
const engineJobTimeout = 55 * time.Second

// engineProbeTimeout bounds the initial --serve capability probe. It has to
// tolerate fork/exec + interpreter startup on a loaded host (the test suite
// runs every package in parallel), while staying well under the job timeout.
//
// A var rather than a const so a test driving a deliberately slow stub engine
// can widen it; production never reassigns it.
var engineProbeTimeout = 15 * time.Second

// maxEngineResponse caps accepted response bodies so garbage on the wire
// (e.g. usage text from a pre-serve binary) cannot be mistaken for framing.
const maxEngineResponse = 4 << 20

// engineConn wraps one resident `media-engine --serve` process speaking the
// length-prefixed framing protocol: requests are u32le length + NUL-separated
// argv tokens, responses are u32le exit code + u32le msgLen + msg bytes.
type engineConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	mu    sync.Mutex
	jobs  int
	alive bool
}

func (c *engineConn) isAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.alive
}

// startEngine spawns a resident engine. It fails (nil, err) when the binary
// predates --serve support or cannot start at all.
func startEngine(ctx context.Context, path string) (*engineConn, error) {
	// #nosec G204 G702 -- configured server-side path only; the path is validated at startup.
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd := exec.Command(path, "--serve")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("engine stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("engine stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting engine: %w", err)
	}
	c := &engineConn{cmd: cmd, stdin: stdin, stdout: stdout, alive: true}

	// Any structured framed exchange proves --serve exists. Pre-serve binaries
	// print usage and exit, surfacing as a protocol violation below.
	probeCtx, cancel := context.WithTimeout(ctx, engineProbeTimeout)
	defer cancel()
	if err := c.probe(probeCtx); err != nil {
		c.kill()
		return nil, fmt.Errorf("engine serve probe: %w", err)
	}
	return c, nil
}

// probe verifies the child speaks the framing protocol without requiring any
// particular subcommand or input file: an unknown subcommand still produces a
// well-formed (nonzero-status) response.
func (c *engineConn) probe(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.alive {
		return fmt.Errorf("engine connection dead")
	}
	type res struct{ err error }
	done := make(chan res, 1)
	go func() {
		_, _, err := c.exchange(ctx, framedJob([]string{"help"}))
		done <- res{err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			cmd := c.cmd
			go func() { _ = cmd.Wait() }()
			c.alive = false
			return r.err
		}
		c.jobs++
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// framedJob encodes argv-style tokens into a request frame.
func framedJob(tokens []string) []byte {
	payload := make([]byte, 0, 64)
	for i, t := range tokens {
		if i > 0 {
			payload = append(payload, 0)
		}
		payload = append(payload, t...)
	}
	frame := make([]byte, 4+len(payload))
	binary.LittleEndian.PutUint32(frame, uint32(len(payload))) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	copy(frame[4:], payload)
	return frame
}

// exchange performs one synchronous framed round trip bounded by ctx and
// engineJobTimeout. It must be called with c.mu held.
func (c *engineConn) exchange(ctx context.Context, frame []byte) (code uint32, msg string, err error) {
	type result struct {
		code uint32
		msg  string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		var head [8]byte
		if _, werr := c.stdin.Write(frame); werr != nil {
			done <- result{0, "", fmt.Errorf("engine write: %w", werr)}
			return
		}
		if _, rerr := io.ReadFull(c.stdout, head[:]); rerr != nil {
			done <- result{0, "", fmt.Errorf("engine read header: %w", rerr)}
			return
		}
		msgLen := binary.LittleEndian.Uint32(head[4:])
		if msgLen > maxEngineResponse {
			done <- result{0, "", fmt.Errorf("engine response too large (%d bytes)", msgLen)}
			return
		}
		msg := make([]byte, msgLen)
		if _, rerr := io.ReadFull(c.stdout, msg); rerr != nil {
			done <- result{0, "", fmt.Errorf("engine read body: %w", rerr)}
			return
		}
		done <- result{binary.LittleEndian.Uint32(head[0:4]), string(msg), nil}
	}()

	select {
	case r := <-done:
		return r.code, r.msg, r.err
	case <-ctx.Done():
		return 0, "", fmt.Errorf("engine job aborted: %w", ctx.Err())
	case <-time.After(engineJobTimeout):
		return 0, "", fmt.Errorf("engine job timed out")
	}
}

// run sends one argv-style job and returns captured stdout. Protocol
// violations or timeouts kill the connection; callers should fall back to
// single-shot execution.
func (c *engineConn) run(ctx context.Context, tokens []string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.alive {
		return "", fmt.Errorf("engine connection dead")
	}

	code, msg, err := c.exchange(ctx, framedJob(tokens))
	if err != nil {
		// A protocol error or ctx timeout usually means the child is hung,
		// not dead — SIGKILL it so the pipes and process actually go away
		// instead of waiting on finalizers.
		c.killLocked()
		return "", err
	}
	c.jobs++
	if code != 0 {
		return string(msg), fmt.Errorf("media-engine exited %d: %s", code, msg)
	}
	return msg, nil
}

// kill terminates the resident process, ignoring shutdown errors.
func (c *engineConn) kill() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.killLocked()
}

// killLocked terminates the resident process; c.mu must be held.
func (c *engineConn) killLocked() {
	c.alive = false
	cmd := c.cmd
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		go func() { _ = cmd.Wait() }()
		c.cmd = nil
	}
}

// enginePool manages up to maxConns resident engines with single-shot exec as
// the fallback whenever resident mode is unavailable or misbehaves.
type enginePool struct {
	path     string
	maxConns int

	mu       sync.Mutex
	idle     []*engineConn
	disabled bool // set once the binary proves it lacks working --serve
	inFlight int
}

func newEnginePool(path string, maxConns int) *enginePool {
	if maxConns < 1 {
		maxConns = 1
	}
	return &enginePool{path: path, maxConns: maxConns}
}

var (
	errEngineDisabled = fmt.Errorf("resident engine unavailable")
	errEngineBusy     = fmt.Errorf("resident engine at capacity")
)

// acquire returns a usable conn, spawning/recycling as needed.
func (p *enginePool) acquire(ctx context.Context) (*engineConn, func(), error) {
	p.mu.Lock()
	if p.disabled {
		p.mu.Unlock()
		return nil, nil, errEngineDisabled
	}
	if len(p.idle) > 0 {
		c := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]
		p.mu.Unlock()
		if !c.isAlive() {
			return p.acquire(ctx)
		}
		return c, func() { p.release(c) }, nil
	}
	if p.inFlight >= p.maxConns {
		p.mu.Unlock()
		return nil, nil, errEngineBusy
	}
	p.inFlight++
	p.mu.Unlock()

	c, err := startEngine(ctx, p.path)
	if err != nil {
		p.mu.Lock()
		p.inFlight--
		if ctx.Err() == nil {
			// Old binaries exit during the probe; latch disabled rather than
			// pay a failed spawn per request. Single-shot keeps working.
			p.disabled = true
		}
		p.mu.Unlock()
		if ctx.Err() != nil {
			// The request went away mid-probe; nothing was learned about the
			// binary, so leave the pool enabled for the next request.
			return nil, nil, fmt.Errorf("%w: %v", errEngineDisabled, context.Canceled)
		}
		return nil, nil, fmt.Errorf("%w: %v", errEngineDisabled, err)
	}
	return c, func() { p.release(c) }, nil
}

func (p *enginePool) release(c *engineConn) {
	keep := c.isAlive() && c.jobs < engineMaxJobs
	if !keep {
		c.kill()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.inFlight--
	if keep {
		p.idle = append(p.idle, c)
	}
}

// runTransformed executes a transformation via the pool, transparently falling
// back to a single-shot invocation when resident mode is unavailable, busy,
// or hits a protocol error.
func (c *CDN) runTransformed(ctx context.Context, args []string) error {
	conn, release, err := c.engine.acquire(ctx)
	switch {
	case err == nil:
		_, runErr := conn.run(ctx, args)
		release()
		if runErr == nil {
			return nil
		}
	case err == errEngineBusy || ctx.Err() == nil && err != context.Canceled:
		// Busy or transient spawn issue: fall through to single-shot rather
		// than fail the request. Genuine cancellation still aborts.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	default:
		return err
	}

	execCtx, cancel := context.WithTimeout(ctx, engineTimeout)
	defer cancel()
	// #nosec G204 G702 -- mediaEngine is a server-side configured binary path; CLI args are built from validated params and temp paths.
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd := exec.CommandContext(execCtx, c.mediaEngine, args...)
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("media-engine failed: %w\noutput: %s", runErr, out)
	}
	return nil
}
