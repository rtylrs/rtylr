package cursor

import (
	"errors"
	"testing"
	"time"
)

func TestCreatedAtIDRoundTrip(t *testing.T) {
	want := CreatedAtID{CreatedAt: time.Date(2026, time.August, 13, 12, 30, 1, 99, time.UTC), ID: 42}
	got, err := DecodeCreatedAtID(EncodeCreatedAtID(want))
	if err != nil {
		t.Fatal(err)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) || got.ID != want.ID {
		t.Fatalf("cursor = %+v, want %+v", got, want)
	}
}

func TestDecodeCreatedAtIDRejectsMalformedToken(t *testing.T) {
	_, err := DecodeCreatedAtID("not-a-cursor")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
}
