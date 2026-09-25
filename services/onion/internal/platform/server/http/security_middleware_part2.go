package http

import (
	"net/http"
	"strings"
	"time"
)

var suspiciousHeaders = []string{
	"x-forwarded-method",
	"x-forwarded-path",
	"x-forwarded-query",
	"x-forwarded-fragment",
	"x-forwarded-prefix",
	"x-forwarded-prefix-path",
	"x-forwarded-prefix-query",
	"x-forwarded-prefix-fragment",
}

var suspiciousParams = []string{
	"cmd",
	"exec",
	"system",
	"shell",
	"bash",
	"sh",
	"python",
	"perl",
	"ruby",
	"php",
	"asp",
	"jsp",
	"sql",
	"union",
	"select",
	"insert",
	"update",
	"delete",
	"drop",
	"create",
	"alter",
	"script",
	"javascript",
	"vbscript",
	"onload",
	"onerror",
	"onclick",
	"onmouseover",
	"onfocus",
	"onblur",
	"onchange",
	"onsubmit",
	"onreset",
	"onkeydown",
	"onkeyup",
	"onkeypress",
	"onmousedown",
	"onmouseup",
	"onmousemove",
	"onmouseout",
	"onmouseenter",
	"onmouseleave",
	"oncontextmenu",
	"onabort",
	"onbeforeunload",
	"onerror",
	"onhashchange",
	"onmessage",
	"onoffline",
	"ononline",
	"onpagehide",
	"onpageshow",
	"onpopstate",
	"onresize",
	"onstorage",
	"onunload",
}

var maliciousPathAllowedPrefixes = []string{
	"/console",
	"/upload",
	"/files",
	"/images",
	"/uploads",
	// Registered anonymous route (GET /.well-known/rtylr.json). The scanner
	// probe pattern for /.well-known/ must not take it down.
	"/.well-known",
}

func SecurityMiddleware() MiddlewareFunc {
	batcher := getBatcher()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientKey(r)
			userAgent := r.UserAgent()
			path := r.URL.Path
			method := r.Method
			headers := sanitizeHeaders(r.Header)
			queryParams := sanitizeQueryParams(r.URL.Query())

			if isMaliciousPath(path) {
				batcher.addRequest(blockedRequest{
					IP:          ip,
					UserAgent:   userAgent,
					Path:        path,
					Method:      method,
					Headers:     headers,
					QueryParams: queryParams,
					Reason:      "malicious_path",
					Timestamp:   time.Now(),
				})
				http.Error(w, "Not Found", http.StatusNotFound)

				return
			}

			if hasSuspiciousHeaders(r) && !isLegitimateTraffic(r) {
				batcher.addRequest(blockedRequest{
					IP:          ip,
					UserAgent:   userAgent,
					Path:        path,
					Method:      method,
					Headers:     headers,
					QueryParams: queryParams,
					Reason:      "suspicious_headers",
					Timestamp:   time.Now(),
				})
				http.Error(w, "Bad Request", http.StatusBadRequest)

				return
			}

			// CDN/media routes advertise transformation params (sharpen,
			// threshold, aspect-ratio, ...) whose names contain short tokens
			// from the suspicious list ("sh", "asp"), so the same allow-list
			// used for malicious paths must exempt their query params from the
			// probe heuristic.
			if !isAllowListedPath(path) && hasSuspiciousParams(r) {
				batcher.addRequest(blockedRequest{
					IP:          ip,
					UserAgent:   userAgent,
					Path:        path,
					Method:      method,
					Headers:     headers,
					QueryParams: queryParams,
					Reason:      "suspicious_params",
					Timestamp:   time.Now(),
				})
				http.Error(w, "Bad Request", http.StatusBadRequest)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isMaliciousPath(path string) bool {
	pathLower := strings.ToLower(path)
	if isAllowListedPath(pathLower) {
		return false
	}

	// Split into segments so that /hr/setup/status does not match the /setup pattern.
	// A pattern like "/setup" should only block a request whose path IS "/setup" or
	// starts with "/setup/" — not one where "setup" appears as a sub-segment of a
	// legitimate API path such as /hr/setup/status.
	segments := strings.Split(pathLower, "/")
	segmentSet := make(map[string]struct{}, len(segments))
	for _, s := range segments {
		if s != "" {
			segmentSet[s] = struct{}{}
		}
	}

	for _, pattern := range maliciousPatterns {
		// Patterns may carry a leading and/or trailing slash ("/.git/"); trim
		// them before matching so entries like "/.git/" are not dead: a path
		// segment can never contain '/' .
		patternLower := strings.ToLower(strings.Trim(pattern, "/"))
		if patternLower == "" {
			continue
		}
		if _, ok := segmentSet[patternLower]; ok {
			// Confirm the pattern actually appears as a standalone first segment
			// (i.e. the path starts with /<pattern> exactly), not just anywhere.
			if strings.HasPrefix(pathLower, "/"+patternLower+"/") || pathLower == "/"+patternLower {
				return true
			}
		}
	}

	return false
}

// isAllowListedPath reports whether the path belongs to a route group that must
// stay reachable even though its name collides with a scanner-probe pattern
// (CDN/upload/files/console routes). It is shared by isMaliciousPath and the
// suspicious-param check so both keep the same exemptions.
func isAllowListedPath(path string) bool {
	pathLower := strings.ToLower(path)
	for _, prefix := range maliciousPathAllowedPrefixes {
		prefixLower := strings.ToLower(prefix)
		if pathLower == prefixLower || strings.HasPrefix(pathLower, prefixLower+"/") {
			return true
		}
	}

	return false
}

func hasSuspiciousHeaders(r *http.Request) bool {
	for _, header := range suspiciousHeaders {
		if r.Header.Get(header) != "" {
			return true
		}
	}

	return false
}

func isLegitimateTraffic(r *http.Request) bool {
	userAgent := strings.ToLower(r.UserAgent())
	requestPath := strings.ToLower(r.URL.Path)

	legitimateAgents := []string{
		"mozilla", "chrome", "safari", "firefox", "edge", "opera",
		"webkit", "khtml", "gecko", "trident", "msie",
		"elb-healthchecker", "amazon", // AWS ALB health checks
	}

	legitimatePaths := []string{
		"/", "/favicon.ico", "/robots.txt", "/sitemap.xml",
		"/health", "/status", "/ping", "/metrics",
	}

	isLegitimateAgent := false

	for _, agent := range legitimateAgents {
		if strings.Contains(userAgent, agent) {
			isLegitimateAgent = true

			break
		}
	}

	isLegitimatePath := false

	for _, path := range legitimatePaths {
		// Exact match only. A substring match would let every URL through
		// because the list contains "/" — strings.Contains(anyPath, "/") is
		// always true.
		if requestPath == path {
			isLegitimatePath = true

			break
		}
	}

	return isLegitimateAgent && isLegitimatePath
}

func hasSuspiciousParams(r *http.Request) bool {
	for param := range r.URL.Query() {
		paramLower := strings.ToLower(param)
		for _, suspicious := range suspiciousParams {
			if strings.Contains(paramLower, suspicious) {
				return true
			}
		}
	}

	return false
}
