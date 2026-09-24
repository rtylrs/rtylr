package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

const (
	attendanceListPath         = "/hr/attendance/v1/attendance"
	defaultAttendancePageLimit = int32(50)
	maximumAttendancePageLimit = int32(100)
)

type attendancePagination struct {
	limit  int32
	offset int32
}

type attendancePaginationContextKey struct{}

func defaultAttendancePagination() attendancePagination {
	return attendancePagination{limit: defaultAttendancePageLimit, offset: 0}
}

func normalizeAttendancePagination(limit, offset int32) (attendancePagination, error) {
	if limit < 0 {
		return attendancePagination{}, onekiterrors.New(http.StatusBadRequest, "limit must be non-negative")
	}
	if limit > maximumAttendancePageLimit {
		return attendancePagination{}, onekiterrors.New(http.StatusBadRequest, "limit must be at most 100")
	}
	if offset < 0 {
		return attendancePagination{}, onekiterrors.New(http.StatusBadRequest, "offset must be non-negative")
	}
	if limit == 0 {
		limit = defaultAttendancePageLimit
	}
	return attendancePagination{limit: limit, offset: offset}, nil
}

func parseAttendancePagination(values url.Values) (attendancePagination, error) {
	var limit, offset int32

	if raw := values.Get("limit"); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 32)
		if parseErr != nil {
			return attendancePagination{}, onekiterrors.New(http.StatusBadRequest, "limit must be an integer")
		}
		limit = int32(parsed) // #nosec G115 -- ParseInt is bounded to int32 above.
	}
	if raw := values.Get("offset"); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 32)
		if parseErr != nil {
			return attendancePagination{}, onekiterrors.New(http.StatusBadRequest, "offset must be an integer")
		}
		offset = int32(parsed) // #nosec G115 -- ParseInt is bounded to int32 above.
	}

	return normalizeAttendancePagination(limit, offset)
}

func attendancePaginationFromContext(ctx context.Context) attendancePagination {
	if pagination, ok := ctx.Value(attendancePaginationContextKey{}).(attendancePagination); ok {
		return pagination
	}
	return defaultAttendancePagination()
}

// AttendancePaginationMiddleware preserves pagination parameters for the
// generated v1 list route. The pinned generated request type predates these
// query fields, so the generated binder otherwise discards them before the
// module can apply them.
func AttendancePaginationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != attendanceListPath {
			next.ServeHTTP(w, r)
			return
		}

		pagination, err := parseAttendancePagination(r.URL.Query())
		if err != nil {
			serverHttp.WriteOneKitError(r.Context(), w, err)
			return
		}

		ctx := context.WithValue(r.Context(), attendancePaginationContextKey{}, pagination)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
