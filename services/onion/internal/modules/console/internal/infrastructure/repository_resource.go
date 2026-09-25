package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func (r *Repository) CreateResourceRow(ctx context.Context, table string, values map[string]any) (int64, error) {
	columns, _, err := r.DescribeResourceTable(ctx, table)
	if err != nil {
		return 0, err
	}
	allowed := map[string]ResourceColumnRow{}
	for _, col := range columns {
		if isWritableColumn(col) {
			allowed[col.Name] = col
		}
	}

	insertColumns := []string{}
	args := []any{}
	for name, value := range values {
		if _, ok := allowed[name]; !ok {
			continue
		}
		normalized, err := normalizeResourceValue(value)
		if err != nil {
			return 0, err
		}
		insertColumns = append(insertColumns, quoteIdentifier(name))
		args = append(args, normalized)
	}
	if len(insertColumns) == 0 {
		return 0, fmt.Errorf("no writable columns provided")
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(insertColumns)), ",")
	// #nosec G201 -- table, writable columns, and placeholders are derived from validated metadata and values remain bound.
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		quoteIdentifier(table),
		strings.Join(insertColumns, ", "),
		placeholders,
	)
	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateResourceRow(ctx context.Context, table, id string, values map[string]any) error {
	columns, primaryKey, err := r.DescribeResourceTable(ctx, table)
	if err != nil {
		return err
	}
	if primaryKey == "" {
		return fmt.Errorf("table %s has no primary key", table)
	}
	allowed := map[string]ResourceColumnRow{}
	for _, col := range columns {
		if isWritableColumn(col) {
			allowed[col.Name] = col
		}
	}

	sets := []string{}
	args := []any{}
	for name, value := range values {
		if _, ok := allowed[name]; !ok {
			continue
		}
		normalized, err := normalizeResourceValue(value)
		if err != nil {
			return err
		}
		sets = append(sets, fmt.Sprintf("%s = ?", quoteIdentifier(name)))
		args = append(args, normalized)
	}
	if len(sets) == 0 {
		return fmt.Errorf("no writable columns provided")
	}

	args = append(args, id)
	// #nosec G201 -- table, writable columns, and primaryKey come from validated metadata and are quoted as identifiers.
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = ?",
		quoteIdentifier(table),
		strings.Join(sets, ", "),
		quoteIdentifier(primaryKey),
	)
	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteResourceRow(ctx context.Context, table, id string) error {
	_, primaryKey, err := r.DescribeResourceTable(ctx, table)
	if err != nil {
		return err
	}
	if primaryKey == "" {
		return fmt.Errorf("table %s has no primary key", table)
	}
	// #nosec G201 -- table and primaryKey come from validated metadata and are quoted as identifiers.
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?",
		quoteIdentifier(table),
		quoteIdentifier(primaryKey),
	)
	_, err = r.pool.ExecContext(ctx, query, id)
	return err
}

func scanResourceRows(rows *sql.Rows, columns []ResourceColumnRow) ([]map[string]any, error) {
	result := []map[string]any{}
	for rows.Next() {
		raw := make([]sql.RawBytes, len(columns))
		dest := make([]any, len(columns))
		for i := range raw {
			dest[i] = &raw[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, col := range columns {
			if raw[i] == nil {
				row[col.Name] = nil
			} else {
				row[col.Name] = string(raw[i])
			}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ── Auth ──────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*UserRow, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return userFromGetUserByEmailRow(row), nil
}

// ── Users CRUD ────────────────────────────────────────────────────────────────

// dynamic-query: squirrel
func (r *Repository) ListUsers(
	ctx context.Context,
	limit, offset int,
	id int64,
	email, fullName string,
) ([]UserRow, int64, error) {
	conds := sq.And{}
	if id > 0 {
		conds = append(conds, sq.Eq{"id": id})
	}
	if email != "" {
		conds = append(conds, sq.Eq{"email": email})
	}
	if fullName != "" {
		conds = append(conds, sq.Eq{"full_name": fullName})
	}

	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("users").
		Where(conds).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL, listArgs, err := sq.
		Select(userColumns...).
		From("users").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]UserRow, 0, limit)
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
			&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*UserRow, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return userFromGetUserByIDRow(row), nil
}
