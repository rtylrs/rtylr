package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidOneOnOne = errors.New("invalid one-on-one")
	ErrStaffNotFound   = errors.New("staff member not found in this business")
)

// NormalizeStatus maps the accepted spellings onto the hr_one_on_ones enum.
func NormalizeStatus(status string) (string, error) {
	switch s := strings.ToLower(strings.TrimSpace(status)); s {
	case "":
		return "scheduled", nil
	case "scheduled", "completed", "cancelled":
		return s, nil
	default:
		return "", fmt.Errorf("%w: unknown status %q", ErrInvalidOneOnOne, status)
	}
}

// NormalizeActionItems returns a JSON array for the action_items JSON column.
// A valid JSON array passes through; any other text is split into one item per
// non-empty line, so plain text from older clients no longer fails the insert.
func NormalizeActionItems(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "[]"
	}
	var arr []json.RawMessage
	if json.Unmarshal([]byte(trimmed), &arr) == nil {
		return trimmed
	}
	items := []string{}
	for _, line := range strings.Split(trimmed, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			items = append(items, line)
		}
	}
	out, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(out)
}
