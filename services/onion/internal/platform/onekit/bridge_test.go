package onekit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	onekithttp "github.com/1homsi/onekit/http"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"

	platformerrors "github.com/voxire/onion/internal/platform/errors"
)

type stubLegacyHandler struct {
	serverError *platformerrors.ServerError
}

func (h stubLegacyHandler) Serve(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, *platformerrors.ServerError) {
	if h.serverError != nil {
		return nil, h.serverError
	}

	return &emptypb.Empty{}, nil
}

func TestServeAdaptsLegacyHandlerError(t *testing.T) {
	want := &platformerrors.ServerError{
		HTTPCode: http.StatusForbidden,
		Code:     http.StatusForbidden,
		Message:  "nope",
	}

	_, err := Serve(context.Background(), stubLegacyHandler{serverError: want}, &emptypb.Empty{})
	if !errors.Is(err, want) {
		t.Fatalf("expected legacy server error, got %v", err)
	}
}

func TestErrorHandlerWritesOnionEnvelopeForValidationError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/demo", nil)

	ErrorHandler(recorder, request, &onekithttp.ValidationError{
		Violations: []*onekithttp.FieldViolation{
			{Field: "email", Description: "must be valid"},
		},
	})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var body platformerrors.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "error" || body.Error.Code != http.StatusBadRequest || body.Error.Message != "email: must be valid" {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

func TestMountRegistersGeneratedPatterns(t *testing.T) {
	router := chi.NewRouter()
	mux := NewMux()
	mux.HandleFunc("GET /demo/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := Mount(router, mux, "GET /demo/{id}"); err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/demo/123", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected mounted mux to handle request, got %d", recorder.Code)
	}
}
