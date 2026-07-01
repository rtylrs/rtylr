package origins

import "regexp"

var allowedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^http://localhost(:\d*)?$`),
	regexp.MustCompile(`^http://127\.0\.0\.1(:\d*)?$`),
	regexp.MustCompile(`^https://voxire\.com$`),
	regexp.MustCompile(`^https://.*\.voxire\.com$`),
	regexp.MustCompile(`^https://voxelect\.com$`),
	regexp.MustCompile(`^https://.*\.voxelect\.com$`),
	regexp.MustCompile(`^https://.*\.voxire\.tech$`),
	regexp.MustCompile(`^https://(.*\.)?rtylr\.com$`),
}

// AllowedPatterns returns the browser origins trusted by both HTTP CORS and
// WebSocket upgrade handling. Patterns are compiled once at package init.
func AllowedPatterns() []*regexp.Regexp {
	return allowedPatterns
}

// AllowedPatternStrings returns the same allowlist in path.Match form for APIs
// that use glob-like patterns, such as WebSocket accept options.
func AllowedPatternStrings() []string {
	return []string{
		"http://localhost",
		"http://localhost:*",
		"http://127.0.0.1",
		"http://127.0.0.1:*",
		"https://voxire.com",
		"https://*.voxire.com",
		"https://voxelect.com",
		"https://*.voxelect.com",
		"https://*.voxire.tech",
		"https://rtylr.com",
		"https://*.rtylr.com",
	}
}

// IsAllowed reports whether origin matches the trusted browser origin policy.
func IsAllowed(origin string) bool {
	for _, pattern := range allowedPatterns {
		if pattern.MatchString(origin) {
			return true
		}
	}
	return false
}
