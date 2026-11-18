package api

import (
	"encoding/json"
	"net/http"

	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
)

// WriteError keeps the generated `{message, violations}` body and adds the
// error's stable code, so the client can tell a disabled directory from a
// missing permission without matching on message text.
func WriteError(w http.ResponseWriter, _ *http.Request, e *hrstaffv1.ServerError) {
	status, code, message, _, violations, _ := e.Parts()
	body := map[string]any{"message": message, "code": code}
	if len(violations) > 0 {
		body["violations"] = violations
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
