package application

import (
	"math"
	"testing"
)

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
