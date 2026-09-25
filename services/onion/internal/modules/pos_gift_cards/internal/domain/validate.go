package domain

import "fmt"

// ValidateGiftCardInput returns an error when the create input is malformed.
func ValidateGiftCardInput(input GiftCardInput) error {
	if input.PaidOrderID <= 0 {
		return fmt.Errorf("paid order id is required")
	}
	if input.Code == "" {
		return fmt.Errorf("gift card code is required")
	}
	if input.InitialAmount <= 0 {
		return fmt.Errorf("initial amount must be positive")
	}
	return nil
}

// ValidateRedeemInput returns an error when the redeem input is malformed.
func ValidateRedeemInput(input RedeemInput) error {
	if input.OrderID <= 0 {
		return fmt.Errorf("order id is required")
	}
	if input.Code == "" {
		return fmt.Errorf("gift card code is required")
	}
	if input.Amount <= 0 {
		return fmt.Errorf("redeem amount must be positive")
	}
	return nil
}
