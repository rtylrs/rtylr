package echo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The Caret panel reads the decision fields at the top level of the JSON
// body, and lists tools with ?mode=. Both must survive the move to the
// generated contract server.
func TestCaretRuntimeRoutesKeepTheirPathsAndJSONShape(t *testing.T) {
	mux := http.NewServeMux()
	(&Module{}).RegisterAPI(mux)

	tools := httptest.NewRecorder()
	mux.ServeHTTP(tools, httptest.NewRequest(http.MethodGet, "/caret/tools?mode=ask", http.NoBody))
	if tools.Code != http.StatusOK {
		t.Fatalf("GET /caret/tools = %d: %s", tools.Code, tools.Body)
	}
	var list struct {
		Tools []struct {
			Name  string `json:"name"`
			Scope string `json:"scope"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(tools.Body.Bytes(), &list); err != nil || len(list.Tools) == 0 {
		t.Fatalf("tool list not decodable: %v %s", err, tools.Body)
	}
	for _, tool := range list.Tools {
		if tool.Scope != "read" {
			t.Fatalf("ask mode listed write tool %s", tool.Name)
		}
	}

	body := `{"tool_call_id":"t1","name":"crm.create_customer","input":{"name":"Rana"},"decision":"deny","mode":"agent"}`
	req := httptest.NewRequest(http.MethodPost, "/caret/tools/decision", strings.NewReader(body)).WithContext(context.Background())
	req.Header.Set("Content-Type", "application/json")
	decision := httptest.NewRecorder()
	mux.ServeHTTP(decision, req)
	if decision.Code != http.StatusOK {
		t.Fatalf("POST /caret/tools/decision = %d: %s", decision.Code, decision.Body)
	}
	var result map[string]any
	if err := json.Unmarshal(decision.Body.Bytes(), &result); err != nil {
		t.Fatalf("decision not decodable: %v", err)
	}
	if result["tool_call_id"] != "t1" || result["status"] == nil || result["message"] == nil {
		t.Fatalf("decision fields must be top-level: %s", decision.Body)
	}

	stream := httptest.NewRecorder()
	mux.ServeHTTP(stream, httptest.NewRequest(http.MethodGet, "/ai/generate-text-stream", http.NoBody))
	if stream.Code != http.StatusMethodNotAllowed {
		t.Fatalf("the stream route should be registered for POST only, GET got %d", stream.Code)
	}
}
