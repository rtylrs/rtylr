package api

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func fieldValuesToProto(values map[int64]domain.FieldValue) map[string]*flowtaskv2.TaskFieldValue {
	out := make(map[string]*flowtaskv2.TaskFieldValue, len(values))
	for fieldID, v := range values {
		out[strconv.FormatInt(fieldID, 10)] = &flowtaskv2.TaskFieldValue{Value: v.Value, Values: v.Values}
	}
	return out
}

func fieldValueInputs(in []*flowtaskv2.TaskFieldValueInput) []domain.FieldValueInput {
	out := make([]domain.FieldValueInput, 0, len(in))
	for _, v := range in {
		out = append(out, domain.FieldValueInput{
			FieldID: v.GetFieldId(),
			Value:   domain.FieldValue{Value: v.GetValue(), Values: v.GetValues()},
		})
	}
	return out
}

// fieldFilterClause is one field_filter clause on the wire. The field id may
// arrive as a JSON string (how int64 travels) or a number (hand-written).
type fieldFilterClause struct {
	FieldID json.Number `json:"field_id"`
	Op      string      `json:"op"`
	Values  []string    `json:"values"`
}

// parseFieldFilter decodes the listTasks field_filter parameter, a JSON array
// of clauses. Empty means no clauses.
func parseFieldFilter(raw string) ([]domain.RawFieldFilter, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var wire []fieldFilterClause
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&wire); err != nil || len(wire) > domain.MaxFieldFilterClauses {
		return nil, domain.ErrFieldFilterInvalid
	}
	out := make([]domain.RawFieldFilter, 0, len(wire))
	for _, c := range wire {
		id, err := strconv.ParseInt(c.FieldID.String(), 10, 64)
		if err != nil || id <= 0 {
			return nil, domain.ErrFieldFilterInvalid
		}
		out = append(out, domain.RawFieldFilter{FieldID: id, Op: c.Op, Values: c.Values})
	}
	return out, nil
}

// isFieldInputError reports whether err is a custom field problem the caller
// can fix, as opposed to a failure of the task write itself.
func isFieldInputError(err error) bool {
	var valueErr *domain.FieldValueError
	return errors.As(err, &valueErr) || errors.Is(err, domain.ErrTooManyFieldValues)
}

func (a *TaskAPIV2) SetTaskFieldValues(ctx context.Context, req *flowtaskv2.SetTaskFieldValuesRequest) (*flowtaskv2.SetTaskFieldValuesResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	task, err := a.Service.SetTaskFieldValues(ctx, req.GetTaskId(), caller.UserID, businessID, fieldValueInputs(req.GetValues()))
	if err != nil {
		return nil, fieldError(err, "Failed to set field values")
	}
	publishTaskEventByProject("task.updated", task.ProjectID, task.ID, caller.UserID)
	publishBusinessEvent("business.task_updated", businessID, task.ID, caller.UserID)
	return &flowtaskv2.SetTaskFieldValuesResponse{FieldValues: fieldValuesToProto(task.FieldValues)}, nil
}
