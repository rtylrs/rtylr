package handlers

import (
	"encoding/json"
	"strconv"
)

// WorkspaceIDField is a workspace_id value as it arrives from the frontend.
// The Flow frontend sends either a numeric DB ID (JSON number) or the
// sentinel string "personal" (meaning the caller's Personal workspace),
// which is resolved server-side. This type accepts both shapes uniformly.
//
// Use `field.String()` to get the raw value, then pass it to
// `Service.ResolveWorkspaceID(ctx, userID, raw)` to get the int64.
type WorkspaceIDField string

// UnmarshalJSON accepts a JSON number, numeric string, or the "personal"
// sentinel. Any other non-empty string is kept verbatim for downstream
// parsing so a useful error can be returned.
func (w *WorkspaceIDField) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*w = ""
		return nil
	}
	// Number form: {"workspace_id": 5}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*w = WorkspaceIDField(strconv.FormatInt(n, 10))
		return nil
	}
	// String form: {"workspace_id": "5"} or {"workspace_id": "personal"}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*w = WorkspaceIDField(s)
	return nil
}

// MarshalJSON emits the value as a JSON string — callers should rarely
// serialise back to JSON, but this keeps round-trips stable.
func (w WorkspaceIDField) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(w))
}

// String returns the raw value. Empty means unset.
func (w WorkspaceIDField) String() string { return string(w) }

// IsEmpty is the canonical "was this field provided?" check.
func (w WorkspaceIDField) IsEmpty() bool { return string(w) == "" }
