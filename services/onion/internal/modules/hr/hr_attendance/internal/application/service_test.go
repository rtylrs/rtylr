package application

import (
	"context"
	"math"
	"testing"
)

func TestPINFailureFallbackLocksAtThreshold(t *testing.T) {
	service := &Service{}
	ctx := context.Background()
	ownerID, employeeID := int64(910001), int64(910002)
	key := service.pinFailureKey(ownerID, employeeID)
	service.ClearPINFailures(ctx, ownerID, employeeID)
	t.Cleanup(func() { service.ClearPINFailures(ctx, ownerID, employeeID) })

	for attempt := 1; attempt <= pinMaxFailures; attempt++ {
		service.RegisterPINFailure(ctx, ownerID, employeeID)
		locked := service.IsPINLocked(ctx, ownerID, employeeID)
		if locked != (attempt == pinMaxFailures) {
			t.Fatalf("locked after attempt %d = %v", attempt, locked)
		}
	}
	if got := localPINFailureCount(key); got > 0 {
		// A Redis-backed run may keep the counter out of the local fallback.
		// When the fallback is active, the threshold must be represented there.
		if got < pinMaxFailures {
			t.Fatalf("fallback count = %d, want at least %d", got, pinMaxFailures)
		}
	}
	service.ClearPINFailures(ctx, ownerID, employeeID)
	if service.IsPINLocked(ctx, ownerID, employeeID) {
		t.Fatal("clear did not reset the PIN lockout")
	}
}

func TestPINFailureFallbackIsScopedByOwnerAndEmployee(t *testing.T) {
	service := &Service{}
	keyA := service.pinFailureKey(920001, 920002)
	keyB := service.pinFailureKey(920001, 920003)
	clearLocalPINFailures(keyA)
	clearLocalPINFailures(keyB)
	t.Cleanup(func() {
		clearLocalPINFailures(keyA)
		clearLocalPINFailures(keyB)
	})

	for i := 0; i < pinMaxFailures; i++ {
		incrementLocalPINFailure(keyA)
	}
	if got := localPINFailureCount(keyA); got < pinMaxFailures {
		t.Fatalf("key A count = %d, want at least %d", got, pinMaxFailures)
	}
	if got := localPINFailureCount(keyB); got != 0 {
		t.Fatalf("key B count = %d, want 0", got)
	}
}

func TestPINFailureKeyIncludesBothScopes(t *testing.T) {
	service := &Service{}
	if a, b := service.pinFailureKey(1, 2), service.pinFailureKey(2, 1); a == b {
		t.Fatalf("keys must differ for reversed scopes: %s", a)
	}
}

func TestHaversineDistance_SamePointIsZero(t *testing.T) {
	if d := haversineDistance(33.8938, 35.5018, 33.8938, 35.5018); d > 0.001 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestHaversineDistance_BeirutToBaalbek(t *testing.T) {
	// Beirut (33.8938, 35.5018) → Baalbek (34.0058, 36.2039)
	// Real-world great-circle distance ≈ 65 km. Allow ±2 km tolerance.
	d := haversineDistance(33.8938, 35.5018, 34.0058, 36.2039)
	const wantMeters = 65_000
	if math.Abs(d-wantMeters) > 2_000 {
		t.Errorf("expected ~65 km, got %v meters", d)
	}
}

func TestHaversineDistance_Symmetric(t *testing.T) {
	a := haversineDistance(0, 0, 10, 10)
	b := haversineDistance(10, 10, 0, 0)
	if math.Abs(a-b) > 0.01 {
		t.Errorf("haversine should be symmetric: a=%v b=%v", a, b)
	}
}

func TestHaversineDistance_AntipodalIsHalfCircumference(t *testing.T) {
	// Distance between (0,0) and (0,180) — half of 2*pi*R = 20015 km.
	d := haversineDistance(0, 0, 0, 180)
	const wantMeters = 20_015_086.0
	if math.Abs(d-wantMeters) > 5_000 {
		t.Errorf("expected ~20015 km, got %v meters", d)
	}
}
