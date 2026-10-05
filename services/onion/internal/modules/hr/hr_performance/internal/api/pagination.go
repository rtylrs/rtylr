package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

const (
	performanceCyclesListPath  = "/hr/performance/v1/review-cycles"
	performanceGoalsListPath   = "/hr/performance/v1/goals"
	defaultPerformancePageSize = int32(20)
	maximumPerformancePageSize = int32(100)
)

type performancePagination struct {
	limit  int32
	offset int32
}

type performancePaginationContextKey struct{}

func defaultPerformancePagination() performancePagination {
	return performancePagination{limit: defaultPerformancePageSize, offset: 0}
}

func normalizePerformancePagination(limit, offset int32) (performancePagination, error) {
	if limit <= 0 {
		limit = defaultPerformancePageSize
	}
	if limit > maximumPerformancePageSize {
		limit = maximumPerformancePageSize
	}
	if offset < 0 {
		offset = 0
	}
	return performancePagination{limit: limit, offset: offset}, nil
}

func parsePerformancePagination(r *http.Request) (performancePagination, error) {
	var limit, offset int32
	query := r.URL.Query()
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return performancePagination{}, onekiterrors.New(http.StatusBadRequest, "limit must be an integer")
		}
		limit = int32(parsed) // #nosec G115 -- ParseInt is bounded to int32 above.
	}
	if raw := query.Get("offset"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return performancePagination{}, onekiterrors.New(http.StatusBadRequest, "offset must be an integer")
		}
		offset = int32(parsed) // #nosec G115 -- ParseInt is bounded to int32 above.
	}
	return normalizePerformancePagination(limit, offset)
}

func performancePaginationFromContext(ctx context.Context) performancePagination {
	if pagination, ok := ctx.Value(performancePaginationContextKey{}).(performancePagination); ok {
		return pagination
	}
	return defaultPerformancePagination()
}

// PerformancePaginationMiddleware preserves pagination parameters for the
// generated v1 list routes. The pinned generated request types predate these
// query fields, so the generated binder otherwise discards them before the
// module can apply them.
func PerformancePaginationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || (r.URL.Path != performanceCyclesListPath && r.URL.Path != performanceGoalsListPath) {
			next.ServeHTTP(w, r)
			return
		}

		pagination, err := parsePerformancePagination(r)
		if err != nil {
			serverHttp.WriteOneKitError(r.Context(), w, err)
			return
		}

		ctx := context.WithValue(r.Context(), performancePaginationContextKey{}, pagination)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
