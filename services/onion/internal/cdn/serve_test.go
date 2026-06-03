package cdn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// serveTransformed must reject invalid params with 400 before it ever touches
// the database or storage, so these run against a CDN with nil deps.
func TestServeTransformed_RejectsOversizeWidth(t *testing.T) {
	c := &CDN{}
	r := chi.NewRouter()
	r.Get("/images/{id}", c.serveTransformed)

	req := httptest.NewRequest("GET", "/images/abc?w=99999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversize width, got %d", rec.Code)
	}
}

func TestServeTransformed_RejectsMissingDimensions(t *testing.T) {
	c := &CDN{}
	r := chi.NewRouter()
	r.Get("/images/{id}", c.serveTransformed)

	req := httptest.NewRequest("GET", "/images/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when neither width nor height is given, got %d", rec.Code)
	}
}
