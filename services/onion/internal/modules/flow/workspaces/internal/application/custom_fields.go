package application

import (
	"context"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Custom fields (T-580).
//
// Definitions take the admin tier, like the status set: adding or removing a
// field reshapes every task in the project. Values take the member tier, like
// every other task edit.

type customFieldStore interface {
	ListCustomFields(ctx context.Context, businessID, projectID int64) ([]domain.CustomField, error)
	GetCustomField(ctx context.Context, id, businessID int64) (domain.CustomField, error)
	CreateCustomField(ctx context.Context, businessID, createdBy int64, in domain.CustomFieldInput) (domain.CustomField, error)
	UpdateCustomField(ctx context.Context, existing domain.CustomField, in domain.CustomFieldInput, removedOptionIDs []string) (domain.CustomField, error)
	ReorderCustomFields(ctx context.Context, businessID, projectID int64, ids []int64) ([]domain.CustomField, error)
	DeleteCustomField(ctx context.Context, id, businessID int64) (int64, error)
	SetTaskFieldValues(ctx context.Context, businessID, taskID int64, writes []domain.FieldValueWrite) error
	ListTaskFieldValues(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]map[int64]domain.FieldValue, error)
	CountFieldValues(ctx context.Context, businessID, projectID int64) (map[int64]int32, error)
}

// requireFieldProject is RequireProjectRole plus the check it does not make:
// that the project belongs to the caller's business. A business administrator
// passes the role check for any project id.
func (s *Service) requireFieldProject(ctx context.Context, projectID, callerID, businessID int64, minRole string) error {
	p, err := s.repo.GetProjectProxyFacts(ctx, projectID)
	if err != nil || p.BusinessID != businessID {
		return domain.ErrUnauthorized
	}
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, minRole); err != nil {
		return domain.ErrUnauthorized
	}
	return nil
}

func (s *Service) ListProjectFields(ctx context.Context, projectID, callerID, businessID int64) ([]domain.CustomField, error) {
	if err := s.requireFieldProject(ctx, projectID, callerID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, err
	}
	fields, err := s.repo.ListCustomFields(ctx, businessID, projectID)
	if err != nil || len(fields) == 0 {
		return fields, err
	}
	// The count lets the settings screen say how many tasks a delete clears.
	counts, err := s.repo.CountFieldValues(ctx, businessID, projectID)
	if err != nil {
		return nil, err
	}
	for i := range fields {
		fields[i].ValueCount = counts[fields[i].ID]
	}
	return fields, nil
}

func (s *Service) CreateProjectField(ctx context.Context, callerID, businessID int64, in domain.CustomFieldInput) (domain.CustomField, error) {
	if err := s.requireFieldProject(ctx, in.ProjectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.CustomField{}, err
	}
	normalized, _, err := domain.NormalizeFieldInput(in, nil)
	if err != nil {
		return domain.CustomField{}, err
	}
	if err := s.checkFieldPeople(ctx, domain.CustomField{ProjectID: in.ProjectID, Name: normalized.Name, Type: normalized.Type}, businessID, normalized.Default); err != nil {
		return domain.CustomField{}, err
	}
	return s.repo.CreateCustomField(ctx, businessID, callerID, normalized)
}

