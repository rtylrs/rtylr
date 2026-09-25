package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestDecodeJSONBodyAcceptsSingleObject(t *testing.T) {
	var req struct {
		Name string `json:"name"`
	}
	r := httptest.NewRequest(http.MethodPost, "/console/test", strings.NewReader(`{"name":"Voxire"}`))
	w := httptest.NewRecorder()

	if !decodeJSONBody(w, r, &req) {
		t.Fatalf("expected decode to succeed, got status %d body %s", w.Code, w.Body.String())
	}
	if req.Name != "Voxire" {
		t.Fatalf("expected decoded name, got %q", req.Name)
	}
}

func TestDecodeJSONBodyRejectsTrailingJSON(t *testing.T) {
	var req struct {
		Name string `json:"name"`
	}
	r := httptest.NewRequest(http.MethodPost, "/console/test", strings.NewReader(`{"name":"Voxire"}{"name":"Rtylr"}`))
	w := httptest.NewRecorder()

	if decodeJSONBody(w, r, &req) {
		t.Fatal("expected trailing JSON to be rejected")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDecodeJSONBodyRejectsOversizedBody(t *testing.T) {
	var req struct {
		Value string `json:"value"`
	}
	body := `{"value":"` + strings.Repeat("x", consoleJSONBodyLimit+1) + `"}`
	r := httptest.NewRequest(http.MethodPost, "/console/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	if decodeJSONBody(w, r, &req) {
		t.Fatal("expected oversized JSON body to be rejected")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDecodeOptionalJSONBodyAllowsKnownEmptyBody(t *testing.T) {
	var req struct {
		Reason string `json:"reason"`
	}
	r := httptest.NewRequest(http.MethodPost, "/console/test", http.NoBody)
	w := httptest.NewRecorder()

	if !decodeOptionalJSONBody(w, r, &req) {
		t.Fatalf("expected empty optional body to be accepted, got status %d", w.Code)
	}
	if req.Reason != "" {
		t.Fatalf("expected zero-value request, got %q", req.Reason)
	}
}

func TestDecodeOptionalJSONBodyRejectsMalformedBody(t *testing.T) {
	var req struct {
		Reason string `json:"reason"`
	}
	r := httptest.NewRequest(http.MethodPost, "/console/test", strings.NewReader(`{"reason":`))
	w := httptest.NewRecorder()

	if decodeOptionalJSONBody(w, r, &req) {
		t.Fatal("expected malformed optional body to be rejected")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetPositiveQueryIntUsesFallbackForMalformedValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/console/resources?limit=not-a-number", nil)

	if got := getPositiveQueryInt(r, "limit", 50); got != 50 {
		t.Fatalf("expected malformed query value to use fallback, got %d", got)
	}
}

func TestUpdateResourceRowRequestRequiresValues(t *testing.T) {
	validate := validator.New()

	if err := validate.Struct(&updateResourceRowRequest{}); err == nil {
		t.Fatal("expected missing values to fail validation")
	}
	if err := validate.Struct(&updateResourceRowRequest{Values: map[string]any{}}); err == nil {
		t.Fatal("expected empty values to fail validation")
	}
	if err := validate.Struct(&updateResourceRowRequest{Values: map[string]any{"name": "updated"}}); err != nil {
		t.Fatalf("expected writable values to pass validation: %v", err)
	}
}
