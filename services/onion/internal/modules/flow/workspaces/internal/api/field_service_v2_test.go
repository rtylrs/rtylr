package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// The portal names the field in the message it shows when a required field is
// missing, so the name has to survive the mapping.
func TestFieldErrorNamesTheField(t *testing.T) {
	mapped := fieldError(&domain.FieldValueError{Code: domain.FieldErrRequired, FieldID: 4, FieldName: "Severity"}, "fallback")
	if got := mappedStatus(t, mapped); got != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", got)
	}
	if !strings.Contains(mapped.Error(), "Severity is required") {
		t.Errorf("message = %q, want it to name the field", mapped.Error())
	}
}

func TestFieldErrorStatusCodes(t *testing.T) {
	cases := map[error]int{
		domain.ErrFieldNameTaken:     http.StatusConflict,
		domain.ErrFieldLimit:         http.StatusConflict,
		domain.ErrFieldNotFound:      http.StatusNotFound,
		domain.ErrUnauthorized:       http.StatusForbidden,
		domain.ErrFieldOrderMismatch: http.StatusBadRequest,
		domain.ErrFieldFilterInvalid: http.StatusBadRequest,
		errors.New("boom"):           http.StatusInternalServerError,
	}
	for err, want := range cases {
		if got := mappedStatus(t, fieldError(err, "fallback")); got != want {
			t.Errorf("%v: status = %d, want %d", err, got, want)
		}
	}
}

func TestSetTaskFieldValuesRequestDecodesStringIDs(t *testing.T) {
	var req flowtaskv2.SetTaskFieldValuesRequest
	if err := json.Unmarshal([]byte(`{"task_id":"77","values":[{"field_id":"5","value":"a1"},{"field_id":"6","values":["1","2"]}]}`), &req); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	in := fieldValueInputs(req.GetValues())
	if req.GetTaskId() != 77 || len(in) != 2 || in[0].FieldID != 5 || in[0].Value.Value != "a1" || len(in[1].Value.Values) != 2 {
		t.Fatalf("decoded %d %+v", req.GetTaskId(), in)
	}
}

func TestTaskFieldValuesEncodeKeyedByFieldID(t *testing.T) {
	out := toTaskV2(domain.Task{ID: 1, FieldValues: map[int64]domain.FieldValue{
		9: {Value: "2026-11-19"}, 10: {Values: []string{"3"}},
	}})
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(b), `"field_values":{`) || !strings.Contains(string(b), `"9":{"value":"2026-11-19"}`) {
		t.Fatalf("wire = %s", b)
	}
}

func TestParseFieldFilterAcceptsStringAndNumberIDs(t *testing.T) {
	clauses, err := parseFieldFilter(`[{"field_id":"12","op":"any","values":["a"]},{"field_id":3,"op":"eq","values":["x"]}]`)
	if err != nil || len(clauses) != 2 || clauses[0].FieldID != 12 || clauses[1].FieldID != 3 || clauses[0].Op != "any" {
		t.Fatalf("clauses = %+v, %v", clauses, err)
	}
	for _, bad := range []string{`{"field_id":1}`, `[{"field_id":"x","op":"eq"}]`, `[{"field_id":0,"op":"eq"}]`} {
		if _, err := parseFieldFilter(bad); !errors.Is(err, domain.ErrFieldFilterInvalid) {
			t.Errorf("%s: err = %v", bad, err)
		}
	}
	if clauses, err := parseFieldFilter("  "); err != nil || clauses != nil {
		t.Errorf("empty filter = %v, %v", clauses, err)
	}
}
