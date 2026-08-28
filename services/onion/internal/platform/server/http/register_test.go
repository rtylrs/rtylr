package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
)

func TestValidateProtoMessagesRejectsInvalidGeneratedMessage(t *testing.T) {
	req := &commonauthv1.LoginRequest{}

	if err := validateProtoMessages(req); err == nil {
		t.Fatal("expected proto validation error")
	}
}

func TestValidateProtoMessagesAcceptsNestedGeneratedMessage(t *testing.T) {
	res := struct {
		Payload *commonauthv1.LoginRequest
	}{
		Payload: &commonauthv1.LoginRequest{
			Email:    "person@example.com",
			Password: "secret",
		},
	}

	if err := validateProtoMessages(res); err != nil {
		t.Fatalf("validateProtoMessages: %v", err)
	}
}

type validatedGETRequest struct {
	ID int64 `json:"id"`
}

type validatedGETResponse struct {
	OK bool `json:"ok"`
}

type validatedGETHandler struct {
	called bool
	seenID int64
}

func (h *validatedGETHandler) Validate(_ context.Context, req *validatedGETRequest) error {
	if req.ID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *validatedGETHandler) Serve(_ context.Context, req *validatedGETRequest) (validatedGETResponse, error) {
	h.called = true
	h.seenID = req.ID
	return validatedGETResponse{OK: true}, nil
}