// UpdateProjectField replaces a field. The field is read first so its own
// project is role-checked rather than one named by the request.
func (s *Service) UpdateProjectField(ctx context.Context, id, callerID, businessID int64, in domain.CustomFieldInput) (domain.CustomField, error) {
	existing, err := s.repo.GetCustomField(ctx, id, businessID)
	if err != nil {
		return domain.CustomField{}, err
	}
	if err := s.requireFieldProject(ctx, existing.ProjectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.CustomField{}, err
	}
	in.ProjectID, in.Type = existing.ProjectID, existing.Type
	normalized, removed, err := domain.NormalizeFieldInput(in, existing.Options)
	if err != nil {
		return domain.CustomField{}, err
	}
	if err := s.checkFieldPeople(ctx, existing, businessID, normalized.Default); err != nil {
		return domain.CustomField{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.UpdateCustomField(ctx, existing, normalized, removed)
}

func (s *Service) ReorderProjectFields(ctx context.Context, projectID, callerID, businessID int64, ids []int64) ([]domain.CustomField, error) {
	if err := s.requireFieldProject(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, err
	}
	return s.repo.ReorderCustomFields(ctx, businessID, projectID, ids)
}

// DeleteProjectField removes a field and its values, returning the deleted
// field and how many tasks had a value.
func (s *Service) DeleteProjectField(ctx context.Context, id, callerID, businessID int64) (domain.CustomField, int64, error) {
	existing, err := s.repo.GetCustomField(ctx, id, businessID)
	if err != nil {
		return domain.CustomField{}, 0, err
	}
	if err := s.requireFieldProject(ctx, existing.ProjectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.CustomField{}, 0, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	cleared, err := s.repo.DeleteCustomField(ctx, id, businessID)
	return existing, cleared, err
}

// checkFieldPeople confirms every user in a person value belongs to the
// business and can see the field's project, so a task never names someone who
// cannot open it.
func (s *Service) checkFieldPeople(ctx context.Context, field domain.CustomField, businessID int64, v domain.FieldValue) error {
	if field.Type != domain.FieldTypePerson {
		return nil
	}
	for _, raw := range v.Values {
		userID, err := strconv.ParseInt(raw, 10, 64)
		allowed := err == nil
		if allowed && s.hubProfiles != nil {
			_, profileErr := s.hubProfiles.GetProfile(ctx, userID, businessID)
			allowed = profileErr == nil
		}
		if !allowed || !s.opensProject(ctx, field.ProjectID, userID, businessID) {
			return &domain.FieldValueError{Code: domain.FieldErrPerson, FieldID: field.ID, FieldName: field.Name}
		}
	}
	return nil
}

func fieldsByID(fields []domain.CustomField) map[int64]domain.CustomField {
	out := make(map[int64]domain.CustomField, len(fields))
	for _, f := range fields {
		out[f.ID] = f
	}
	return out
}

// SetTaskFieldValues writes or clears several values of one task. A field id
// that is not one of the task's own project is refused, which is what stops a
// field of project A landing on a task of project B. Clearing a required field
// is refused; a required field the task never had may stay empty.
func (s *Service) SetTaskFieldValues(ctx context.Context, taskID, callerID, businessID int64, inputs []domain.FieldValueInput) (domain.Task, error) {
	if len(inputs) > domain.MaxFieldValuesPerCall {
		return domain.Task{}, domain.ErrTooManyFieldValues
	}
	if err := s.RequireTaskRole(ctx, taskID, callerID, businessID, domain.FlowRoleMember); err != nil {
		return domain.Task{}, domain.ErrUnauthorized
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil || task.BusinessID != businessID {
		return domain.Task{}, domain.ErrUnauthorized
	}
	fields, err := s.repo.ListCustomFields(ctx, businessID, task.ProjectID)
	if err != nil {
		return domain.Task{}, err
	}
	byID := fieldsByID(fields)
	writes := make([]domain.FieldValueWrite, 0, len(inputs))
	for _, in := range inputs {
		field, ok := byID[in.FieldID]
		if !ok {
			return domain.Task{}, &domain.FieldValueError{Code: domain.FieldErrUnknownField, FieldID: in.FieldID}
		}
		value, valueErr := domain.NormalizeFieldValue(field, in.Value)
		if valueErr != nil {
			return domain.Task{}, valueErr
		}
		if value.IsEmpty() && field.Required {
			return domain.Task{}, &domain.FieldValueError{Code: domain.FieldErrRequired, FieldID: field.ID, FieldName: field.Name}
		}
		if peopleErr := s.checkFieldPeople(ctx, field, businessID, value); peopleErr != nil {
			return domain.Task{}, peopleErr
		}
		writes = append(writes, domain.FieldValueWrite{Field: field, Value: value})
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	if err := s.repo.SetTaskFieldValues(ctx, businessID, taskID, writes); err != nil {
		return domain.Task{}, err
	}
	tasks := []domain.Task{task}
	if err := s.attachFieldValues(ctx, businessID, tasks); err != nil {
		return domain.Task{}, err
	}
	return tasks[0], nil
}

// fieldValuesForNewTask resolves what a new task in projectID stores: the
// values sent, the default of every field left out, and an error when a
// required field ends up with neither.
func (s *Service) fieldValuesForNewTask(ctx context.Context, businessID, projectID int64, inputs []domain.FieldValueInput) ([]domain.FieldValueWrite, error) {
	if len(inputs) > domain.MaxFieldValuesPerCall {
		return nil, domain.ErrTooManyFieldValues
	}
	fields, err := s.repo.ListCustomFields(ctx, businessID, projectID)
	if err != nil {
		return nil, err
	}
	byID := fieldsByID(fields)
	sent := make(map[int64]domain.FieldValue, len(inputs))
	for _, in := range inputs {
		field, ok := byID[in.FieldID]
		if !ok {
			return nil, &domain.FieldValueError{Code: domain.FieldErrUnknownField, FieldID: in.FieldID}
		}
		value, valueErr := domain.NormalizeFieldValue(field, in.Value)
		if valueErr != nil {
			return nil, valueErr
		}
		if peopleErr := s.checkFieldPeople(ctx, field, businessID, value); peopleErr != nil {
			return nil, peopleErr
		}
		sent[field.ID] = value
	}
	writes := make([]domain.FieldValueWrite, 0, len(fields))
	for _, field := range fields {
		value := sent[field.ID]
		if value.IsEmpty() {
			value = field.Default
		}
		if value.IsEmpty() {
			if field.Required {
				return nil, &domain.FieldValueError{Code: domain.FieldErrRequired, FieldID: field.ID, FieldName: field.Name}
			}
			continue
		}
		writes = append(writes, domain.FieldValueWrite{Field: field, Value: value})
	}
	return writes, nil
}

// storeNewTaskFieldValues writes the resolved values of a task that was just
// inserted. A task that should carry required values must not survive without
// them, so a failed write removes it.
func (s *Service) storeNewTaskFieldValues(ctx context.Context, businessID, taskID int64, writes []domain.FieldValueWrite) error {
	err := s.repo.SetTaskFieldValues(ctx, businessID, taskID, writes)
	if err == nil {
		return nil
	}
	if delErr := s.repo.DeleteTask(ctx, taskID, businessID); delErr != nil {
		logger.Ctx(ctx).Error("CreateTask: failed to remove task after field value error", logger.Error(delErr))
	}
	return err
}

// attachFieldValues fills FieldValues on a page of tasks with one query.
func (s *Service) attachFieldValues(ctx context.Context, businessID int64, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID)
	}
	values, err := s.repo.ListTaskFieldValues(ctx, businessID, ids)
	if err != nil {
		return err
	}
	for i := range tasks {
		tasks[i].FieldValues = values[tasks[i].ID]
	}
	return nil
}

// ResolveTaskFieldQuery validates the custom field part of a task listing: the
// field_filter clauses (decoded by the API) and a "field:<id>" sort. Both need a project, because a
// field belongs to one, and every field id is looked up inside that project.
func (s *Service) ResolveTaskFieldQuery(ctx context.Context, callerID, businessID, projectID int64, clauses []domain.RawFieldFilter, sortBy, sortDir string) ([]domain.TaskFieldFilter, *domain.TaskFieldSort, error) {
	if len(clauses) > domain.MaxFieldFilterClauses {
		return nil, nil, domain.ErrFieldFilterInvalid
	}
	sortFieldID, sorting, err := domain.ParseFieldSort(sortBy)
	if err != nil {
		return nil, nil, err
	}
	if len(clauses) == 0 && !sorting {
		return nil, nil, nil
	}
	if projectID == 0 {
		return nil, nil, domain.ErrFieldFilterInvalid
	}
	if err := s.requireFieldProject(ctx, projectID, callerID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, nil, err
	}
	fields, err := s.repo.ListCustomFields(ctx, businessID, projectID)
	if err != nil {
		return nil, nil, err
	}
	byID := fieldsByID(fields)

	filters := make([]domain.TaskFieldFilter, 0, len(clauses))
	for _, c := range clauses {
		field, ok := byID[c.FieldID]
		if !ok {
			return nil, nil, domain.ErrFieldFilterInvalid
		}
		filter, buildErr := domain.BuildTaskFieldFilter(field, c.Op, c.Values)
		if buildErr != nil {
			return nil, nil, buildErr
		}
		filters = append(filters, filter)
	}
	if !sorting {
		return filters, nil, nil
	}
	field, ok := byID[sortFieldID]
	if !ok || !domain.FieldSortable(field.Type) {
		return nil, nil, domain.ErrFieldSortInvalid
	}
	return filters, &domain.TaskFieldSort{Field: field, Desc: strings.EqualFold(sortDir, "desc")}, nil
}
