package domain

import "fmt"

// ValidateOrderInput returns an error if the order input is structurally invalid.
func ValidateOrderInput(input OrderInput) error {
	if len(input.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	for i, item := range input.Items {
		if item.Name == "" {
			return fmt.Errorf("item[%d]: name is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item[%d]: quantity must be positive", i)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("item[%d]: unit price cannot be negative", i)
		}
	}
	return nil
}

// ValidateRefundInput returns an error if the refund input is invalid.
func ValidateRefundInput(input RefundInput) error {
	if input.OrderID == 0 {
		return fmt.Errorf("order id is required")
	}
	if input.Amount <= 0 {
		return fmt.Errorf("refund amount must be positive")
	}
	return nil
}

// ValidateDiscountInput returns an error if the discount input is invalid.
func ValidateDiscountInput(input DiscountInput) error {
	if input.OrderID == 0 {
		return fmt.Errorf("order id is required")
	}
	if input.DiscountValue < 0 {
		return fmt.Errorf("discount value cannot be negative")
	}
	if input.DiscountType == "" {
		return fmt.Errorf("discount type is required")
	}
	if input.DiscountType != "fixed" && input.DiscountType != "percentage" && input.DiscountType != "percent" && input.DiscountType != "pct" {
		return fmt.Errorf("discount type must be fixed or percentage")
	}
	return nil
}
