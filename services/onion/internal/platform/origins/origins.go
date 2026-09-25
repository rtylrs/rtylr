package origins

import (
	"regexp"

	"github.com/voxire/onion/internal/platform/config"
)

// productionPatterns lists browser origins trusted in every environment.
// Wildcards match exactly one subdomain label: a takeover chain like
// a.b.voxire.com must not inherit the apex's CORS trust, and one-label
// coverage matches how the zones are actually used (app.voxire.com etc.).
var productionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^https://voxire\.com$`),
	regexp.MustCompile(`^https://[a-z0-9-]+\.voxire\.com$`),
	regexp.MustCompile(`^https://voxelect\.com$`),
	regexp.MustCompile(`^https://[a-z0-9-]+\.voxelect\.com$`),
	regexp.MustCompile(`^https://[a-z0-9-]+\.voxire\.tech$`),
	regexp.MustCompile(`^https://rtylr\.com$`),
	regexp.MustCompile(`^https://[a-z0-9-]+\.rtylr\.com$`),
}

// developmentPatterns additionally trust local dev servers on any port. They
// are pure risk when compiled into a production binary, so they are gated to
// non-production environments (#777).
var developmentPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^http://localhost(:\d+)?$`),
	regexp.MustCompile(`^http://127\.0\.0\.1(:\d+)?$`),
}

var productionPatternStrings = []string{
	"https://voxire.com",
	"https://*.voxire.com",
	"https://voxelect.com",
	"https://*.voxelect.com",
	"https://*.voxire.tech",
	"https://rtylr.com",
	"https://*.rtylr.com",
}

var developmentPatternStrings = []string{
	"http://localhost",
	"http://localhost:*",
	"http://127.0.0.1",
	"http://127.0.0.1:*",
}

func activePatterns() []*regexp.Regexp {
	if config.IsProduction() {
		return productionPatterns
	}
	return append(append(make([]*regexp.Regexp, 0, len(productionPatterns)+len(developmentPatterns)), productionPatterns...), developmentPatterns...)
}

func activePatternStrings() []string {
	if config.IsProduction() {
		return productionPatternStrings
	}
	return append(append(make([]string, 0, len(productionPatternStrings)+len(developmentPatternStrings)), productionPatternStrings...), developmentPatternStrings...)
}

// AllowedPatternStrings returns the same allowlist in path.Match form for APIs
// that use glob-like patterns, such as WebSocket accept options.
func AllowedPatternStrings() []string {
	return activePatternStrings()
}

// IsAllowed reports whether origin matches the trusted browser origin policy
// for the current environment.
func IsAllowed(origin string) bool {
	for _, pattern := range activePatterns() {
		if pattern.MatchString(origin) {
			return true
		}
	}
	return false
}
