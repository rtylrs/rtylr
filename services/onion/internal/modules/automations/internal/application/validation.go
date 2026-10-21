package application

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
)

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
		if !ok || json.Unmarshal(typeValue, &actionType) != nil || !domain.IsSupportedActionType(actionType) {
			return nil, fmt.Errorf("actions[%d].type is unsupported", i)
		}
		normalized = append(normalized, action)
	}

	return json.Marshal(normalized)
}

// ValidateAutomationDefinition validates the complete persisted contract at
// the HTTP boundary before an automation can be written.
func ValidateAutomationDefinition(triggerType string, actions []byte) error {
	if !domain.IsSupportedTriggerType(triggerType) {
		return fmt.Errorf("trigger_type %q is unsupported", triggerType)
	}
	_, err := NormalizeAutomationActions(actions)
	return err
}
