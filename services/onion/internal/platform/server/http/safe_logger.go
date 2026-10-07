package http

import (
	"log"
	stdhttp "net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

var sensitiveQueryKeys = map[string]struct{}{
	"api_key":       {},
	"code":          {},
	"key":           {},
	"profile_token": {},
	"refresh_token": {},
	"secret":        {},
	"signature":     {},
	"sig":           {},
	"ticket":        {},
	"token":         {},
	"access_token":  {},
}

// The request/error logger uses the shared substring redactor in
// request_logger.go. Extend that list here so the two logging paths cannot
// drift again when a credential is added to the access-log map.
//
//nolint:gochecknoinits // The shared package-level redactor is initialized before handlers run.
func init() {
	sensitiveFieldParts = append(sensitiveFieldParts, "ticket", "sig")
}

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
		if _, ok := sensitiveQueryKeys[strings.ToLower(key)]; ok {
			query.Set(key, "[REDACTED]")
		}
	}
	urlCopy.RawQuery = query.Encode()
	clone.URL = &urlCopy
	clone.RequestURI = urlCopy.RequestURI()
	return clone
}
