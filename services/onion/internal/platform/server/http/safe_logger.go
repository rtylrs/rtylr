package http

import (
	"bufio"
	"errors"
	"log"
	"net"
	stdhttp "net/http"
	"os"
	"time"
)

// safeRequestLogger writes one access-log line per request, with the route
// template instead of the real path and sensitive query values redacted.
func safeRequestLogger() func(stdhttp.Handler) stdhttp.Handler {
	out := log.New(os.Stdout, "", log.LstdFlags)
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			started := time.Now()
			recorder := &accessLogWriter{ResponseWriter: w}
			next.ServeHTTP(recorder, r)
			status := recorder.status
			if status == 0 {
				status = stdhttp.StatusOK
			}
			logged := sanitizeRequestForLogging(r)
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			out.Printf("%q from %s - %d %dB in %s",
				logged.Method+" "+scheme+"://"+r.Host+logged.RequestURI+" "+r.Proto,
				r.RemoteAddr, status, recorder.bytes, time.Since(started))
		})
	}
}

// accessLogWriter records the status and size while passing flushes and
// connection hijacks (SSE, WebSockets) through to the real writer.
type accessLogWriter struct {
	stdhttp.ResponseWriter
	status int
	bytes  int
}

func (w *accessLogWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *accessLogWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = stdhttp.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func (w *accessLogWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(stdhttp.Flusher); ok {
		flusher.Flush()
	}
}

func (w *accessLogWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(stdhttp.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("response writer does not support hijacking")
}

func (w *accessLogWriter) Unwrap() stdhttp.ResponseWriter { return w.ResponseWriter }

func accessLogsEnabled() bool {
	return envFlag("HTTP_ACCESS_LOGS") || envFlag("ACCESS_LOGS")
}

func sanitizeRequestForLogging(r *stdhttp.Request) *stdhttp.Request {
	if r == nil || r.URL == nil {
		return r
	}

	clone := new(stdhttp.Request)
	*clone = *r
	urlCopy := *r.URL
	urlCopy.Path = requestRouteTemplate(r)
	urlCopy.RawPath = ""
	query := urlCopy.Query()
	for key := range query {
		if isSensitiveField(key) {
			query.Set(key, "[REDACTED]")
		}
	}
	urlCopy.RawQuery = query.Encode()
	clone.URL = &urlCopy
	clone.RequestURI = urlCopy.RequestURI()
	return clone
}
