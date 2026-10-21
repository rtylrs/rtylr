package domain

import "fmt"

// ValidateEarnPointsInput returns an error when the earn input is malformed.
func ValidateEarnPointsInput(input EarnPointsInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.CustomerID == 0 {
		return fmt.Errorf("customer id is required")
	}
	if input.Points <= 0 {
		return fmt.Errorf("points must be positive")
	}
	return nil
}

// ValidateRedeemPointsInput returns an error when the redeem input is malformed.
func ValidateRedeemPointsInput(input RedeemPointsInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.CustomerID == 0 {
		return fmt.Errorf("customer id is required")
	}
	if input.Points <= 0 {
		return fmt.Errorf("points to redeem must be positive")
	}
	return nil
}
