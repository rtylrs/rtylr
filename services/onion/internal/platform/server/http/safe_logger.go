package http

import (
	"log"
	stdhttp "net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
)

type sanitizedLogFormatter struct {
	base middleware.DefaultLogFormatter
}

func safeRequestLogger() func(stdhttp.Handler) stdhttp.Handler {
	return middleware.RequestLogger(&sanitizedLogFormatter{
		base: middleware.DefaultLogFormatter{
			Logger:  log.New(os.Stdout, "", log.LstdFlags),
			NoColor: true,
		},
	})
}

func accessLogsEnabled() bool {
	return envFlag("HTTP_ACCESS_LOGS") || envFlag("ACCESS_LOGS")
}

func (f *sanitizedLogFormatter) NewLogEntry(r *stdhttp.Request) middleware.LogEntry {
	return f.base.NewLogEntry(sanitizeRequestForLogging(r))
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
