package http

import (
	"net/http"
	"net/url"
	"regexp"
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

// suspiciousParamNames is an EXACT (case-insensitive) denylist of query
// parameter names that map onto a dangerous downstream sink. Names are attacker
// controlled, so the name and the value are judged separately: a name is
// blocked only on exact equality, never on substring.
//
// This was previously matched with strings.Contains against SQL/HTML keyword
// fragments, which made ordinary filters return 400 in production:
// "created_at" contains "create", "updated_at" contains "update",
// "deleted_at" contains "delete" and "description" contains "script".
//
// Bare SQL verbs (select/insert/update/delete/drop/alter/create) and the
// "on*" DOM event handlers are deliberately absent: as parameter names they
// are not a sink, and exact-matching them would re-break legitimate bulk and
// filter APIs for no security gain.
var suspiciousParamNames = map[string]struct{}{
	// Command-execution sinks.
	"cmd": {}, "command": {}, "exec": {}, "evaluate": {}, "passthru": {},
	"popen": {}, "system": {}, "shell": {}, "shell_exec": {}, "system_exec": {},

	// Raw-SQL / query-fragment passthroughs.
	"sql": {}, "sqlquery": {}, "sql_query": {}, "sqlinjection": {}, "sql_injection": {},
	"rawsql": {}, "raw_sql": {}, "query_sql": {},
	"orderbyraw": {}, "order_by_raw": {}, "groupbyraw": {}, "group_by_raw": {},
	"where_raw": {}, "having_raw": {},

	// Script / template injection sinks.
	"script": {}, "javascript": {}, "vbscript": {},
	"raw_template": {}, "template_source": {},

	// Object-protocol keys. Never legitimate query parameter names.
	"__proto__": {}, "constructor": {}, "prototype_pollution": {},
}

const (
	// maxParamValueBytes bounds the work spent on one parameter value. A longer
	// value is rejected instead of truncated: truncating would let an attacker
	// push the payload past the cut and bypass the check.
	maxParamValueBytes = 8 << 10
	// maxQueryScanBytes bounds the total value bytes inspected per request.
	// Exceeding it rejects the request rather than skipping the remainder, so
	// the budget cannot be used to hide a payload outside the scanned window.
	maxQueryScanBytes = 32 << 10
	// maxParamDecodeRounds bounds the percent-decoding chase used to defeat
	// encoded payloads, keeping the scan O(n).
	maxParamDecodeRounds = 3
)

// Go's regexp is RE2: linear time with no backtracking, so these patterns have
// bounded cost on adversarial input. Each entry is a full injection idiom — a
// bare quote or a bare keyword is far too common in real filter values
// (O'Brien, "delete from my cart") to be a signal on its own.
var sqliValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)['"]\s*(or|and)\s+['"]?\s*\d`), // ' OR 1=1--
	regexp.MustCompile(`(?i)['"]\s*(--|#|/\*)`),            // quote + comment terminator
	regexp.MustCompile(`(?i)\bunion\s+(all\s+)?select\b`),  // UNION SELECT
	regexp.MustCompile(`(?i)\b(drop|alter|truncate)\s+table\b`),
	regexp.MustCompile(`(?i);\s*(drop|delete|update|insert|select|union|truncate|alter)\b`), // stacked query
	regexp.MustCompile(`(?i)\b(sleep|benchmark|pg_sleep|waitfor\s+delay)\s*\(`),             // time-based blind
}

// sqliStatementPatterns are SQL statements whose keyword pairs also occur in
// ordinary English ("delete from my cart", "insert into the order"). They only
// count when the same value also carries a structural marker, which a payload
// needs to escape a quoted context but prose does not.
var sqliStatementPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bdelete\s+from\b`),
	regexp.MustCompile(`(?i)\binsert\s+into\b`),
	regexp.MustCompile(`(?i)\bupdate\s+\S+\s+set\b`),
}

var sqlStructureMarker = regexp.MustCompile(`['";]|--|/\*|\(`)

var traversalValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.[/\\]`),  // ../
	regexp.MustCompile(`[/\\]\.\.`),  // /..
	regexp.MustCompile(`(?i)%2e%2e`), // percent-encoded ..
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

			// No path exemption here: the denylist is exact, so the media/CDN
			// transform params (sharpen, threshold, aspect-ratio) that used to
			// need an allow-list no longer collide with it. Exempting a path
			// would instead hand an attacker a route-around (?cmd=... on
			// /images/...).
			if hasSuspiciousParams(r) {
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
// (CDN/upload/files/console routes). It only exempts isMaliciousPath: the
// parameter check is exact and must not become path-conditional, or an attacker
// picks an allow-listed path to skip it.
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

// hasSuspiciousParams reports whether the request's query string carries a
// denylisted parameter NAME or a parameter VALUE matching an injection idiom
// (SQLi, CRLF, path traversal, percent-encoded payload).
//
// Only the query string is inspected, deliberately. Nothing in this service
// parses form bodies (no ParseForm/PostForm anywhere), and this middleware runs
// before RequestBodyLimitMiddleware, so reading the body here would consume the
// stream and defeat the request-size ceiling. Body values are the handler's job.
func hasSuspiciousParams(r *http.Request) bool {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil && len(query) == 0 && strings.Contains(r.URL.RawQuery, "&") {
		// r.URL.Query() swallows this error and hands back an empty map, which
		// made the whole check a silent no-op once an attacker padded the query
		// past Go's 10,000-pair limit. A completely unparseable multi-parameter
		// query is never legitimate: the handler would receive no parameters.
		return true
	}

	for name := range query {
		if _, blocked := suspiciousParamNames[normalizeParamName(name)]; blocked {
			return true
		}
	}

	scanned := 0
	for _, values := range query {
		for _, value := range values {
			scanned += len(value)
			if len(value) > maxParamValueBytes || scanned > maxQueryScanBytes {
				return true
			}
			if isSuspiciousParamValue(value) {
				return true
			}
		}
	}

	return false
}

// normalizeParamName folds a parameter name for exact denylist comparison.
// Trimming surrounding brackets defeats the trivial "[cmd]" / "cmd[]" evasion
// used by frameworks that index params as arrays.
func normalizeParamName(name string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(name), "[]"))
}

// isSuspiciousParamValue inspects a parameter value, re-inspecting it after
// percent-decoding so double/triple encoded payloads cannot slip through.
func isSuspiciousParamValue(value string) bool {
	current := value

	for round := 0; round < maxParamDecodeRounds; round++ {
		if hasControlBytes(current) ||
			matchesAny(sqliValuePatterns, current) ||
			matchesAny(traversalValuePatterns, current) ||
			(sqlStructureMarker.MatchString(current) && matchesAny(sqliStatementPatterns, current)) {
			return true
		}

		if !strings.Contains(current, "%") {
			return false
		}

		decoded, err := url.QueryUnescape(current)
		if err != nil || decoded == current {
			return false
		}

		current = decoded
	}

	return false
}

// hasControlBytes reports CR, LF, NUL and other C0/DEL bytes: response
// splitting, header injection and truncation. Tab is allowed.
func hasControlBytes(value string) bool {
	for i := 0; i < len(value); i++ {
		if c := value[i]; (c < 0x20 && c != '\t') || c == 0x7f {
			return true
		}
	}

	return false
}

func matchesAny(patterns []*regexp.Regexp, value string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}

	return false
}
