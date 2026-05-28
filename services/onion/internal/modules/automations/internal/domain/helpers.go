package domain

const (
	TriggerOrderCreated    = "order_created"
	TriggerCustomerSignup  = "customer_signup"
	TriggerLoyaltyTierUp   = "loyalty_tier_up"
	TriggerLoyaltyTierDown = "loyalty_tier_down"
	TriggerDateBirthday    = "date_birthday"
	TriggerInactivity      = "inactivity"

	ActionSendEmail = "send_email"
	ActionSendSMS   = "send_sms"
	ActionAwardPts  = "award_points"
	ActionTagAdd    = "tag_add"
)

// IsActive returns true if the automation is currently enabled.
func (a *Automation) IsEnabled() bool {
	return a.IsActive
}

// HasTrigger returns true if the automation has a non-empty trigger type.
func (a *Automation) HasTrigger() bool {
	return a.TriggerType != ""
}

// ValidateAutomationInput checks if the required fields are present.
// Returns a non-empty error string if invalid.
func ValidateAutomationInput(input AutomationInput) string {
	if input.Name == "" {
		return "automation name is required"
	}
	if input.TriggerType == "" {
		return "trigger type is required"
	}
	return ""
}
