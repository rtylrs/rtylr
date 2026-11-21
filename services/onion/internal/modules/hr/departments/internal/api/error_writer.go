package api

import (
	"encoding/json"
	"net/http"

	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
)

func WriteError(w http.ResponseWriter, _ *http.Request, e *hrdepartmentv1.ServerError) {
	status, code, message, _, violations, _ := e.Parts()
	body := map[string]any{"message": message, "code": code}
	if len(violations) > 0 {
		body["violations"] = violations
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
