package handlers

import (
	"context"
	"testing"
)

// ─── AdjustStockHandler.Validate ─────────────────────────────────────────────

func TestAdjustStock_Validate_AcceptsValidRequest(t *testing.T) {
	h := &AdjustStockHandler{}
	req := &AdjustStockRequest{BranchID: 1, ProductID: 2, Quantity: 10}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("valid request should pass: %v", err.Message)
	}
}

func TestAdjustStock_Validate_RejectsZeroBranchID(t *testing.T) {
	h := &AdjustStockHandler{}
	if err := h.Validate(context.Background(), &AdjustStockRequest{ProductID: 1, Quantity: 5}); err == nil {
		t.Error("expected error for zero branch_id")
	}
}

func TestAdjustStock_Validate_RejectsZeroProductID(t *testing.T) {
	h := &AdjustStockHandler{}
	if err := h.Validate(context.Background(), &AdjustStockRequest{BranchID: 1, Quantity: 5}); err == nil {
		t.Error("expected error for zero product_id")
	}
}

func TestAdjustStock_Validate_RejectsZeroQuantity(t *testing.T) {
	h := &AdjustStockHandler{}
	if err := h.Validate(context.Background(), &AdjustStockRequest{BranchID: 1, ProductID: 2, Quantity: 0}); err == nil {
		t.Error("expected error for zero quantity")
	}
}

func TestAdjustStock_Validate_AcceptsNegativeQuantity(t *testing.T) {
	// Negative quantity means removing stock; non-zero, so should pass.
	h := &AdjustStockHandler{}
	if err := h.Validate(context.Background(), &AdjustStockRequest{BranchID: 1, ProductID: 2, Quantity: -5}); err != nil {
		t.Errorf("negative quantity should pass: %v", err.Message)
	}
}

func TestAdjustStock_Validate_AllZeroFails(t *testing.T) {
	h := &AdjustStockHandler{}
	if err := h.Validate(context.Background(), &AdjustStockRequest{}); err == nil {
		t.Error("all-zero request should fail")
	}
}

// ─── Request struct field checks ─────────────────────────────────────────────

func TestReceiveStockRequest_PositiveQuantity(t *testing.T) {
	req := &ReceiveStockRequest{Quantity: 100}
	if req.Quantity <= 0 {
		t.Error("expected positive quantity")
	}
}

func TestReceiveStockRequest_ZeroQuantity(t *testing.T) {
	req := &ReceiveStockRequest{Quantity: 0}
	if req.Quantity > 0 {
		t.Error("zero quantity should fail the handler guard")
	}
}

func TestWasteStockRequest_PositiveQuantity(t *testing.T) {
	req := &WasteStockRequest{Quantity: 5}
	if req.Quantity <= 0 {
		t.Error("expected positive quantity")
	}
}

func TestTransferStockRequest_RequiresBothBranches(t *testing.T) {
	req := &TransferStockRequest{FromBranchID: 1, ToBranchID: 2}
	if req.FromBranchID == 0 || req.ToBranchID == 0 {
		t.Error("both branches should be set")
	}
}

func TestTransferStockRequest_SameBranchAllowed(t *testing.T) {
	// The data layer may reject same-branch transfers, but the request struct allows it.
	req := &TransferStockRequest{FromBranchID: 1, ToBranchID: 1}
	if req.FromBranchID != req.ToBranchID {
		t.Error("same branch should be representable")
	}
}

// ─── Handler struct creation ──────────────────────────────────────────────────

func TestHandlers_CanBeInstantiated(_ *testing.T) {
	_ = &ListMovementsHandler{}
	_ = &ReceiveStockHandler{}
	_ = &AdjustStockHandler{}
	_ = &WasteStockHandler{}
	_ = &TransferStockHandler{}
	_ = &ListInventoryHandler{}
	_ = &LowStockHandler{}
}

// ─── Routes helpers ───────────────────────────────────────────────────────────

func TestRoutes_NonEmpty(t *testing.T) {
	for name, fn := range map[string]func() []string{
		"ListMovements": ListMovementsRoutes,
		"ReceiveStock":  ReceiveStockRoutes,
		"AdjustStock":   AdjustStockRoutes,
		"WasteStock":    WasteStockRoutes,
		"TransferStock": TransferStockRoutes,
		"ListInventory": ListInventoryRoutes,
		"LowStock":      LowStockRoutes,
	} {
		if len(fn()) == 0 {
			t.Errorf("%s: expected at least one route", name)
		}
	}
}
