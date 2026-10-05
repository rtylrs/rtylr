package api

import (
	"reflect"
	"testing"
)

func TestAvailabilityRoutesExposeContractAndLegacyPaths(t *testing.T) {
	want := []string{"/hr/shifts/availability", "/hr/shift/v1/availability"}
	if got := ListAvailabilityRoutes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("list routes = %#v, want %#v", got, want)
	}
	if got := SetAvailabilityRoutes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("set routes = %#v, want %#v", got, want)
	}
}

func TestDeleteAvailabilityRoutesExposeContractPathParameter(t *testing.T) {
	want := []string{"/hr/shifts/availability", "/hr/shift/v1/availability/{id}"}
	if got := DeleteAvailabilityRoutes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("delete routes = %#v, want %#v", got, want)
	}
}
