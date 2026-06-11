package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"

	platformErrors "github.com/voxire/onion/internal/platform/errors"
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

func (h *validatedGETHandler) Validate(_ context.Context, req *validatedGETRequest) *platformErrors.ServerError {
	if req.ID <= 0 {
		return &platformErrors.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "id is required",
		}
	}
	return nil
}

func (h *validatedGETHandler) Serve(_ context.Context, req *validatedGETRequest) (validatedGETResponse, *platformErrors.ServerError) {
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
