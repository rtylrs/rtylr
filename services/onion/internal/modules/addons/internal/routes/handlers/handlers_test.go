package handlers

import (
	"context"
	"testing"
)

func TestEnableAddonHandler_Validate_MissingID(t *testing.T) {
	h := &EnableAddonHandler{}
	req := &EnableAddonRequest{AddonID: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing addon_id")
	}
}

func TestEnableAddonHandler_Validate_Valid(t *testing.T) {
	h := &EnableAddonHandler{}
	req := &EnableAddonRequest{AddonID: "payroll"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDisableAddonHandler_Validate_MissingID(t *testing.T) {
	h := &DisableAddonHandler{}
	req := &DisableAddonRequest{AddonID: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing addon_id")
	}
}

func TestDisableAddonHandler_Validate_Valid(t *testing.T) {
	h := &DisableAddonHandler{}
	req := &DisableAddonRequest{AddonID: "payroll"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
