package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

// --- CreateEquipmentRequestHandler ---

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestCreateEquipmentRequestHandler_Validate_MissingEmployeeID(t *testing.T) {
	h := &CreateEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &CreateEquipmentRequestRequest{
		ItemName: "Laptop",
		Quantity: 1,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400, got %v", se)
	}
}

func TestCreateEquipmentRequestHandler_Validate_MissingItemName(t *testing.T) {
	h := &CreateEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &CreateEquipmentRequestRequest{
		EmployeeID: 1,
		Quantity:   1,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing item_name, got %v", se)
	}
}

func TestCreateEquipmentRequestHandler_Validate_WhitespaceItemName(t *testing.T) {
	h := &CreateEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &CreateEquipmentRequestRequest{
		EmployeeID: 1,
		ItemName:   "   ",
		Quantity:   1,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for whitespace item_name, got %v", se)
	}
}

func TestCreateEquipmentRequestHandler_Validate_ZeroQuantity(t *testing.T) {
	h := &CreateEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &CreateEquipmentRequestRequest{
		EmployeeID: 1,
		ItemName:   "Monitor",
		Quantity:   0,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for zero quantity, got %v", se)
	}
}

func TestCreateEquipmentRequestHandler_Validate_NegativeQuantity(t *testing.T) {
	h := &CreateEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &CreateEquipmentRequestRequest{
		EmployeeID: 1,
		ItemName:   "Monitor",
		Quantity:   -1,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for negative quantity, got %v", se)
	}
}

func TestCreateEquipmentRequestHandler_Validate_ValidRequest(t *testing.T) {
	h := &CreateEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &CreateEquipmentRequestRequest{
		EmployeeID: 1,
		ItemName:   "Laptop",
		Quantity:   2,
	})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

// --- CreateAssetHandler ---

func TestCreateAssetHandler_Validate_MissingName(t *testing.T) {
	h := &CreateAssetHandler{}
	se := h.Validate(context.Background(), &CreateAssetRequest{
		Category:                "Electronics",
		PurchasePriceMinorUnits: 0,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %v", se)
	}
}

func TestCreateAssetHandler_Validate_MissingCategory(t *testing.T) {
	h := &CreateAssetHandler{}
	se := h.Validate(context.Background(), &CreateAssetRequest{
		Name: "MacBook Pro",
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing category, got %v", se)
	}
}

func TestCreateAssetHandler_Validate_NegativePurchasePrice(t *testing.T) {
	h := &CreateAssetHandler{}
	se := h.Validate(context.Background(), &CreateAssetRequest{
		Name:                    "MacBook Pro",
		Category:                "Electronics",
		PurchasePriceMinorUnits: -10000,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for negative purchase_price, got %v", se)
	}
}

func TestCreateAssetHandler_Validate_ValidRequest(t *testing.T) {
	h := &CreateAssetHandler{}
	se := h.Validate(context.Background(), &CreateAssetRequest{
		Name:                    "MacBook Pro",
		Category:                "Electronics",
		PurchasePriceMinorUnits: 250000,
	})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

func TestCreateAssetHandler_Validate_ZeroPriceIsValid(t *testing.T) {
	h := &CreateAssetHandler{}
	se := h.Validate(context.Background(), &CreateAssetRequest{
		Name:     "Donated Chair",
		Category: "Furniture",
	})
	if se != nil {
		t.Errorf("zero purchase_price should be valid, got %v", se)
	}
}

func TestListAssetsHandler_RejectsLifecycleStatusAsCondition(t *testing.T) {
	_, err := (&ListAssetsHandler{}).Serve(authenticatedCtx(), &ListAssetsRequest{Condition: "available"})
	if err == nil || oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected lifecycle status to be rejected as a physical condition, got %v", err)
	}
}

// --- ApproveEquipmentRequestHandler ---

func TestApproveEquipmentRequestHandler_Validate_RejectWithoutReason(t *testing.T) {
	h := &ApproveEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &ApproveEquipmentRequestRequest{
		Approve: false,
		Reason:  "",
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for rejection without reason, got %v", se)
	}
}

func TestApproveEquipmentRequestHandler_Validate_ApproveWithoutReason_IsValid(t *testing.T) {
	h := &ApproveEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &ApproveEquipmentRequestRequest{
		Approve: true,
		Reason:  "",
	})
	if se != nil {
		t.Errorf("approve without reason should pass, got %v", se)
	}
}

func TestApproveEquipmentRequestHandler_Validate_RejectWithReason_IsValid(t *testing.T) {
	h := &ApproveEquipmentRequestHandler{}
	se := h.Validate(context.Background(), &ApproveEquipmentRequestRequest{
		Approve: false,
		Reason:  "Budget constraints",
	})
	if se != nil {
		t.Errorf("reject with reason should pass, got %v", se)
	}
}

// --- AssignAssetHandler ---

func TestAssignAssetHandler_Validate_MissingEmployeeID(t *testing.T) {
	h := &AssignAssetHandler{}
	se := h.Validate(context.Background(), &AssignAssetRequest{EmployeeID: 0})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing employee_id, got %v", se)
	}
}

func TestAssignAssetHandler_Validate_ValidRequest(t *testing.T) {
	h := &AssignAssetHandler{}
	se := h.Validate(context.Background(), &AssignAssetRequest{EmployeeID: 5})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

// --- EquipmentAPI.ApproveEquipment ---

func authenticatedCtx() context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 7, BusinessID: 5, EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}})
}

func TestEquipmentAPIApproveEquipmentRejectsMissingBody(t *testing.T) {
	api := &EquipmentAPI{}
	_, err := api.ApproveEquipment(authenticatedCtx(), &hrequipmentv1.ApproveEquipmentRequest{EquipmentRequestId: 7})
	if err == nil || oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected 400 for a missing approval body, got %v", err)
	}
}

func TestEquipmentAPIApproveEquipmentPassesBodyCheck(t *testing.T) {
	api := &EquipmentAPI{}
	_, err := api.ApproveEquipment(authenticatedCtx(), &hrequipmentv1.ApproveEquipmentRequest{
		EquipmentRequestId: 7,
		Body:               &hrequipmentv1.ApproveEquipmentBody{Approved: true},
	})
	if err != nil && oneKitStatus(err) == http.StatusBadRequest {
		t.Fatalf("a body carrying an explicit decision must pass the body check, got %v", err)
	}
}
