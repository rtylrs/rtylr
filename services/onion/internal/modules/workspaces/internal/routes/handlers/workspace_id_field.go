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

// ─────────────────────────────────────────────────────────────────────────────
// JSONID — accepts either a JSON number or a numeric string and unmarshals
// into an int64. The Flow frontend stringifies DB IDs for URL params + store
// comparisons, then sends them back in POST bodies as strings. Without this
// the Go JSON decoder rejects `"1"` → `int64` with "Invalid request" before
// Validate runs, so the handler never gets a chance to return a helpful
// message.
//
// Use as the JSON-decoded field type; call `.Int64()` to pass downstream:
//
//	type CreateTaskRequest struct {
//	    ProjectID JSONID `json:"project_id"`
//	}
//	id := req.ProjectID.Int64()
// ─────────────────────────────────────────────────────────────────────────────

type JSONID int64

func (j *JSONID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*j = 0
		return nil
	}
	// Number form.
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*j = JSONID(n)
		return nil
	}
	// String form — parse it.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*j = 0
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*j = JSONID(n)
	return nil
}

func (j JSONID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(j))
}

// Int64 returns the underlying numeric value. Zero means unset.
func (j JSONID) Int64() int64 { return int64(j) }

// IsZero is the canonical "was this field provided?" check.
func (j JSONID) IsZero() bool { return int64(j) == 0 }

// jsonIDPtrToInt64Ptr converts an optional, string-or-number id from a request
// body into the *int64 the domain layer expects. nil (field omitted or null) and
// an explicit 0 both map to nil so the caller's "leave unchanged" semantics hold;
// the frontend sends ids as strings, which a plain *int64 field cannot decode.
func jsonIDPtrToInt64Ptr(j *JSONID) *int64 {
	if j == nil || int64(*j) == 0 {
		return nil
	}
	v := int64(*j)
	return &v
}
