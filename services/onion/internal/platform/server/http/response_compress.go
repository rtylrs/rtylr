package http

import (
	"fmt"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

// compressibleContentTypes are the response types worth compressing: text,
// JSON and the other formats the old chi compressor handled.
var compressibleContentTypes = []string{
	"text/html",
	"text/css",
	"text/plain",
	"text/javascript",
	"application/javascript",
	"application/x-javascript",
	"application/json",
	"application/atom+xml",
	"application/rss+xml",
	"image/svg+xml",
}

// responseCompressMiddleware gzip-compresses JSON/HTML/text responses when the
// client sends Accept-Encoding: gzip. Long-lived streams (SSE, WebSocket) and
// /metrics are skipped — compressing those mid-flight breaks clients.
func responseCompressMiddleware() MiddlewareFunc {
	wrap, err := gzhttp.NewWrapper(
		gzhttp.CompressionLevel(5),
		gzhttp.MinSize(0),
		gzhttp.ContentTypes(compressibleContentTypes),
	)
	if err != nil {
		panic(fmt.Errorf("configure response compression: %w", err))
	}

	return func(next http.Handler) http.Handler {
		compressed := wrap(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isLongLivedRequest(r) || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			compressed.ServeHTTP(w, r)
		})
	}
}
