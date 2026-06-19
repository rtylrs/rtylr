package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

// ErrSectionNotFound is used by the service layer to translate a
// failed lookup into a 404. Keeping it alongside the other sentinel
// errors in this package means handlers don't need to import the
// infrastructure package for anything else.
var ErrSectionNotFound = errors.New("section not found")

// sectionSelectColumns is the column projection shared by every
// section read so the order stays locked with scanSection.
var sectionSelectColumns = []string{
	"id", "owner_id", "business_id", "kind",
	"COALESCE(title,'')", "COALESCE(body,'')",
	"COALESCE(media_url,'')", "COALESCE(items_json,'') AS items_json",
	"position", "enabled",
	"created_at", "updated_at",
}

func scanSection(row interface{ Scan(...interface{}) error }) (domain.Section, error) {
	var s domain.Section
	err := row.Scan(
		&s.ID, &s.OwnerID, &s.BusinessID, &s.Kind,
		&s.Title, &s.Body,
		&s.MediaURL, &s.ItemsJSON,
		&s.Position, &s.Enabled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *Repository) ListSections(ctx context.Context, businessID int64, enabledOnly bool) ([]domain.Section, error) {
	b := sq.
		Select(sectionSelectColumns...).
		From("careers_sections").
		Where(sq.Eq{"business_id": businessID})
	if enabledOnly {
		b = b.Where(sq.Eq{"enabled": true})
	}
	q, args, err := b.
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
	q, args, err := sq.
		Insert("careers_sections").
		Columns("owner_id", "business_id", "kind", "title", "body", "media_url", "items_json", "position", "enabled").
		Values(in.OwnerID, in.BusinessID, in.Kind,
			nullStr(in.Title), nullStr(in.Body),
			nullStr(in.MediaURL), nullStr(in.ItemsJSON),
			in.Position, in.Enabled).
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

func (r *Repository) UpdateSection(ctx context.Context, in domain.SectionInput) error {
	q, args, err := sq.
		Update("careers_sections").
		Set("kind", in.Kind).
		Set("title", nullStr(in.Title)).
		Set("body", nullStr(in.Body)).
		Set("media_url", nullStr(in.MediaURL)).
		Set("items_json", nullStr(in.ItemsJSON)).
		Set("position", in.Position).
		Set("enabled", in.Enabled).
		Where(sq.Eq{"id": in.ID}).
		Where(sq.Eq{"business_id": in.BusinessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, q, args...)
	return err
}

// ReorderSections rewrites the position column for a batch of
// sections in one transaction. The admin UI writes the whole
// ordered list after drag-drop; doing it in a txn means a partial
// reorder can't leave duplicate positions behind.
func (r *Repository) ReorderSections(ctx context.Context, businessID int64, orderedIDs []int64) error {
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
		q, args, buildErr := sq.
			Update("careers_sections").
			Set("position", int32(idx)).
			Where(sq.Eq{"id": id}).
			Where(sq.Eq{"business_id": businessID}).
			PlaceholderFormat(sq.Question).
			ToSql()
		if buildErr != nil {
			_ = tx.Rollback()
			return buildErr
		}
		if _, execErr := tx.ExecContext(ctx, q, args...); execErr != nil {
			_ = tx.Rollback()
			return execErr
		}
	}
	return tx.Commit()
}

func (r *Repository) DeleteSection(ctx context.Context, businessID, sectionID int64) error {
	q, args, err := sq.
		Delete("careers_sections").
		Where(sq.Eq{"id": sectionID}).
		Where(sq.Eq{"business_id": businessID}).
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
		return ErrSectionNotFound
	}
	return nil
}

// GetSection is used by the handler before mutating — lets us 404
// cleanly on cross-tenant access rather than silently no-op via the
// owner_id filter in UPDATE/DELETE.
func (r *Repository) GetSection(ctx context.Context, businessID, sectionID int64) (domain.Section, error) {
	q, args, err := sq.
		Select(sectionSelectColumns...).
		From("careers_sections").
		Where(sq.Eq{"id": sectionID}).
		Where(sq.Eq{"business_id": businessID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Section{}, err
	}
	row := r.pool.QueryRowContext(ctx, q, args...)
	s, err := scanSection(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Section{}, ErrSectionNotFound
		}
		return domain.Section{}, err
	}
	return s, nil
}
