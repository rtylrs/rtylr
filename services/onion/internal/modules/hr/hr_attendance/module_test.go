package hr_attendance

import (
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

)

func TestRegisterRoutesIncludesValidatedSetPINPath(t *testing.T) {
	router := serverHttp.NewLegacyRoutes()
	(&Module{}).RegisterRoutes(context.Background(), router)

	req := httptest.NewRequest(http.MethodPut, "/hr/employees/me/pin", strings.NewReader(`{"pin":"12"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.Handler().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; a missing route would return 404", recorder.Code, http.StatusBadRequest)
	}
}