func TestRegisterGETWithValidationRejectsInvalidQuery(t *testing.T) {
	router := chi.NewRouter()
	handler := &validatedGETHandler{}
	RegisterGETWithValidation[validatedGETRequest, validatedGETResponse](
		context.Background(),
		router,
		[]string{"/validated"},
		handler,
	)

	req := httptest.NewRequest(http.MethodGet, "/validated", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if handler.called {
		t.Fatal("handler should not run after validation failure")
	}
}

func TestRegisterGETWithValidationParsesQuery(t *testing.T) {
	router := chi.NewRouter()
	handler := &validatedGETHandler{}
	RegisterGETWithValidation[validatedGETRequest, validatedGETResponse](
		context.Background(),
		router,
		[]string{"/validated"},
		handler,
	)

	req := httptest.NewRequest(http.MethodGet, "/validated?id=7", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !handler.called || handler.seenID != 7 {
		t.Fatalf("handler saw id %d, called=%v; want id 7", handler.seenID, handler.called)
	}
}

type typedQueryRequest struct {
	Limit    int64   `json:"limit"`
	Active   bool    `json:"active"`
	Search   string  `json:"search"`
	Score    float64 `json:"score"`
	IDs      []int64 `json:"ids"`
	Optional *int64  `json:"optional"`
	Ignored  int64   `json:"-"`
	Any      any     `json:"any"`
}

func TestParseQueryParamsAssignsTypedFieldsWithoutJSONRoundTrip(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/typed?limit=25&active=true&search=register&score=1.5&ids=1&ids=2&optional=9&ignored=3&any=17",
		http.NoBody,
	)
	var got typedQueryRequest

	if err := parseQueryParams(req, &got); err != nil {
		t.Fatalf("parseQueryParams: %v", err)
	}

	if got.Limit != 25 {
		t.Fatalf("limit = %d, want 25", got.Limit)
	}
	if !got.Active {
		t.Fatal("active = false, want true")
	}
	if got.Search != "register" {
		t.Fatalf("search = %q, want register", got.Search)
	}
	if got.Score != 1.5 {
		t.Fatalf("score = %f, want 1.5", got.Score)
	}
	if len(got.IDs) != 2 || got.IDs[0] != 1 || got.IDs[1] != 2 {
		t.Fatalf("ids = %#v, want [1 2]", got.IDs)
	}
	if got.Optional == nil || *got.Optional != 9 {
		t.Fatalf("optional = %#v, want pointer to 9", got.Optional)
	}
	if got.Ignored != 0 {
		t.Fatalf("ignored = %d, want 0", got.Ignored)
	}
	if got.Any != int64(17) {
		t.Fatalf("any = %#v, want int64(17)", got.Any)
	}
}

func TestRegisterGETRejectsMalformedTypedQuery(t *testing.T) {
	router := chi.NewRouter()
	handler := &validatedGETHandler{}
	RegisterGET[validatedGETRequest, validatedGETResponse](
		context.Background(),
		router,
		[]string{"/validated"},
		handler,
	)

	req := httptest.NewRequest(http.MethodGet, "/validated?id=not-a-number", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed query, got %d: %s", rec.Code, rec.Body.String())
	}
	if handler.called {
		t.Fatal("handler should not run after query parsing failure")
	}
}

type emptyRequest struct{}

type invalidProtoResponseHandler struct{}

func (h invalidProtoResponseHandler) Serve(_ context.Context, _ *emptyRequest) (*commonauthv1.LoginRequest, error) {
	return &commonauthv1.LoginRequest{}, nil
}

func TestRegisterGETSkipsResponseProtoValidationByDefault(t *testing.T) {
	t.Setenv("VALIDATE_RESPONSE_PROTOS", "")
	t.Setenv("RESPONSE_PROTO_VALIDATION", "")

	router := chi.NewRouter()
	RegisterGET[emptyRequest, *commonauthv1.LoginRequest](
		context.Background(),
		router,
		[]string{"/proto"},
		invalidProtoResponseHandler{},
	)

	req := httptest.NewRequest(http.MethodGet, "/proto", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterGETCanValidateResponseProtosWhenEnabled(t *testing.T) {
	t.Setenv("VALIDATE_RESPONSE_PROTOS", "true")
	t.Setenv("RESPONSE_PROTO_VALIDATION", "")

	router := chi.NewRouter()
	RegisterGET[emptyRequest, *commonauthv1.LoginRequest](
		context.Background(),
		router,
		[]string{"/proto"},
		invalidProtoResponseHandler{},
	)

	req := httptest.NewRequest(http.MethodGet, "/proto", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestWriteOneKitErrorDoesNotExposeInternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteOneKitError(context.Background(), rec, errors.New("duplicate entry in users.email: secret@example.com"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "duplicate entry") || strings.Contains(body, "secret@example.com") {
		t.Fatalf("response leaked internal error: %s", body)
	}
	if !strings.Contains(body, "Internal server error") {
		t.Fatalf("response missing generic error message: %s", body)
	}
}

type jsonBodyRequest struct {
	Name string `json:"name"`
}

type jsonBodyResponse struct {
	Name string `json:"name"`
}

type jsonBodyHandler struct {
	called bool
	seen   string
}

func (h *jsonBodyHandler) Serve(_ context.Context, req *jsonBodyRequest) (jsonBodyResponse, error) {
	h.called = true
	h.seen = req.Name
	return jsonBodyResponse{Name: req.Name}, nil
}

func TestRegisterPOSTRejectsTrailingJSON(t *testing.T) {
	router := chi.NewRouter()
	handler := &jsonBodyHandler{}
	RegisterPOST[jsonBodyRequest, jsonBodyResponse](
		context.Background(),
		router,
		[]string{"/json"},
		handler,
	)

	req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(`{"name":"one"}{"name":"two"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if handler.called {
		t.Fatal("handler should not run after trailing JSON")
	}
}

func TestRegisterPOSTAcceptsWhitespaceAfterJSON(t *testing.T) {
	router := chi.NewRouter()
	handler := &jsonBodyHandler{}
	RegisterPOST[jsonBodyRequest, jsonBodyResponse](
		context.Background(),
		router,
		[]string{"/json"},
		handler,
	)

	req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader("{\"name\":\"one\"}\n\t "))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !handler.called || handler.seen != "one" {
		t.Fatalf("handler saw name %q, called=%v; want one", handler.seen, handler.called)
	}
}

func TestRegisterDELETERejectsTrailingJSONBody(t *testing.T) {
	router := chi.NewRouter()
	handler := &jsonBodyHandler{}
	RegisterDELETE[jsonBodyRequest, jsonBodyResponse](
		context.Background(),
		router,
		[]string{"/json"},
		handler,
	)

	req := httptest.NewRequest(http.MethodDelete, "/json", strings.NewReader(`{"name":"one"}{"name":"two"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if handler.called {
		t.Fatal("handler should not run after trailing JSON")
	}
}
