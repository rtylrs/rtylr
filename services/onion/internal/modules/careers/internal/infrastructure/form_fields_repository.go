package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

var ErrFormFieldNotFound = errors.New("form field not found")

// formFieldSelectColumns is the column projection shared by every
// form-field read so the order stays locked with scanFormField.
var formFieldSelectColumns = []string{
	"id", "career_id", "owner_id", "kind", "label",
	"COALESCE(placeholder,'')", "required",
	"COALESCE(options_json,'') AS options_json",
	"position", "created_at", "updated_at",
}

func scanFormField(row interface{ Scan(...interface{}) error }) (domain.FormField, error) {
	var f domain.FormField
	err := row.Scan(
		&f.ID, &f.CareerID, &f.OwnerID, &f.Kind, &f.Label,
		&f.Placeholder, &f.Required,
		&f.OptionsJSON,
		&f.Position, &f.CreatedAt, &f.UpdatedAt,
	)
	return f, err
}

func (r *Repository) ListFormFields(ctx context.Context, ownerID, careerID int64) ([]domain.FormField, error) {
	q, args, err := sq.
		Select(formFieldSelectColumns...).
		From("careers_form_fields").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.Eq{"career_id": careerID}).
		OrderBy("position ASC", "id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FormField, 0)
	for rows.Next() {
		f, err := scanFormField(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// ListFormFieldsByCareer skips the owner filter — used by the public
// apply endpoint where the caller is anonymous and the career_id has
// already been resolved (and confirmed published) via the
// customization slug lookup.
func (r *Repository) ListFormFieldsByCareer(ctx context.Context, careerID int64) ([]domain.FormField, error) {
	q, args, err := sq.
		Select(formFieldSelectColumns...).
		From("careers_form_fields").
		Where(sq.Eq{"career_id": careerID}).
		OrderBy("position ASC", "id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FormField, 0)
	for rows.Next() {
		f, err := scanFormField(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *Repository) CreateFormField(ctx context.Context, in domain.FormFieldInput) (int64, error) {
	q, args, err := sq.
		Insert("careers_form_fields").
		Columns("career_id", "owner_id", "kind", "label", "placeholder", "required", "options_json", "position").
		Values(in.CareerID, in.OwnerID, in.Kind, in.Label,
			nullStr(in.Placeholder), in.Required,
			nullStr(in.OptionsJSON), in.Position).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateFormField(ctx context.Context, in domain.FormFieldInput) error {
	q, args, err := sq.
		Update("careers_form_fields").
		Set("kind", in.Kind).
		Set("label", in.Label).
		Set("placeholder", nullStr(in.Placeholder)).
		Set("required", in.Required).
		Set("options_json", nullStr(in.OptionsJSON)).
		Set("position", in.Position).
		Where(sq.Eq{"id": in.ID}).
		Where(sq.Eq{"owner_id": in.OwnerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, q, args...)
	return err
}

func (r *Repository) DeleteFormField(ctx context.Context, ownerID, fieldID int64) error {
	q, args, err := sq.
		Delete("careers_form_fields").
		Where(sq.Eq{"id": fieldID}).
		Where(sq.Eq{"owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrFormFieldNotFound
	}
	return nil
}

// GetFormField is used before update/delete so the handler can 404
// cleanly on cross-tenant access.
func (r *Repository) GetFormField(ctx context.Context, ownerID, fieldID int64) (domain.FormField, error) {
	q, args, err := sq.
		Select(formFieldSelectColumns...).
		From("careers_form_fields").
		Where(sq.Eq{"id": fieldID}).
		Where(sq.Eq{"owner_id": ownerID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.FormField{}, err
	}
	row := r.pool.QueryRowContext(ctx, q, args...)
	f, err := scanFormField(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FormField{}, ErrFormFieldNotFound
		}
		return domain.FormField{}, err
	}
	return f, nil
}
