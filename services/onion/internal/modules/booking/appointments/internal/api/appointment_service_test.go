package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	bookingappointmentv1 "github.com/voxire/voxie/gen/go/booking/appointment/v1"
)

func TestUpdateAppointmentRejectsMissingTimes(t *testing.T) {
	api := &AppointmentAPI{}

	_, err := api.UpdateAppointment(context.Background(), &bookingappointmentv1.UpdateAppointmentRequest{
		Id: 1,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var oneKitErr *onekithttp.Error
	if !stdErrors.As(err, &oneKitErr) {
		t.Fatalf("expected oneKit error, got %T", err)
	}
	if got := int(oneKitErr.GetStatusCode()); got != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", got, http.StatusBadRequest)
	}
}

func TestUpdateAppointmentRejectsMissingEndTime(t *testing.T) {
	api := &AppointmentAPI{}

	_, err := api.UpdateAppointment(context.Background(), &bookingappointmentv1.UpdateAppointmentRequest{
		Id:        1,
		StartTime: "2026-07-29T10:00:00Z",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var oneKitErr *onekithttp.Error
	if !stdErrors.As(err, &oneKitErr) {
		t.Fatalf("expected oneKit error, got %T", err)
	}
	if got := int(oneKitErr.GetStatusCode()); got != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", got, http.StatusBadRequest)
	}
}
