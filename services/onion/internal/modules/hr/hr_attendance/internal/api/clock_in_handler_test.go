package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
)

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

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
	if oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", oneKitStatus(err))
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
	if oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", oneKitStatus(err))
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
	if oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", oneKitStatus(err))
	}
}
