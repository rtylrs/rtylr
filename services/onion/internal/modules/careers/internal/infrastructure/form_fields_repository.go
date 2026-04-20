package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

var ErrFormFieldNotFound = errors.New("form field not found")

const formFieldSelectColumns = `
    id, career_id, owner_id, kind, label,
    COALESCE(placeholder,''), required,
    COALESCE(options_json,'') AS options_json,
    position, created_at, updated_at`

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
	q := `SELECT ` + formFieldSelectColumns + `
	      FROM careers_form_fields
	      WHERE owner_id = ? AND career_id = ?
	      ORDER BY position ASC, id ASC`
	rows, err := r.pool.QueryContext(ctx, q, ownerID, careerID)
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
	q := `SELECT ` + formFieldSelectColumns + `
	      FROM careers_form_fields
	      WHERE career_id = ?
	      ORDER BY position ASC, id ASC`
	rows, err := r.pool.QueryContext(ctx, q, careerID)
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
	q := `INSERT INTO careers_form_fields
	        (career_id, owner_id, kind, label, placeholder, required, options_json, position)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.pool.ExecContext(ctx, q,
		in.CareerID, in.OwnerID, in.Kind, in.Label,
		nullStr(in.Placeholder), in.Required,
		nullStr(in.OptionsJSON), in.Position)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateFormField(ctx context.Context, in domain.FormFieldInput) error {
	q := `UPDATE careers_form_fields SET
	        kind = ?, label = ?, placeholder = ?, required = ?,
	        options_json = ?, position = ?
	      WHERE id = ? AND owner_id = ?`
	_, err := r.pool.ExecContext(ctx, q,
		in.Kind, in.Label, nullStr(in.Placeholder), in.Required,
		nullStr(in.OptionsJSON), in.Position,
		in.ID, in.OwnerID)
	return err
}

func (r *Repository) DeleteFormField(ctx context.Context, ownerID, fieldID int64) error {
	res, err := r.pool.ExecContext(ctx,
		`DELETE FROM careers_form_fields WHERE id = ? AND owner_id = ?`,
		fieldID, ownerID)
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
	q := `SELECT ` + formFieldSelectColumns + `
	      FROM careers_form_fields
	      WHERE id = ? AND owner_id = ? LIMIT 1`
	row := r.pool.QueryRowContext(ctx, q, fieldID, ownerID)
	f, err := scanFormField(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FormField{}, ErrFormFieldNotFound
		}
		return domain.FormField{}, err
	}
	return f, nil
}
