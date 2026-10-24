package http

import "strings"

// maliciousPatterns are the first path segments vulnerability scanners probe
// for: WordPress admin pages, dotfiles, framework consoles, OS directories. The
// API serves none of them, so the request logger drops their failures instead
// of reporting them as errors. Blocking the probes is the Cloudflare WAF's job,
// in front of the API.
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
	"/static",
	"/assets",
	"/css",
	"/js",
	"/fonts",
	"/media",
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

var maliciousPathAllowedPrefixes = []string{
	"/console",
	"/upload",
	"/files",
	"/images",
	"/uploads",
	// Registered anonymous route (GET /.well-known/rtylr.json). Scanners probe
	// /.well-known/ too, but this route's errors must still be logged.
	"/.well-known",
}

// isMaliciousPath reports whether the path's first segment is a scanner-probe
// pattern from maliciousPatterns.
func isMaliciousPath(path string) bool {
	pathLower := strings.ToLower(path)
	if isAllowListedPath(pathLower) {
		return false
	}

	// Split into segments so that /hr/setup/status does not match the /setup pattern.
	// A pattern like "/setup" should only match a request whose path IS "/setup" or
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

// isAllowListedPath reports whether the path belongs to a real route group whose
// name collides with a scanner-probe pattern (CDN/upload/files/console routes),
// so its errors are logged like any other route's.
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
