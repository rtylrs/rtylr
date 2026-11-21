package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowfieldv2 "github.com/voxire/voxie/gen/go/flow/field/v2"
)

// FieldAPIV2 serves a project's custom field definitions (T-580).
type FieldAPIV2 struct{ Service *application.Service }

func NewFieldAPIV2(s *application.Service) *FieldAPIV2 { return &FieldAPIV2{Service: s} }

// publishFieldEvent tells a project's subscribers its field list changed, on
// the same topic the status set uses.
func publishFieldEvent(eventType string, projectID, fieldID, actorID int64) {
	if projectID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("project:%d", projectID),
		Type:     eventType,
		EntityID: fieldID,
		ActorID:  actorID,
	})
}

func toFieldPayload(f domain.CustomField) *flowfieldv2.Field {
	options := make([]*flowfieldv2.FieldOption, 0, len(f.Options))
	for _, o := range f.Options {
		options = append(options, &flowfieldv2.FieldOption{Id: o.ID, Label: o.Label, Color: o.Color})
	}
	return &flowfieldv2.Field{
		Id:            f.ID,
		ProjectId:     f.ProjectID,
		Name:          f.Name,
		Type:          f.Type,
		Options:       options,
		Required:      f.Required,
		DefaultValue:  f.Default.Value,
		DefaultValues: f.Default.Values,
		SortOrder:     f.SortOrder,
		ValueCount:    f.ValueCount,
		CreatedAt:     f.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     f.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func fieldsToPayload(fields []domain.CustomField) []*flowfieldv2.Field {
	out := make([]*flowfieldv2.Field, 0, len(fields))
	for _, f := range fields {
		out = append(out, toFieldPayload(f))
	}
	return out
}

func fieldOptionsFromPayload(in []*flowfieldv2.FieldOption) []domain.FieldOption {
	out := make([]domain.FieldOption, 0, len(in))
	for _, o := range in {
		out = append(out, domain.FieldOption{ID: o.GetId(), Label: o.GetLabel(), Color: o.GetColor()})
	}
	return out
}

// fieldValueMessages is the reason a value was refused, by FieldValueError code.
var fieldValueMessages = map[string]string{
	domain.FieldErrUnknownField: "is not a field of this task's project",
	domain.FieldErrRequired:     "is required",
	domain.FieldErrShape:        "was sent in the wrong shape for its type",
	domain.FieldErrTextTooLong:  "is longer than 2000 characters",
	domain.FieldErrNumber:       "must be a number with at most 15 digits and 6 decimals",
	domain.FieldErrDate:         "must be a date like 2026-11-19",
	domain.FieldErrOption:       "must be one of the field's options",
	domain.FieldErrPerson:       "must be people who can see this project",
	domain.FieldErrURL:          "must be an http or https link",
	domain.FieldErrCheckbox:     "must be true or false",
	domain.FieldErrTooMany:      "has too many values",
}

// fieldError maps the custom field errors that are the caller's fault onto a
// status code and a message. Messages are written out here rather than taken
// from the error, which this package's raw-error guard forbids.
func fieldError(err error, fallback string) error {
	var valueErr *domain.FieldValueError
	if errors.As(err, &valueErr) {
		reason, ok := fieldValueMessages[valueErr.Code]
		if !ok {
			reason = "is invalid"
		}
		name := valueErr.FieldName
		if name == "" {
			name = fmt.Sprintf("Field %d", valueErr.FieldID)
		}
		return onekiterrors.New(http.StatusBadRequest, name+" "+reason)
	}
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrFieldNotFound):
		return onekiterrors.New(http.StatusNotFound, "Field not found")
	case errors.Is(err, domain.ErrFieldNameTaken):
		return onekiterrors.New(http.StatusConflict, "A field with that name already exists in this project")
	case errors.Is(err, domain.ErrFieldLimit):
		return onekiterrors.New(http.StatusConflict, "A project can have at most 50 custom fields")
	case errors.Is(err, domain.ErrFieldNameRequired):
		return onekiterrors.New(http.StatusBadRequest, "Field name is required")
	case errors.Is(err, domain.ErrFieldNameTooLong):
		return onekiterrors.New(http.StatusBadRequest, "Field name can be at most 64 characters")
	case errors.Is(err, domain.ErrInvalidFieldType):
		return onekiterrors.New(http.StatusBadRequest, "type must be one of text, number, date, select, multi_select, person, url, checkbox")
	case errors.Is(err, domain.ErrFieldOptionsRequired):
		return onekiterrors.New(http.StatusBadRequest, "A select field needs at least one option")
	case errors.Is(err, domain.ErrFieldOptionsNotAllowed):
		return onekiterrors.New(http.StatusBadRequest, "Only select and multi_select fields have options")
	case errors.Is(err, domain.ErrFieldOptionLimit):
		return onekiterrors.New(http.StatusBadRequest, "A field can have at most 100 options")
	case errors.Is(err, domain.ErrFieldOptionLabel):
		return onekiterrors.New(http.StatusBadRequest, "Every option needs a label of at most 64 characters")
	case errors.Is(err, domain.ErrFieldOptionDuplicate):
		return onekiterrors.New(http.StatusBadRequest, "Option labels must be unique within a field")
	case errors.Is(err, domain.ErrFieldOptionUnknown):
		return onekiterrors.New(http.StatusBadRequest, "An option id does not belong to this field. Reload the field and try again.")
	case errors.Is(err, domain.ErrFieldOrderMismatch):
		return onekiterrors.New(http.StatusBadRequest, "The new order must list every field in this project exactly once. Reload and try again.")
	case errors.Is(err, domain.ErrTooManyFieldValues):
		return onekiterrors.New(http.StatusBadRequest, "At most 50 field values can be written at once")
	case errors.Is(err, domain.ErrFieldFilterInvalid):
		return onekiterrors.New(http.StatusBadRequest, "Invalid field_filter")
	case errors.Is(err, domain.ErrFieldSortInvalid):
		return onekiterrors.New(http.StatusBadRequest, "Tasks cannot be sorted by that field")
	}
	return onekiterrors.New(http.StatusInternalServerError, fallback)
}

