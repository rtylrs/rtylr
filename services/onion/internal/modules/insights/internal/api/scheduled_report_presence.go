package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type scheduledReportFieldPresence struct {
	tracked     bool
	isActiveSet bool
}

type scheduledReportPresenceKey struct{}

func ScheduledReportPresenceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isScheduledReportWrite(r.Method, r.URL.Path) || r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<20))
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		presence := scheduledReportFieldPresence{tracked: true}
		var fields struct {
			IsActive *bool `json:"is_active"`
		}
		if len(bytes.TrimSpace(body)) > 0 && json.Unmarshal(body, &fields) == nil {
			presence.isActiveSet = fields.IsActive != nil
		}
		ctx := context.WithValue(r.Context(), scheduledReportPresenceKey{}, presence)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isScheduledReportWrite(method, path string) bool {
	if method == http.MethodPost && path == "/insights/report/v1/scheduled" {
		return true
	}
	return method == http.MethodPut && strings.HasPrefix(path, "/insights/report/v1/scheduled/")
}

func scheduledReportIsActive(ctx context.Context, value bool) *bool {
	if presence, ok := ctx.Value(scheduledReportPresenceKey{}).(scheduledReportFieldPresence); ok && presence.tracked && !presence.isActiveSet {
		return nil
	}
	return &value
}
