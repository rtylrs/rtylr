package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Custom fields (T-580). Every statement carries business_id. The queries are
// written by hand because most of them run inside a transaction or are built
// from a variable list of ids.

const customFieldColumns = `id, business_id, project_id, name, type, options, required, default_value,
	sort_order, created_by, created_at, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

// storedOption and storedValue are the JSON shapes of the options and
// default_value columns. The domain types carry no tags, so the column format
// is defined here, next to the SQL that reads and writes it.
type storedOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"`
}

type storedValue struct {
	Value  string   `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

func scanCustomField(row rowScanner) (domain.CustomField, error) {
	var f domain.CustomField
	var options, defaultValue []byte
	if err := row.Scan(&f.ID, &f.BusinessID, &f.ProjectID, &f.Name, &f.Type, &options, &f.Required,
		&defaultValue, &f.SortOrder, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return domain.CustomField{}, err
	}
	f.Options = []domain.FieldOption{}
	if len(options) > 0 {
		var stored []storedOption
		if err := json.Unmarshal(options, &stored); err != nil {
			return domain.CustomField{}, err
		}
		for _, o := range stored {
			f.Options = append(f.Options, domain.FieldOption{ID: o.ID, Label: o.Label, Color: o.Color})
		}
	}
	if len(defaultValue) > 0 {
		var stored storedValue
		if err := json.Unmarshal(defaultValue, &stored); err != nil {
			return domain.CustomField{}, err
		}
		f.Default = domain.FieldValue{Value: stored.Value, Values: stored.Values}
	}
	return f, nil
}

func fieldJSONColumns(options []domain.FieldOption, def domain.FieldValue) (optionsJSON, defaultJSON any, err error) {
	if len(options) > 0 {
		stored := make([]storedOption, 0, len(options))
		for _, o := range options {
			stored = append(stored, storedOption{ID: o.ID, Label: o.Label, Color: o.Color})
		}
		b, marshalErr := json.Marshal(stored)
		if marshalErr != nil {
			return nil, nil, marshalErr
		}
		optionsJSON = string(b)
	}
	if !def.IsEmpty() {
		b, marshalErr := json.Marshal(storedValue{Value: def.Value, Values: def.Values})
		if marshalErr != nil {
			return nil, nil, marshalErr
		}
		defaultJSON = string(b)
	}
	return optionsJSON, defaultJSON, nil
}

func (r *Repository) ListCustomFields(ctx context.Context, businessID, projectID int64) ([]domain.CustomField, error) {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT "+customFieldColumns+" FROM flow_custom_fields WHERE business_id = ? AND project_id = ? ORDER BY sort_order, id",
		businessID, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("ListCustomFields failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []domain.CustomField{}
	for rows.Next() {
		f, scanErr := scanCustomField(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// CountFieldValues returns, per field of a project, how many tasks hold a
// value. A field without values is absent from the map.
func (r *Repository) CountFieldValues(ctx context.Context, businessID, projectID int64) (map[int64]int32, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT v.field_id, COUNT(*) FROM flow_task_field_values v
		JOIN flow_custom_fields f ON f.id = v.field_id AND f.business_id = v.business_id
		WHERE v.business_id = ? AND f.project_id = ? GROUP BY v.field_id`, businessID, projectID)
	if err != nil {
		logger.Ctx(ctx).Error("CountFieldValues failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[int64]int32{}
	for rows.Next() {
		var fieldID int64
		var count int32
		if scanErr := rows.Scan(&fieldID, &count); scanErr != nil {
			return nil, scanErr
		}
		out[fieldID] = count
	}
	return out, rows.Err()
}

func (r *Repository) GetCustomField(ctx context.Context, id, businessID int64) (domain.CustomField, error) {
	f, err := scanCustomField(r.pool.QueryRowContext(ctx,
		"SELECT "+customFieldColumns+" FROM flow_custom_fields WHERE id = ? AND business_id = ?", id, businessID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CustomField{}, domain.ErrFieldNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetCustomField failed", logger.Error(err))
	}
	return f, err
}

// CreateCustomField appends a field to the project. The project row is locked
// first so two concurrent creates cannot both pass the per-project cap.
func (r *Repository) CreateCustomField(ctx context.Context, businessID, createdBy int64, in domain.CustomFieldInput) (domain.CustomField, error) {
	optionsJSON, defaultJSON, err := fieldJSONColumns(in.Options, in.Default)
	if err != nil {
		return domain.CustomField{}, err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.CustomField{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var projectID int64
	if err := tx.QueryRowContext(ctx,
		"SELECT id FROM flow_projects WHERE id = ? AND business_id = ? FOR UPDATE", in.ProjectID, businessID,
	).Scan(&projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.CustomField{}, domain.ErrUnauthorized
		}
		return domain.CustomField{}, err
	}
	var count int
	var maxOrder sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*), MAX(sort_order) FROM flow_custom_fields WHERE project_id = ? AND business_id = ?",
		in.ProjectID, businessID).Scan(&count, &maxOrder); err != nil {
		return domain.CustomField{}, err
	}
	if count >= domain.MaxFieldsPerProject {
		return domain.CustomField{}, domain.ErrFieldLimit
	}
	sortOrder := int64(0)
	if maxOrder.Valid {
		sortOrder = maxOrder.Int64 + domain.FieldOrderSpacing
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO flow_custom_fields
		(business_id, project_id, name, type, options, required, default_value, sort_order, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		businessID, in.ProjectID, in.Name, in.Type, optionsJSON, in.Required, defaultJSON, sortOrder, createdBy)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.CustomField{}, domain.ErrFieldNameTaken
		}
		logger.Ctx(ctx).Error("CreateCustomField failed", logger.Error(err))
		return domain.CustomField{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.CustomField{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.CustomField{}, err
	}
	return r.GetCustomField(ctx, id, businessID)
}

// UpdateCustomField replaces a field's definition and, in the same
// transaction, clears every removed option out of the task values.
func (r *Repository) UpdateCustomField(ctx context.Context, existing domain.CustomField, in domain.CustomFieldInput, removedOptionIDs []string) (domain.CustomField, error) {
	optionsJSON, defaultJSON, err := fieldJSONColumns(in.Options, in.Default)
	if err != nil {
		return domain.CustomField{}, err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.CustomField{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `UPDATE flow_custom_fields
		SET name = ?, options = ?, required = ?, default_value = ?
		WHERE id = ? AND business_id = ?`,
		in.Name, optionsJSON, in.Required, defaultJSON, existing.ID, existing.BusinessID); err != nil {
		if isDuplicateKey(err) {
			return domain.CustomField{}, domain.ErrFieldNameTaken
		}
		logger.Ctx(ctx).Error("UpdateCustomField failed", logger.Error(err))
		return domain.CustomField{}, err
	}
	if err := clearRemovedOptions(ctx, tx, existing, removedOptionIDs); err != nil {
		logger.Ctx(ctx).Error("UpdateCustomField option cleanup failed", logger.Error(err))
		return domain.CustomField{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.CustomField{}, err
	}
	return r.GetCustomField(ctx, existing.ID, existing.BusinessID)
}

func clearRemovedOptions(ctx context.Context, tx *sql.Tx, field domain.CustomField, removed []string) error {
	if len(removed) == 0 {
		return nil
	}
	if field.Type == domain.FieldTypeSelect {
		q, args, err := sq.Delete("flow_task_field_values").
			Where(sq.Eq{"field_id": field.ID, "business_id": field.BusinessID, "value_text": removed}).
			PlaceholderFormat(sq.Question).ToSql()
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, q, args...)
		return err
	}
	for _, optionID := range removed {
		if _, err := tx.ExecContext(ctx, `UPDATE flow_task_field_values
			SET value_json = JSON_REMOVE(value_json, JSON_UNQUOTE(JSON_SEARCH(value_json, 'one', ?)))
			WHERE field_id = ? AND business_id = ? AND JSON_SEARCH(value_json, 'one', ?) IS NOT NULL`,
			optionID, field.ID, field.BusinessID, optionID); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx,
		"DELETE FROM flow_task_field_values WHERE field_id = ? AND business_id = ? AND JSON_LENGTH(value_json) = 0",
		field.ID, field.BusinessID)
	return err
}

// ReorderCustomFields rewrites a project's whole order. The set is locked and
// checked inside the transaction, so a field created or deleted after the
// caller loaded the list fails the check.
func (r *Repository) ReorderCustomFields(ctx context.Context, businessID, projectID int64, ids []int64) ([]domain.CustomField, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx,
		"SELECT id FROM flow_custom_fields WHERE project_id = ? AND business_id = ? FOR UPDATE", projectID, businessID)
	if err != nil {
		return nil, err
	}
	var current []int64
	for rows.Next() {
		var id int64
		if scanErr := rows.Scan(&id); scanErr != nil {
			_ = rows.Close()
			return nil, scanErr
		}
		current = append(current, id)
	}
	if closeErr := rows.Close(); closeErr != nil {
		return nil, closeErr
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	if err := domain.ValidateFieldOrder(current, ids); err != nil {
		return nil, err
	}
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			"UPDATE flow_custom_fields SET sort_order = ? WHERE id = ? AND project_id = ? AND business_id = ?",
			i*domain.FieldOrderSpacing, id, projectID, businessID); err != nil {
			logger.Ctx(ctx).Error("ReorderCustomFields failed", logger.Error(err))
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.ListCustomFields(ctx, businessID, projectID)
}

// DeleteCustomField removes a field with its values and returns how many
// tasks had one.
func (r *Repository) DeleteCustomField(ctx context.Context, id, businessID int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		"DELETE FROM flow_task_field_values WHERE field_id = ? AND business_id = ?", id, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteCustomField values failed", logger.Error(err))
		return 0, err
	}
	cleared, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	res, err = tx.ExecContext(ctx, "DELETE FROM flow_custom_fields WHERE id = ? AND business_id = ?", id, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteCustomField failed", logger.Error(err))
		return 0, err
	}
	if affected, affErr := res.RowsAffected(); affErr != nil || affected == 0 {
		return 0, domain.ErrFieldNotFound
	}
	return cleared, tx.Commit()
}

// SetTaskFieldValues applies a batch of validated writes in one transaction.
func (r *Repository) SetTaskFieldValues(ctx context.Context, businessID, taskID int64, writes []domain.FieldValueWrite) error {
	if len(writes) == 0 {
		return nil
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, w := range writes {
		if w.Value.IsEmpty() {
			if _, err := tx.ExecContext(ctx,
				"DELETE FROM flow_task_field_values WHERE task_id = ? AND field_id = ? AND business_id = ?",
				taskID, w.Field.ID, businessID); err != nil {
				logger.Ctx(ctx).Error("SetTaskFieldValues clear failed", logger.Error(err))
				return err
			}
			continue
		}
		text, number, date, jsonValue, colErr := fieldValueColumns(w.Field.Type, w.Value)
		if colErr != nil {
			return colErr
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO flow_task_field_values
			(business_id, task_id, field_id, value_text, value_number, value_date, value_json)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE value_text = VALUES(value_text), value_number = VALUES(value_number),
				value_date = VALUES(value_date), value_json = VALUES(value_json)`,
			businessID, taskID, w.Field.ID, text, number, date, jsonValue); err != nil {
			logger.Ctx(ctx).Error("SetTaskFieldValues failed", logger.Error(err))
			return err
		}
	}
	return tx.Commit()
}

// fieldValueColumns places a canonical value in the column its type uses.
func fieldValueColumns(fieldType string, v domain.FieldValue) (text, number, date, jsonValue any, err error) {
	switch fieldType {
	case domain.FieldTypeText, domain.FieldTypeURL, domain.FieldTypeSelect:
		text = v.Value
	case domain.FieldTypeNumber:
		number = v.Value
	case domain.FieldTypeCheckbox:
		number = 0
		if v.Value == "true" {
			number = 1
		}
	case domain.FieldTypeDate:
		date = v.Value
	default:
		b, marshalErr := json.Marshal(v.Values)
		if marshalErr != nil {
			return nil, nil, nil, nil, marshalErr
		}
		jsonValue = string(b)
	}
	return text, number, date, jsonValue, nil
}

// ListTaskFieldValues reads the values of a set of tasks in one query, keyed
// by task id then field id.
// dynamic-query: squirrel
func (r *Repository) ListTaskFieldValues(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]map[int64]domain.FieldValue, error) {
	out := make(map[int64]map[int64]domain.FieldValue, len(taskIDs))
	if len(taskIDs) == 0 {
		return out, nil
	}
	q, args, err := sq.Select("v.task_id", "v.field_id", "f.type", "v.value_text", "v.value_number",
		"DATE_FORMAT(v.value_date, '%Y-%m-%d')", "v.value_json").
		From("flow_task_field_values v").
		Join("flow_custom_fields f ON f.id = v.field_id AND f.business_id = v.business_id").
		Where(sq.Eq{"v.business_id": businessID, "v.task_id": taskIDs}).
		PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListTaskFieldValues failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var taskID, fieldID int64
		var fieldType string
		var text, number, date sql.NullString
		var jsonValue []byte
		if scanErr := rows.Scan(&taskID, &fieldID, &fieldType, &text, &number, &date, &jsonValue); scanErr != nil {
			return nil, scanErr
		}
		value, valueErr := fieldValueFromColumns(fieldType, text.String, number.String, date.String, jsonValue)
		if valueErr != nil {
			return nil, valueErr
		}
		if value.IsEmpty() {
			continue
		}
		if out[taskID] == nil {
			out[taskID] = map[int64]domain.FieldValue{}
		}
		out[taskID][fieldID] = value
	}
	return out, rows.Err()
}

func fieldValueFromColumns(fieldType, text, number, date string, jsonValue []byte) (domain.FieldValue, error) {
	switch fieldType {
	case domain.FieldTypeText, domain.FieldTypeURL, domain.FieldTypeSelect:
		return domain.FieldValue{Value: text}, nil
	case domain.FieldTypeNumber:
		if number == "" {
			return domain.FieldValue{}, nil
		}
		return domain.FieldValue{Value: domain.CanonicalFieldNumber(number)}, nil
	case domain.FieldTypeCheckbox:
		if number == "" {
			return domain.FieldValue{}, nil
		}
		if strings.HasPrefix(number, "1") {
			return domain.FieldValue{Value: "true"}, nil
		}
		return domain.FieldValue{Value: "false"}, nil
	case domain.FieldTypeDate:
		return domain.FieldValue{Value: date}, nil
	}
	var values []string
	if len(jsonValue) > 0 {
		if err := json.Unmarshal(jsonValue, &values); err != nil {
			return domain.FieldValue{}, err
		}
	}
	return domain.FieldValue{Values: values}, nil
}

// fieldFilterExpr turns a validated clause into a predicate on flow_tasks t.
// Each one is an EXISTS on the (task_id, field_id) unique key.
func fieldFilterExpr(f domain.TaskFieldFilter, businessID int64) sq.Sqlizer {
	exists := func(negate bool, cond string, args ...any) sq.Sqlizer {
		prefix := "EXISTS"
		if negate {
			prefix = "NOT EXISTS"
		}
		return sq.Expr(prefix+` (SELECT 1 FROM flow_task_field_values fv
			WHERE fv.task_id = t.id AND fv.field_id = ? AND fv.business_id = ? AND `+cond+`)`,
			append([]any{f.Field.ID, businessID}, args...)...)
	}
	anyArgs := func() []any {
		args := make([]any, 0, len(f.Values))
		for _, v := range f.Values {
			args = append(args, v)
		}
		return args
	}
	const decimal = "CAST(? AS DECIMAL(24,6))"

	switch {
	case f.Op == domain.FieldOpContains:
		return exists(false, "fv.value_text LIKE ? ESCAPE '!'", "%"+escapeLike(f.Values[0])+"%")
	case f.Op == domain.FieldOpRange:
		col, placeholder := "fv.value_date", "?"
		if f.Field.Type == domain.FieldTypeNumber {
			col, placeholder = "fv.value_number", decimal
		}
		var conds []string
		var args []any
		if f.Values[0] != "" {
			conds, args = append(conds, col+" >= "+placeholder), append(args, f.Values[0])
		}
		if f.Values[1] != "" {
			conds, args = append(conds, col+" <= "+placeholder), append(args, f.Values[1])
		}
		return exists(false, strings.Join(conds, " AND "), args...)
	case f.Field.Type == domain.FieldTypeMultiSelect || f.Field.Type == domain.FieldTypePerson:
		conds := make([]string, 0, len(f.Values))
		for range f.Values {
			conds = append(conds, "JSON_CONTAINS(fv.value_json, JSON_QUOTE(?))")
		}
		return exists(false, "("+strings.Join(conds, " OR ")+")", anyArgs()...)
	case f.Field.Type == domain.FieldTypeCheckbox:
		// Unchecked and never set are the same thing to the person filtering.
		return exists(f.Values[0] != "true", "fv.value_number = 1")
	case f.Field.Type == domain.FieldTypeNumber:
		return exists(false, "fv.value_number = "+decimal, f.Values[0])
	case f.Field.Type == domain.FieldTypeDate:
		return exists(false, "fv.value_date = ?", f.Values[0])
	}
	return exists(false, "fv.value_text IN ("+sq.Placeholders(len(f.Values))+")", anyArgs()...)
}

func escapeLike(s string) string {
	return strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(s)
}

// applyFieldSort orders a task listing by one custom field. The join is on the
// (task_id, field_id) unique key, so it adds at most one row per task. Tasks
// without a value come last in either direction.
func applyFieldSort(builder sq.SelectBuilder, sort domain.TaskFieldSort, businessID int64) sq.SelectBuilder {
	dir := "ASC"
	if sort.Desc {
		dir = "DESC"
	}
	builder = builder.
		LeftJoin("flow_task_field_values sv ON sv.task_id = t.id AND sv.field_id = ? AND sv.business_id = ?",
			sort.Field.ID, businessID).
		OrderBy("(sv.id IS NULL) ASC")
	switch sort.Field.Type {
	case domain.FieldTypeNumber, domain.FieldTypeCheckbox:
		builder = builder.OrderBy("sv.value_number " + dir)
	case domain.FieldTypeDate:
		builder = builder.OrderBy("sv.value_date " + dir)
	case domain.FieldTypeSelect:
		// A select sorts by the position of its option, not by the option id.
		args := make([]any, 0, len(sort.Field.Options))
		for _, o := range sort.Field.Options {
			args = append(args, o.ID)
		}
		if len(args) > 0 {
			builder = builder.OrderByClause("FIELD(sv.value_text, "+sq.Placeholders(len(args))+") "+dir, args...)
		}
	default:
		builder = builder.OrderBy("sv.value_text " + dir)
	}
	return builder
}