func (a *FieldAPIV2) ListFields(ctx context.Context, req *flowfieldv2.ListFieldsRequest) (*flowfieldv2.ListFieldsResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	fields, err := a.Service.ListProjectFields(ctx, req.GetProjectId(), caller.UserID, businessID)
	if err != nil {
		return nil, fieldError(err, "Failed to list fields")
	}
	return &flowfieldv2.ListFieldsResponse{Fields: fieldsToPayload(fields)}, nil
}

func (a *FieldAPIV2) CreateField(ctx context.Context, req *flowfieldv2.CreateFieldRequest) (*flowfieldv2.CreateFieldResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	field, err := a.Service.CreateProjectField(ctx, caller.UserID, businessID, domain.CustomFieldInput{
		ProjectID: req.GetProjectId(),
		Name:      req.GetName(),
		Type:      req.GetType(),
		Options:   fieldOptionsFromPayload(req.GetOptions()),
		Required:  req.GetRequired(),
		Default:   domain.FieldValue{Value: req.GetDefaultValue(), Values: req.GetDefaultValues()},
	})
	if err != nil {
		return nil, fieldError(err, "Failed to create field")
	}
	publishFieldEvent("field.created", field.ProjectID, field.ID, caller.UserID)
	return &flowfieldv2.CreateFieldResponse{Field: toFieldPayload(field)}, nil
}

func (a *FieldAPIV2) UpdateField(ctx context.Context, req *flowfieldv2.UpdateFieldRequest) (*flowfieldv2.UpdateFieldResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "body is required")
	}
	field, err := a.Service.UpdateProjectField(ctx, req.GetId(), caller.UserID, businessID, domain.CustomFieldInput{
		Name:     body.GetName(),
		Options:  fieldOptionsFromPayload(body.GetOptions()),
		Required: body.GetRequired(),
		Default:  domain.FieldValue{Value: body.GetDefaultValue(), Values: body.GetDefaultValues()},
	})
	if err != nil {
		return nil, fieldError(err, "Failed to update field")
	}
	publishFieldEvent("field.updated", field.ProjectID, field.ID, caller.UserID)
	return &flowfieldv2.UpdateFieldResponse{Field: toFieldPayload(field)}, nil
}

func (a *FieldAPIV2) ReorderFields(ctx context.Context, req *flowfieldv2.ReorderFieldsRequest) (*flowfieldv2.ReorderFieldsResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	fields, err := a.Service.ReorderProjectFields(ctx, req.GetProjectId(), caller.UserID, businessID, req.GetFieldIds())
	if err != nil {
		return nil, fieldError(err, "Failed to reorder fields")
	}
	if len(fields) > 0 {
		publishFieldEvent("field.updated", req.GetProjectId(), fields[0].ID, caller.UserID)
	}
	return &flowfieldv2.ReorderFieldsResponse{Fields: fieldsToPayload(fields)}, nil
}

func (a *FieldAPIV2) DeleteField(ctx context.Context, req *flowfieldv2.DeleteFieldRequest) (*flowfieldv2.DeleteFieldResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	field, cleared, err := a.Service.DeleteProjectField(ctx, req.GetId(), caller.UserID, businessID)
	if err != nil {
		return nil, fieldError(err, "Failed to delete field")
	}
	publishFieldEvent("field.deleted", field.ProjectID, field.ID, caller.UserID)
	return &flowfieldv2.DeleteFieldResponse{
		Message:          "Field deleted",
		ClearedTaskCount: int32(cleared), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}, nil
}
