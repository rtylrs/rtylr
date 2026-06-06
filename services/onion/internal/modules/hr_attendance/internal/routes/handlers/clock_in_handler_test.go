package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
)

func TestClockInValidateRejectsInvalidMethod(t *testing.T) {
	h := &ClockInHandler{}

	err := h.Validate(context.Background(), &ClockInRequest{
		OwnerID:       1,
		EmployeeID:    2,
		ClockInMethod: "badge",
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", err.Code)
	}
}

func TestClockInValidateRejectsAuthenticatedOnlyMethod(t *testing.T) {
	h := &ClockInHandler{}

	err := h.Validate(context.Background(), &ClockInRequest{
		OwnerID:       1,
		EmployeeID:    2,
		ClockInMethod: domain.ClockInMethodGPS,
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", err.Code)
	}
}

func TestClockInValidateRequiresProofForPublicMethods(t *testing.T) {
	h := &ClockInHandler{}

	err := h.Validate(context.Background(), &ClockInRequest{
		OwnerID:       1,
		EmployeeID:    2,
		ClockInMethod: domain.ClockInMethodPin,
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", err.Code)
	}
}
