package http

import (
	"net/http"
	"regexp"
	"strings"
)

func getAllowedOrigins() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`^http://localhost(:\d*)?$`),
		regexp.MustCompile(`^http://voxire\.com$`),
		regexp.MustCompile(`^https://voxire\.com$`),
		regexp.MustCompile(`^https?://.*\.voxire\.com$`),
		regexp.MustCompile(`^http://voxelect\.com$`),
		regexp.MustCompile(`^https://voxelect\.com$`),
		regexp.MustCompile(`^https?://.*\.voxelect\.com$`),
		regexp.MustCompile(`^https?://.*\.voxire\.tech$`),
		regexp.MustCompile(`^https?://(.*\.)?rtylr\.com$`),
	}
}

func allowedHeaders() []string {
	return []string{
		"Authorization",
		"Content-Type",
		"Accept",
		"Origin",
		"User-Agent",
		"Referer",
		"X-Requested-With",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Methods",
		"Access-Control-Expose-Headers",
		"x-request-id",
		"profile_token",
		"client_platform",
		"X-Product-Platform",
		"Product-Platform",
		"client_version",
		"language",
		"client_ip",
	}
}

func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := false
			allowedOrigins := getAllowedOrigins()

			for _, regex := range allowedOrigins {
				if regex.MatchString(origin) {
					allowed = true

					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders(), ", "))
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
			w.Header().Set("Access-Control-Max-Age", "600")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
