package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

// ErrSectionNotFound is used by the service layer to translate a
// failed lookup into a 404. Keeping it alongside the other sentinel
// errors in this package means handlers don't need to import the
// infrastructure package for anything else.
var ErrSectionNotFound = errors.New("section not found")

const sectionSelectColumns = `
    id, owner_id, kind,
    COALESCE(title,''), COALESCE(body,''),
    COALESCE(media_url,''), COALESCE(items_json,'') AS items_json,
    position, enabled,
    created_at, updated_at`

func scanSection(row interface{ Scan(...interface{}) error }) (domain.Section, error) {
	var s domain.Section
	err := row.Scan(
		&s.ID, &s.OwnerID, &s.Kind,
		&s.Title, &s.Body,
		&s.MediaURL, &s.ItemsJSON,
		&s.Position, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *Repository) ListSections(ctx context.Context, ownerID int64, enabledOnly bool) ([]domain.Section, error) {
	q := `SELECT ` + sectionSelectColumns + ` FROM careers_sections WHERE owner_id = ?`
	if enabledOnly {
		q += ` AND enabled = true`
	}
	q += ` ORDER BY position ASC, id ASC`

	rows, err := r.pool.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Section, 0)
	for rows.Next() {
		s, err := scanSection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) CreateSection(ctx context.Context, in domain.SectionInput) (int64, error) {
	q := `INSERT INTO careers_sections
	        (owner_id, kind, title, body, media_url, items_json, position, enabled)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.pool.ExecContext(ctx, q,
		in.OwnerID, in.Kind,
		nullStr(in.Title), nullStr(in.Body),
		nullStr(in.MediaURL), nullStr(in.ItemsJSON),
		in.Position, in.Enabled)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateSection(ctx context.Context, in domain.SectionInput) error {
	q := `UPDATE careers_sections SET
	        kind = ?, title = ?, body = ?, media_url = ?, items_json = ?,
	        position = ?, enabled = ?
	      WHERE id = ? AND owner_id = ?`
	_, err := r.pool.ExecContext(ctx, q,
		in.Kind,
		nullStr(in.Title), nullStr(in.Body),
		nullStr(in.MediaURL), nullStr(in.ItemsJSON),
		in.Position, in.Enabled,
		in.ID, in.OwnerID)
	return err
}

// ReorderSections rewrites the position column for a batch of
// sections in one transaction. The admin UI writes the whole
// ordered list after drag-drop; doing it in a txn means a partial
// reorder can't leave duplicate positions behind.
func (r *Repository) ReorderSections(ctx context.Context, ownerID int64, orderedIDs []int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	for idx, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx,
			`UPDATE careers_sections SET position = ? WHERE id = ? AND owner_id = ?`,
			int32(idx), id, ownerID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) DeleteSection(ctx context.Context, ownerID, sectionID int64) error {
	res, err := r.pool.ExecContext(ctx,
		`DELETE FROM careers_sections WHERE id = ? AND owner_id = ?`,
		sectionID, ownerID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSectionNotFound
	}
	return nil
}

// GetSection is used by the handler before mutating — lets us 404
// cleanly on cross-tenant access rather than silently no-op via the
// owner_id filter in UPDATE/DELETE.
func (r *Repository) GetSection(ctx context.Context, ownerID, sectionID int64) (domain.Section, error) {
	q := `SELECT ` + sectionSelectColumns + ` FROM careers_sections WHERE id = ? AND owner_id = ? LIMIT 1`
	row := r.pool.QueryRowContext(ctx, q, sectionID, ownerID)
	s, err := scanSection(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Section{}, ErrSectionNotFound
		}
		return domain.Section{}, err
	}
	return s, nil
}
