package stock_core

import "testing"

func TestMovementBusinessIDDefaultsToOwnerID(t *testing.T) {
	if got := movementBusinessID(Movement{OwnerID: 42}); got != 42 {
		t.Fatalf("movementBusinessID() = %d, want owner id 42", got)
	}
}

func TestMovementBusinessIDUsesExplicitBusinessID(t *testing.T) {
	if got := movementBusinessID(Movement{OwnerID: 42, BusinessID: 99}); got != 99 {
		t.Fatalf("movementBusinessID() = %d, want explicit business id 99", got)
	}
}
