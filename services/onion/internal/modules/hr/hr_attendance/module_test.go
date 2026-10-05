package hr_attendance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutesIncludesValidatedSetPINPath(t *testing.T) {
	router := chi.NewRouter()
	(&Module{}).RegisterRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodPut, "/hr/employees/me/pin", strings.NewReader(`{"pin":"12"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; a missing route would return 404", recorder.Code, http.StatusBadRequest)
	}
}
