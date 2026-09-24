package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAttendancePaginationMiddlewarePassesValidatedPage(t *testing.T) {
	var got attendancePagination
	handler := AttendancePaginationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = attendancePaginationFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, attendanceListPath+"?limit=7&offset=14", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got.limit != 7 || got.offset != 14 {
		t.Fatalf("pagination = limit %d offset %d, want 7/14", got.limit, got.offset)
	}
}

func TestAttendancePaginationMiddlewareRejectsInvalidPage(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "negative limit", url: attendanceListPath + "?limit=-1"},
		{name: "large limit", url: attendanceListPath + "?limit=101"},
		{name: "negative offset", url: attendanceListPath + "?offset=-1"},
		{name: "non numeric offset", url: attendanceListPath + "?offset=nope"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := AttendancePaginationMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
			if called {
				t.Fatal("downstream handler should not run")
			}
		})
	}
}

func TestAttendancePaginationMiddlewareLeavesNonListRequestsAlone(t *testing.T) {
	called := false
	handler := AttendancePaginationMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodPost, attendanceListPath+"/clock-in?limit=-1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if !called {
		t.Fatal("non-list request should reach its handler")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestAttendancePaginationDefaultsWithoutQuery(t *testing.T) {
	pagination := attendancePaginationFromContext(context.Background())
	if pagination != defaultAttendancePagination() {
		t.Fatalf("pagination = %+v, want default %+v", pagination, defaultAttendancePagination())
	}
}
