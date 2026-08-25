package http

import (
	"reflect"
	"testing"
)

func TestSetQueryScalarRejectsMalformedIntegerWithoutMutatingField(t *testing.T) {
	target := 17
	field := reflect.ValueOf(&target).Elem()

	if setQueryScalar(field, "not-a-number", "limit") {
		t.Fatal("setQueryScalar() accepted malformed integer")
	}
	if target != 17 {
		t.Fatalf("target = %d, want unchanged value 17", target)
	}
}

func TestSetQueryScalarRejectsMalformedBooleanWithoutMutatingField(t *testing.T) {
	target := true
	field := reflect.ValueOf(&target).Elem()

	if setQueryScalar(field, "maybe", "enabled") {
		t.Fatal("setQueryScalar() accepted malformed boolean")
	}
	if !target {
		t.Fatal("target changed after malformed boolean")
	}
}

func TestSetQueryScalarParsesValidFloat(t *testing.T) {
	var target float64
	field := reflect.ValueOf(&target).Elem()

	if !setQueryScalar(field, "12.5", "amount") {
		t.Fatal("setQueryScalar() rejected valid float")
	}
	if target != 12.5 {
		t.Fatalf("target = %v, want 12.5", target)
	}
}
