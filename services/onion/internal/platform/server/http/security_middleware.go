package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/slack"
)

type blockedRequest struct {
	IP          string
	UserAgent   string
	Path        string
	Method      string
	Headers     map[string][]string
	QueryParams map[string][]string
	Reason      string
	Timestamp   time.Time
}

type securityBatcher struct {
	requests    map[string][]blockedRequest
	mutex       sync.RWMutex
	ticker      *time.Ticker
	done        chan bool
	serviceName string
}

var (
	batcher *securityBatcher
	once    sync.Once
)

func getBatcher() *securityBatcher {
	once.Do(func() {
		serviceName := logger.ServiceName()

		batcher = &securityBatcher{
			requests:    make(map[string][]blockedRequest),
			ticker:      time.NewTicker(30 * time.Second),
			done:        make(chan bool),
			serviceName: serviceName,
		}
		go batcher.flushRoutine()
	})

	return batcher
}

func (b *securityBatcher) addRequest(req blockedRequest) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.requests[req.IP] = append(b.requests[req.IP], req)
}

func (b *securityBatcher) flushRoutine() {
	for {
		select {
		case <-b.ticker.C:
			b.flush()
		case <-b.done:
			b.ticker.Stop()

			return
		}
	}
}

func (b *securityBatcher) flush() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for ip, requests := range b.requests {
		if len(requests) == 0 {
			continue
		}

		reasons := make(map[string]int)
		paths := make(map[string]int)
		userAgents := make(map[string]int)
		methods := make(map[string]int)
		allHeaders := make(map[string]int)
		allQueryParams := make(map[string]int)

		for _, req := range requests {
			reasons[req.Reason]++
			paths[req.Path]++
			userAgents[req.UserAgent]++
			methods[req.Method]++

			for header := range req.Headers {
				allHeaders[header]++
			}

			for param := range req.QueryParams {
				allQueryParams[param]++
			}
		}

		message := formatSecurityMessage(ip, len(requests), reasons, paths, userAgents, methods, allHeaders, allQueryParams, b.serviceName)
		logger.Global().Warn("Blocked malicious requests batch",
			logger.String("ip", ip),
			logger.Int("total_requests", len(requests)),
			logger.String("reasons", formatMap(reasons)),
			logger.String("paths", formatMap(paths)),
			logger.String("user_agents", formatMap(userAgents)),
			logger.String("summary", message),
		)
		slack.Post(slack.Suspicious, message)

		delete(b.requests, ip)
	}
}

func formatSecurityMessage(ip string, totalRequests int, reasons, paths, userAgents, methods, headers, queryParams map[string]int, serviceName string) string {
	reasonsList := formatMapToList(reasons)
	pathsList := formatMapToList(paths)
	userAgentsList := formatMapToList(userAgents)
	methodsList := formatMapToList(methods)
	headersList := formatMapToList(headers)
	queryParamsList := formatMapToList(queryParams)

	return fmt.Sprintf(`🚨 Security Alert - Blocked Malicious Requests

🏢 Service: %s
🖥️ IP Address: %s
📊 Total Requests: %d
⏰ Time Window: 30 seconds

🎯 Attack Types:
%s

🌐 Targeted Paths:
%s

🤖 User Agents:
%s

📝 HTTP Methods:
%s

📋 Headers Used:
%s

🔍 Query Parameters:
%s

📅 Timestamp: %s`,
		serviceName,
		ip,
		totalRequests,
		reasonsList,
		pathsList,
		userAgentsList,
		methodsList,
		headersList,
		queryParamsList,
		time.Now().Format("2006-01-02 15:04:05 UTC"),
	)
}

func formatMapToList(m map[string]int) string {
	if len(m) == 0 {
		return "None"
	}

	parts := make([]string, 0, len(m))
	for key, count := range m {
		parts = append(parts, fmt.Sprintf("• %s: %d", key, count))
	}

	return strings.Join(parts, "\n")
}

func formatMap(m map[string]int) string {
	parts := make([]string, 0, len(m))
	for key, count := range m {
		parts = append(parts, key+":"+strconv.Itoa(count))
	}

	return strings.Join(parts, ",")
}

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitized := make(map[string][]string)
	sensitiveHeaders := []string{"authorization", "cookie", "x-api-key", "token", "password"}

	for key, values := range headers {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveHeaders {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true

				break
			}
		}

		if isSensitive {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}

	return sanitized
}

func sanitizeQueryParams(params map[string][]string) map[string][]string {
	sanitized := make(map[string][]string)
	sensitiveParams := []string{"token", "key", "password", "secret", "auth"}

	for key, values := range params {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveParams {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true

				break
			}
		}

		if isSensitive {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}

	return sanitized
}

var maliciousPatterns = []string{
	"/.git/",
	"/.env",
	"/config.json",
	"/_all_dbs",
	"/ecp/",
	"/debug/",
	"/@vite/",
	"/server",
	"/about",
	"/phpmyadmin",
	"/admin",
	"/wp-admin",
	"/wp-login",
	"/wp-includes",
	"/wp-content",
	"/wp-config",
	"/xmlrpc.php",
	"/wlwmanifest.xml",
	"/.htaccess",
	"/.htpasswd",
	"/.well-known/",
	"/api/",
	"/graphql",
	"/swagger",
	"/openapi",
	"/actuator",
	"/info",
	"/pong",
	"/ready",
	"/live",
	"/startup",
	"/shutdown",
	"/restart",
	"/reload",
	"/settings",
	"/setup",
	"/install",
	"/upgrade",
	"/migrate",
	"/seed",
	"/backup",
	"/restore",
	"/export",
	"/import",
	"/download",
	"/upload",
	"/file",
	"/files",
	"/static",
	"/assets",
	"/images",
	"/css",
	"/js",
	"/fonts",
	"/media",
	"/uploads",
	"/public",
	"/private",
	"/secret",
	"/hidden",
	"/internal",
	"/system",
	"/root",
	"/home",
	"/tmp",
	"/var",
	"/etc",
	"/usr",
	"/bin",
	"/sbin",
	"/lib",
	"/lib64",
	"/proc",
	"/sys",
	"/dev",
	"/mnt",
	"/opt",
	"/srv",
	"/data",
	"/log/",
	"/logs/",
	"/temp",
	"/run",
	"/lock/",
	"/pid",
	"/socket",
	"/pipe",
	"/fifo",
	"/null",
	"/zero",
	"/random",
	"/urandom",
	"/full",
	"/tty",
	"/console",
	"/pts",
	"/ptmx",
	"/fd",
	"/stdin",
	"/stdout",
	"/stderr",
}

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
}

func SecurityMiddleware() MiddlewareFunc {
	batcher := getBatcher()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getRemoteIP(r)
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
	for _, prefix := range maliciousPathAllowedPrefixes {
		prefixLower := strings.ToLower(prefix)
		if pathLower == prefixLower || strings.HasPrefix(pathLower, prefixLower+"/") {
			return false
		}
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
		patternLower := strings.ToLower(strings.TrimPrefix(pattern, "/"))
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
