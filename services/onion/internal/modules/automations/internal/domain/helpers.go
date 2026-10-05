package domain

const (
	TriggerOrderCreated    = "order_created"
	TriggerCustomerSignup  = "customer_signup"
	TriggerLoyaltyTierUp   = "loyalty_tier_up"
	TriggerLoyaltyTierDown = "loyalty_tier_down"
	TriggerDateBirthday    = "date_birthday"
	TriggerInactivity      = "inactivity"
	TriggerTagAdded        = "tag_added"
	TriggerPointsThreshold = "points_threshold"

	ActionSendEmail = "send_email"
	ActionSendSMS   = "send_sms"
	ActionAwardPts  = "award_points"
	ActionTagAdd    = "tag_add"
)

// These identifiers are the storage/API contract. An executor must dispatch
// trigger_type using this set and each actions[i].type using the action set;
// presentation labels and generic values are deliberately not accepted.
var supportedTriggerTypes = map[string]struct{}{
	TriggerOrderCreated:    {},
	TriggerCustomerSignup:  {},
	TriggerLoyaltyTierUp:   {},
	TriggerLoyaltyTierDown: {},
	TriggerDateBirthday:    {},
	TriggerInactivity:      {},
	TriggerTagAdded:        {},
	TriggerPointsThreshold: {},
}

var supportedActionTypes = map[string]struct{}{
	ActionSendEmail: {},
	ActionSendSMS:   {},
	ActionAwardPts:  {},
	ActionTagAdd:    {},
}

func IsSupportedTriggerType(triggerType string) bool {
	_, ok := supportedTriggerTypes[triggerType]
	return ok
}

func IsSupportedActionType(actionType string) bool {
	_, ok := supportedActionTypes[actionType]
	return ok
}
