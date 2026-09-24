package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyItemName          = errors.New("equipment item name is required")
	ErrNonPositiveQuantity    = errors.New("quantity must be at least 1")
	ErrInvalidEquipmentStatus = errors.New("invalid equipment request status")
	ErrInvalidAssetStatus     = errors.New("invalid asset status")
	ErrInvalidAssetCondition  = errors.New("invalid asset condition")
	ErrEmptyAssetName         = errors.New("asset name is required")
)

// IsValidEquipmentStatus returns true when the status is recognised.
func IsValidEquipmentStatus(status string) bool {
	switch status {
	case EquipmentStatusPending, EquipmentStatusApproved, EquipmentStatusRejected, EquipmentStatusFulfilled:
		return true
	}
	return false
}

// IsValidAssetStatus returns true when the status is recognised.
func IsValidAssetStatus(status string) bool {
	switch status {
	case AssetStatusAvailable, AssetStatusAssigned, AssetStatusRetired, AssetStatusRepair:
		return true
	}
	return false
}

// IsValidAssetCondition reports whether a physical-condition value matches
// the generated AssetPayload contract. It is intentionally separate from the
// lifecycle status predicate.
func IsValidAssetCondition(condition string) bool {
	switch condition {
	case AssetConditionNew, AssetConditionGood, AssetConditionFair, AssetConditionDamaged:
		return true
	}
	return false
}

// ValidateEquipmentRequest checks that required fields are present.
func ValidateEquipmentRequest(req EquipmentRequestRecord) error {
	if strings.TrimSpace(req.ItemName) == "" {
		return ErrEmptyItemName
	}
	if req.Quantity < 1 {
		return ErrNonPositiveQuantity
	}
	return nil
}

// ValidateAsset checks that required asset fields are present.
func ValidateAsset(a AssetRecord) error {
	if strings.TrimSpace(a.Name) == "" {
		return ErrEmptyAssetName
	}
	if a.Status != "" && !IsValidAssetStatus(a.Status) {
		return ErrInvalidAssetStatus
	}
	if a.Condition != "" && !IsValidAssetCondition(a.Condition) {
		return ErrInvalidAssetCondition
	}
	return nil
}
