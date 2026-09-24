package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPerformancePaginationMiddlewarePassesValidatedPage(t *testing.T) {
	var got performancePagination
	handler := PerformancePaginationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = performancePaginationFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, performanceGoalsListPath+"?limit=7&offset=14", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got.limit != 7 || got.offset != 14 {
		t.Fatalf("pagination = limit %d offset %d, want 7/14", got.limit, got.offset)
	}
}

func TestPerformancePaginationMiddlewareRejectsInvalidPage(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "non numeric limit", url: performanceCyclesListPath + "?limit=nope"},
		{name: "overflowing limit", url: performanceCyclesListPath + "?limit=999999999999"},
		{name: "non numeric offset", url: performanceGoalsListPath + "?offset=nope"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := PerformancePaginationMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
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

func TestPerformancePaginationMiddlewareClampsBounds(t *testing.T) {
	var got performancePagination
	handler := PerformancePaginationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = performancePaginationFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, performanceCyclesListPath+"?limit=101&offset=-5", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got.limit != maximumPerformancePageSize || got.offset != 0 {
		t.Fatalf("pagination = limit %d offset %d, want %d/0", got.limit, got.offset, maximumPerformancePageSize)
	}
}

func TestPerformancePaginationMiddlewareLeavesNonListRequestsAlone(t *testing.T) {
	called := false
	handler := PerformancePaginationMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodPost, performanceGoalsListPath+"?limit=-1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if !called {
		t.Fatal("non-list request should reach its handler")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestPerformancePaginationDefaultsWithoutQuery(t *testing.T) {
	pagination := performancePaginationFromContext(context.Background())
	if pagination != defaultPerformancePagination() {
		t.Fatalf("pagination = %+v, want default %+v", pagination, defaultPerformancePagination())
	}
}
