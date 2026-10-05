package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/domain"
)

func TestIsValidEquipmentStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.EquipmentStatusPending,
		domain.EquipmentStatusApproved,
		domain.EquipmentStatusRejected,
		domain.EquipmentStatusFulfilled,
	} {
		if !domain.IsValidEquipmentStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidEquipmentStatus_Invalid(t *testing.T) {
	if domain.IsValidEquipmentStatus("cancelled") {
		t.Error("expected 'cancelled' to be invalid")
	}
}

func TestIsValidAssetStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.AssetStatusAvailable,
		domain.AssetStatusAssigned,
		domain.AssetStatusRetired,
		domain.AssetStatusRepair,
	} {
		if !domain.IsValidAssetStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidAssetStatus_Invalid(t *testing.T) {
	if domain.IsValidAssetStatus("broken") {
		t.Error("expected 'broken' to be invalid")
	}
}

func TestIsValidAssetCondition_AllValid(t *testing.T) {
	for _, condition := range []string{
		domain.AssetConditionNew,
		domain.AssetConditionGood,
		domain.AssetConditionFair,
		domain.AssetConditionDamaged,
	} {
		if !domain.IsValidAssetCondition(condition) {
			t.Errorf("expected %q to be a valid asset condition", condition)
		}
	}
}

func TestIsValidAssetCondition_Invalid(t *testing.T) {
	if domain.IsValidAssetCondition("available") {
		t.Fatal("lifecycle status must not be accepted as a physical condition")
	}
}

func TestValidateEquipmentRequest_Valid(t *testing.T) {
	req := domain.EquipmentRequestRecord{
		ItemName: "Laptop",
		Quantity: 1,
		Status:   domain.EquipmentStatusPending,
	}
	if err := domain.ValidateEquipmentRequest(req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateEquipmentRequest_EmptyItemName(t *testing.T) {
	req := domain.EquipmentRequestRecord{ItemName: "", Quantity: 1}
	if err := domain.ValidateEquipmentRequest(req); !errors.Is(err, domain.ErrEmptyItemName) {
		t.Errorf("expected ErrEmptyItemName, got %v", err)
	}
}

func TestValidateEquipmentRequest_WhitespaceOnlyName(t *testing.T) {
	req := domain.EquipmentRequestRecord{ItemName: "   ", Quantity: 1}
	if err := domain.ValidateEquipmentRequest(req); !errors.Is(err, domain.ErrEmptyItemName) {
		t.Errorf("expected ErrEmptyItemName for whitespace name, got %v", err)
	}
}

func TestValidateEquipmentRequest_ZeroQuantity(t *testing.T) {
	req := domain.EquipmentRequestRecord{ItemName: "Mouse", Quantity: 0}
	if err := domain.ValidateEquipmentRequest(req); !errors.Is(err, domain.ErrNonPositiveQuantity) {
		t.Errorf("expected ErrNonPositiveQuantity for 0 quantity, got %v", err)
	}
}

func TestValidateEquipmentRequest_NegativeQuantity(t *testing.T) {
	req := domain.EquipmentRequestRecord{ItemName: "Keyboard", Quantity: -1}
	if err := domain.ValidateEquipmentRequest(req); !errors.Is(err, domain.ErrNonPositiveQuantity) {
		t.Errorf("expected ErrNonPositiveQuantity, got %v", err)
	}
}

func TestValidateAsset_Valid(t *testing.T) {
	a := domain.AssetRecord{Name: "MacBook Pro", Status: domain.AssetStatusAvailable}
	if err := domain.ValidateAsset(a); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateAsset_EmptyName(t *testing.T) {
	a := domain.AssetRecord{Name: "", Status: domain.AssetStatusAvailable}
	if err := domain.ValidateAsset(a); !errors.Is(err, domain.ErrEmptyAssetName) {
		t.Errorf("expected ErrEmptyAssetName, got %v", err)
	}
}

func TestValidateAsset_InvalidStatus(t *testing.T) {
	a := domain.AssetRecord{Name: "Monitor", Status: "broken"}
	if err := domain.ValidateAsset(a); !errors.Is(err, domain.ErrInvalidAssetStatus) {
		t.Errorf("expected ErrInvalidAssetStatus, got %v", err)
	}
}

func TestValidateAsset_EmptyStatusIsAllowed(t *testing.T) {
	a := domain.AssetRecord{Name: "Desk", Status: ""}
	if err := domain.ValidateAsset(a); err != nil {
		t.Errorf("unexpected error for empty status: %v", err)
	}
}

func TestValidateAsset_InvalidCondition(t *testing.T) {
	a := domain.AssetRecord{Name: "Desk", Status: domain.AssetStatusAvailable, Condition: "available"}
	if err := domain.ValidateAsset(a); !errors.Is(err, domain.ErrInvalidAssetCondition) {
		t.Errorf("expected ErrInvalidAssetCondition, got %v", err)
	}
}
