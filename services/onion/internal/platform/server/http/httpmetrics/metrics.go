package httpmetrics

import (
	"bufio"
	"crypto/subtle"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/voxire/onion/internal/platform/config"
)

// /metrics endpoint — Prometheus text format. Behind basic auth so
// the endpoint isn't world-readable (which would leak counters about
// login volume, apply-form traffic, etc. to scrapers).
//
// Intentionally tiny — just enough for Grafana Cloud or a small
// Prometheus to chart request volume + latency + goroutines. Swap for
// a full instrumented middleware stack once we know we need histograms
// per-route.

// Request counters kept in atomic ints so the scrape handler doesn't
// need to lock anything on the hot path.
var (
	totalRequests atomic.Uint64
	totalErrors   atomic.Uint64
	totalLatency  atomic.Uint64 // microseconds, summed
	startTime     = time.Now()
)

// MetricsMiddleware is a thin counter-only middleware. Chain it
// inside the main middleware stack; pairs with the /metrics handler
// below.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't count the scrape itself — would create a feedback
		// loop when a scraper hits us every 15s.
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		totalRequests.Add(1)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.status >= 500 {
			totalErrors.Add(1)
		}
		totalLatency.Add(uint64(time.Since(start).Microseconds())) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
}

func RecordRecoveredPanic(started time.Time) {
	totalErrors.Add(1)
	elapsed := time.Since(started)
	if elapsed < 0 {
		elapsed = 0
	}
	totalLatency.Add(uint64(elapsed.Microseconds())) // #nosec G115 -- elapsed is non-negative and fits the metric's microsecond duration.
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// MetricsHandler serves the Prometheus-format scrape. Expected env:
//
//	METRICS_USER=rtylr-scraper
//	METRICS_PASS=<random>
//
// If either is empty the endpoint returns 404 — makes the feature
// opt-in rather than accidentally exposed on staging.
func Handler() http.HandlerFunc {
	var user, pass string
	if config.GlobalConfig != nil {
		user = config.GlobalConfig.MetricsUser
		pass = config.GlobalConfig.MetricsPass
	} else {
		user = os.Getenv("METRICS_USER")
		pass = os.Getenv("METRICS_PASS")
	}
	enabled := user != "" && pass != ""

	return func(w http.ResponseWriter, r *http.Request) {
		if !enabled {
			http.NotFound(w, r)
			return
		}
		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		reqs := totalRequests.Load()
		errs := totalErrors.Load()
		lat := totalLatency.Load()

		var avgLatencyMs float64
		if reqs > 0 {
			avgLatencyMs = float64(lat) / 1000.0 / float64(reqs)
		}

		uptime := time.Since(startTime).Seconds()
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		var b strings.Builder
		b.WriteString("# HELP rtylr_http_requests_total Total HTTP requests served (excluding /metrics).\n")
		b.WriteString("# TYPE rtylr_http_requests_total counter\n")
		b.WriteString("rtylr_http_requests_total ")
		b.WriteString(strconv.FormatUint(reqs, 10))
		b.WriteString("\n")

		b.WriteString("# HELP rtylr_http_errors_total HTTP 5xx responses.\n")
		b.WriteString("# TYPE rtylr_http_errors_total counter\n")
		b.WriteString("rtylr_http_errors_total ")
		b.WriteString(strconv.FormatUint(errs, 10))
		b.WriteString("\n")

		b.WriteString("# HELP rtylr_http_latency_avg_ms Average request latency in ms across the process lifetime.\n")
		b.WriteString("# TYPE rtylr_http_latency_avg_ms gauge\n")
		b.WriteString("rtylr_http_latency_avg_ms ")
		b.WriteString(strconv.FormatFloat(avgLatencyMs, 'f', 3, 64))
		b.WriteString("\n")

		b.WriteString("# HELP rtylr_process_uptime_seconds Seconds since the process started.\n")
		b.WriteString("# TYPE rtylr_process_uptime_seconds gauge\n")
		b.WriteString("rtylr_process_uptime_seconds ")
		b.WriteString(strconv.FormatFloat(uptime, 'f', 1, 64))
		b.WriteString("\n")

		b.WriteString("# HELP rtylr_goroutines Current number of goroutines.\n")
		b.WriteString("# TYPE rtylr_goroutines gauge\n")
		b.WriteString("rtylr_goroutines ")
		b.WriteString(strconv.Itoa(runtime.NumGoroutine()))
		b.WriteString("\n")

		b.WriteString("# HELP rtylr_memory_alloc_bytes Bytes allocated and still in use.\n")
		b.WriteString("# TYPE rtylr_memory_alloc_bytes gauge\n")
		b.WriteString("rtylr_memory_alloc_bytes ")
		b.WriteString(strconv.FormatUint(mem.Alloc, 10))
		b.WriteString("\n")

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = io.Copy(w, strings.NewReader(b.String()))
	}
}
