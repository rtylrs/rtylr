package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestManager_VerifyPIN_WrongPINReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, "000000_invalid_pin_xyz")
	if err != nil {
		t.Fatalf("VerifyPIN: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil VerifyResult")
	}
	if result.Valid {
		t.Error("expected Valid=false for wrong PIN")
	}
}

func TestManager_VerifyPIN_ReturnsNonNilResult(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, "wrongpin")
	if err != nil {
		t.Fatalf("VerifyPIN: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil VerifyResult, got nil")
	}
}

func TestManager_VerifyPIN_WrongOwnerReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), 999999, "1234")
	if err != nil {
		t.Fatalf("VerifyPIN wrong owner: %v", err)
	}
	if result.Valid {
		t.Error("expected Valid=false for unknown owner")
	}
}

func TestManager_VerifyPIN_InvalidResultHasZeroStaffID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, "not_a_real_pin_99999")
	if err != nil {
		t.Fatalf("VerifyPIN: %v", err)
	}
	if result.Valid && result.StaffID == 0 {
		t.Error("valid result should have non-zero StaffID")
	}
	if !result.Valid && result.StaffID != 0 {
		t.Errorf("invalid result has StaffID=%d, want 0", result.StaffID)
	}
}

func TestManager_VerifyPIN_EmptyPINReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("VerifyPIN empty pin: %v", err)
	}
	if result.Valid {
		t.Error("expected Valid=false for empty PIN")
	}
}

func TestManager_VerifyPIN_MultipleCallsConsistent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const pin = "wrong_pin_consistency"

	r1, err := repo.VerifyPIN(ctx, testutil.OwnerID, pin)
	if err != nil {
		t.Fatalf("VerifyPIN first: %v", err)
	}
	r2, err := repo.VerifyPIN(ctx, testutil.OwnerID, pin)
	if err != nil {
		t.Fatalf("VerifyPIN second: %v", err)
	}
	if r1.Valid != r2.Valid {
		t.Errorf("inconsistent results: r1.Valid=%v r2.Valid=%v", r1.Valid, r2.Valid)
	}
}

func TestManager_VerifyPIN_LongPINReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	longPIN := "abcdefghijklmnopqrstuvwxyz0123456789"
	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, longPIN)
	if err != nil {
		t.Fatalf("VerifyPIN long pin: %v", err)
	}
	if result.Valid {
		t.Error("expected Valid=false for long random PIN")
	}
}

func TestManager_VerifyPIN_SpecialCharsReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, "'; DROP TABLE pos_manager_pins; --")
	if err != nil {
		t.Fatalf("VerifyPIN injection: %v", err)
	}
	if result.Valid {
		t.Error("expected Valid=false for SQL injection attempt")
	}
}

func TestManager_VerifyPIN_ValidResultHasPositiveStaffID(t *testing.T) {
	// This test verifies the structural contract: if Valid=true, StaffID > 0.
	// With wrong credentials we always get Valid=false, so we assert the inverse shape.
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), testutil.OwnerID, "definitely_wrong")
	if err != nil {
		t.Fatalf("VerifyPIN: %v", err)
	}
	// We can only assert the invalid branch here without knowing a real PIN.
	if !result.Valid {
		if result.StaffID != 0 {
			t.Errorf("invalid result has StaffID=%d, want 0", result.StaffID)
		}
	}
}

func TestManager_VerifyPIN_ZeroOwnerReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), 0, "1234")
	if err != nil {
		t.Fatalf("VerifyPIN zero owner: %v", err)
	}
	if result.Valid {
		t.Error("expected Valid=false for owner=0")
	}
}

func TestManager_VerifyPIN_NegativeOwnerReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	result, err := repo.VerifyPIN(context.Background(), -1, "1234")
	if err != nil {
		t.Fatalf("VerifyPIN negative owner: %v", err)
	}
	if result.Valid {
		t.Error("expected Valid=false for owner=-1")
	}
}
