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
	"token":         {},
	"access_token":  {},
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

func (f *sanitizedLogFormatter) NewLogEntry(r *stdhttp.Request) middleware.LogEntry {
	return f.base.NewLogEntry(sanitizeRequestForLogging(r))
}

func sanitizeRequestForLogging(r *stdhttp.Request) *stdhttp.Request {
	if r == nil || r.URL == nil || r.URL.RawQuery == "" {
		return r
	}

	clone := new(stdhttp.Request)
	*clone = *r
	urlCopy := *r.URL
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
