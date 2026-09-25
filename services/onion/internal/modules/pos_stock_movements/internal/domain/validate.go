package domain

import "fmt"

// ValidateReceiveStockInput returns an error if the input is malformed.
func ValidateReceiveStockInput(input ReceiveStockInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.ProductID == 0 {
		return fmt.Errorf("product id is required")
	}
	if input.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return nil
}

// ValidateAdjustStockInput returns an error if the input is malformed.
func ValidateAdjustStockInput(input AdjustStockInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.ProductID == 0 {
		return fmt.Errorf("product id is required")
	}
	return nil
}

// ValidateWasteStockInput returns an error if the input is malformed.
func ValidateWasteStockInput(input WasteStockInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.ProductID == 0 {
		return fmt.Errorf("product id is required")
	}
	if input.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return nil
}

// ValidateTransferStockInput returns an error if the input is malformed.
func ValidateTransferStockInput(input TransferStockInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.FromBranchID == 0 {
		return fmt.Errorf("from branch id is required")
	}
	if input.ToBranchID == 0 {
		return fmt.Errorf("to branch id is required")
	}
	if input.FromBranchID == input.ToBranchID {
		return fmt.Errorf("source and destination branches must differ")
	}
	if input.ProductID == 0 {
		return fmt.Errorf("product id is required")
	}
	if input.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return nil
}
