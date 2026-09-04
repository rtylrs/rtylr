package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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

// NormalizeAutomationActions validates the JSON action list and makes the
// legacy string form explicit objects. Persisted actions always have a
// canonical `type` field, with optional executor-specific fields retained.
func NormalizeAutomationActions(raw []byte) ([]byte, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return []byte("[]"), nil
	}

	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("actions must be a JSON array: %w", err)
	}

	normalized := make([]map[string]json.RawMessage, 0, len(entries))
	for i, entry := range entries {
		entry = bytes.TrimSpace(entry)
		if len(entry) == 0 {
			return nil, fmt.Errorf("actions[%d] must not be empty", i)
		}

		var action map[string]json.RawMessage
		if entry[0] == '"' {
			var actionType string
			if err := json.Unmarshal(entry, &actionType); err != nil {
				return nil, fmt.Errorf("actions[%d] must contain a canonical type: %w", i, err)
			}
			action = map[string]json.RawMessage{}
			typeValue, err := json.Marshal(actionType)
			if err != nil {
				return nil, fmt.Errorf("actions[%d] type could not be encoded: %w", i, err)
			}
			action["type"] = typeValue
		} else if err := json.Unmarshal(entry, &action); err != nil || action == nil {
			return nil, fmt.Errorf("actions[%d] must be an object with a canonical type", i)
		}

		var actionType string
		typeValue, ok := action["type"]
		if !ok || json.Unmarshal(typeValue, &actionType) != nil || !IsSupportedActionType(actionType) {
			return nil, fmt.Errorf("actions[%d].type is unsupported", i)
		}
		normalized = append(normalized, action)
	}

	return json.Marshal(normalized)
}

// ValidateAutomationDefinition validates the complete persisted contract at
// the HTTP boundary before an automation can be written.
func ValidateAutomationDefinition(triggerType string, actions []byte) error {
	if !IsSupportedTriggerType(triggerType) {
		return fmt.Errorf("trigger_type %q is unsupported", triggerType)
	}
	_, err := NormalizeAutomationActions(actions)
	return err
}

// IsActive returns true if the automation is currently enabled.
func (a *Automation) IsEnabled() bool {
	return a.IsActive
}

// HasTrigger returns true if the automation has a non-empty trigger type.
func (a *Automation) HasTrigger() bool {
	return a.TriggerType != ""
}

// HasRuns returns true if the automation has been executed at least once.
func (a *Automation) HasRuns() bool {
	return a.RunCount > 0
}

// HasLastRun returns true if the automation recorded a last run timestamp.
func (a *Automation) HasLastRun() bool {
	return a.LastRunAt != nil
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

// ValidateAutomationUpdateInput checks required fields for updating an automation.
func ValidateAutomationUpdateInput(input AutomationUpdateInput) string {
	if input.ID <= 0 {
		return "automation id is required"
	}
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.Name == "" {
		return "automation name is required"
	}
	if input.TriggerType == "" {
		return "trigger type is required"
	}
	return ""
}
