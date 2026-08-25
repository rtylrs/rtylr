package http

import (
	"context"
	"fmt"
	"net/http"
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
	requests map[string][]blockedRequest
	mutex    sync.RWMutex
	ticker   *time.Ticker
	done     chan bool
}

var (
	batcher *securityBatcher
	once    sync.Once
)

func getBatcher() *securityBatcher {
	once.Do(func() {
		batcher = &securityBatcher{
			requests: make(map[string][]blockedRequest),
			ticker:   time.NewTicker(30 * time.Second),
			done:     make(chan bool),
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
	snapshot := b.requests
	b.requests = make(map[string][]blockedRequest)
	b.mutex.Unlock()

	if len(snapshot) == 0 {
		return
	}

	ctx := context.Background()
	for ip, reqs := range snapshot {
		logger.Ctx(ctx).Warn("blocked suspicious requests batch",
			logger.String("ip", ip),
			logger.Int("count", len(reqs)),
		)
		for _, req := range reqs {
			slack.Post(slack.Suspicious, fmt.Sprintf("⚠️ blocked request · %s %s · %s · %s", req.Method, req.Path, req.Reason, req.IP))
			logger.Ctx(ctx).Warn("blocked suspicious request",
				logger.String("ip", req.IP),
				logger.String("reason", req.Reason),
				logger.String("method", req.Method),
				logger.String("path", req.Path),
				logger.String("user_agent", req.UserAgent),
			)
		}
	}
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
