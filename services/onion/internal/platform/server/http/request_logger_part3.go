package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/observability"
)

func logRequest(ctx context.Context, r *http.Request, statusCode int, duration time.Duration, headers, query map[string][]string, serviceName string, count int) {
	routeParams := routeParamNamesFromRequest(r)

	queryParams := make([]string, 0, len(query))
	for key, values := range query {
		queryParams = append(queryParams, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
	}

	headerParams := make([]string, 0, len(headers))
	for key, values := range headers {
		headerParams = append(headerParams, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
	}

	var statusIcon string
	if statusCode >= 400 && statusCode < 500 {
		statusIcon = "❌"
	} else if statusCode >= 500 {
		statusIcon = "💥"
	} else {
		statusIcon = "✅"
	}

	var countInfo string
	if count > 1 {
		countInfo = fmt.Sprintf(" (occurred %d times)", count)
	}

	message := fmt.Sprintf("%s HTTP request%s", statusIcon, countInfo)
	fields := []logger.Field{
		logger.String("service", serviceName),
		logger.String("method", r.Method),
		logger.String("url", requestRouteTemplate(r)),
		logger.Int("status", statusCode),
		logger.String("duration", duration.String()),
		logger.String("user_agent", truncateString(r.UserAgent(), 100)),
		logger.String("remote_ip", getRemoteIP(r)),
		logger.String("route_params", formatList(routeParams, "• ")),
		logger.String("query_params", formatList(queryParams, "• ")),
		logger.String("headers", formatList(headerParams, "• ")),
		logger.String("timestamp", time.Now().Format("2006-01-02 15:04:05 UTC")),
	}

	if statusCode >= 500 {
		logger.Ctx(ctx).Error(message, fields...)

		return
	}

	logger.Ctx(ctx).Warn(message, fields...)
}

func requestRouteTemplate(r *http.Request) string {
	return observability.RequestRouteTemplate(r)
}

func routeContextForRequest(r *http.Request) *chi.Context {
	if r == nil {
		return nil
	}

	routeCtx := chi.RouteContext(r.Context())
	if routeCtx != nil && routeCtx.RoutePattern() != "" {
		return routeCtx
	}
	if r.URL == nil {
		return routeCtx
	}

	router, ok := r.Context().Value(observability.RouteLookupContextKey{}).(*chi.Mux)
	if ok && router != nil {
		resolved := chi.NewRouteContext()
		if router.Find(resolved, r.Method, r.URL.Path) != "" {
			return resolved
		}
	}

	return routeCtx
}

func routeParamNamesFromRequest(r *http.Request) []string {
	if r == nil {
		return nil
	}

	routeCtx := routeContextForRequest(r)
	if routeCtx == nil || len(routeCtx.URLParams.Keys) == 0 {
		return observability.RouteParamNames(requestRouteTemplate(r))
	}

	routeParams := make([]string, 0, len(routeCtx.URLParams.Keys))
	for _, key := range routeCtx.URLParams.Keys {
		routeParams = append(routeParams, key)
	}

	return routeParams
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

func getRemoteIP(r *http.Request) string {
	return getClientKey(r)
}

func formatList(items []string, prefix string) string {
	if len(items) == 0 {
		return "None"
	}

	var result strings.Builder
	for _, item := range items {
		result.WriteString(prefix)
		result.WriteString(item)
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String())
}
