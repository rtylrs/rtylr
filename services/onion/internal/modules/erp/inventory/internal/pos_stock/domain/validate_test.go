package domain

import "testing"

func TestValidateReceiveStockInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateReceiveStockInput(ReceiveStockInput{BranchID: 1, ProductID: 2, Quantity: 5}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateReceiveStockInput_RejectsMissingBranchID(t *testing.T) {
	if err := ValidateReceiveStockInput(ReceiveStockInput{BusinessID: 1, ProductID: 2, Quantity: 5}); err == nil {
		t.Fatal("expected error for missing branch id")
	}
}

func TestValidateReceiveStockInput_RejectsMissingProductID(t *testing.T) {
	if err := ValidateReceiveStockInput(ReceiveStockInput{BusinessID: 1, BranchID: 2, Quantity: 5}); err == nil {
		t.Fatal("expected error for missing product id")
	}
}

func TestValidateReceiveStockInput_RejectsZeroQuantity(t *testing.T) {
	if err := ValidateReceiveStockInput(ReceiveStockInput{BusinessID: 1, BranchID: 2, ProductID: 3, Quantity: 0}); err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestValidateReceiveStockInput_AcceptsValid(t *testing.T) {
	if err := ValidateReceiveStockInput(ReceiveStockInput{BusinessID: 1, BranchID: 2, ProductID: 3, Quantity: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTransferStockInput_RejectsSameBranch(t *testing.T) {
	if err := ValidateTransferStockInput(TransferStockInput{
		BusinessID: 1, FromBranchID: 2, ToBranchID: 2, ProductID: 3, Quantity: 5,
	}); err == nil {
		t.Fatal("expected error for same source and destination branch")
	}
}

func TestValidateTransferStockInput_RejectsMissingFromBranch(t *testing.T) {
	if err := ValidateTransferStockInput(TransferStockInput{
		BusinessID: 1, ToBranchID: 2, ProductID: 3, Quantity: 5,
	}); err == nil {
		t.Fatal("expected error for missing from branch")
	}
}

func TestValidateTransferStockInput_RejectsMissingToBranch(t *testing.T) {
	if err := ValidateTransferStockInput(TransferStockInput{
		BusinessID: 1, FromBranchID: 2, ProductID: 3, Quantity: 5,
	}); err == nil {
		t.Fatal("expected error for missing to branch")
	}
}

func TestValidateTransferStockInput_RejectsZeroQuantity(t *testing.T) {
	if err := ValidateTransferStockInput(TransferStockInput{
		BusinessID: 1, FromBranchID: 2, ToBranchID: 3, ProductID: 4, Quantity: 0,
	}); err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestValidateTransferStockInput_AcceptsValid(t *testing.T) {
	if err := ValidateTransferStockInput(TransferStockInput{
		BusinessID: 1, FromBranchID: 2, ToBranchID: 3, ProductID: 4, Quantity: 10,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAdjustStockInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateAdjustStockInput(AdjustStockInput{BranchID: 1, ProductID: 2}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateAdjustStockInput_AcceptsValid(t *testing.T) {
	if err := ValidateAdjustStockInput(AdjustStockInput{BusinessID: 1, BranchID: 2, ProductID: 3}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateWasteStockInput_RejectsZeroQuantity(t *testing.T) {
	if err := ValidateWasteStockInput(WasteStockInput{BusinessID: 1, BranchID: 2, ProductID: 3}); err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestValidateWasteStockInput_AcceptsValid(t *testing.T) {
	if err := ValidateWasteStockInput(WasteStockInput{BusinessID: 1, BranchID: 2, ProductID: 3, Quantity: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
