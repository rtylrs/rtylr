package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
)

func quoteValidationStatus(t *testing.T, err error) int {
	t.Helper()
	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an API error, got %T: %v", err, err)
	}
	return apiErr.HTTPStatusCode()
}

func TestCreateHandlerValidateUsesDomainRules(t *testing.T) {
	err := (&CreateHandler{}).Validate(context.Background(), &CreateRequest{
		CustomerName: "Acme",
		Subtotal:     -1,
	})

	if quoteValidationStatus(t, err) != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative subtotal, got %d", quoteValidationStatus(t, err))
	}
}

func TestUpdateHandlerValidateUsesDomainRules(t *testing.T) {
	err := (&UpdateHandler{}).Validate(context.Background(), &UpdateRequest{
		ID:           1,
		CustomerName: "Acme",
		Status:       "not-a-status",
	})

	if quoteValidationStatus(t, err) != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", quoteValidationStatus(t, err))
	}
}
