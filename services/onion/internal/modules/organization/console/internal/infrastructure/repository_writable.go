package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

func isWritableColumn(c ResourceColumnRow) bool {
	extra := strings.ToLower(c.Extra)
	return !c.IsPrimary && !isSensitiveResourceColumn(c.Name) && !strings.Contains(extra, "auto_increment") && !strings.Contains(extra, "generated")
}

func isSensitiveResourceColumn(name string) bool {
	column := strings.ToLower(name)
	sensitiveParts := []string{
		"password",
		"secret",
		"token",
		"api_key",
		"apikey",
		"private_key",
		"credential",
		"cookie",
	}
	for _, part := range sensitiveParts {
		if strings.Contains(column, part) {
			return true
		}
	}
	return column == "session_id" || strings.HasSuffix(column, "_session_id")
}

func isAllowedResourceTable(table string) bool {
	if _, ok := allowedResourceTables[table]; ok {
		return true
	}
	for _, prefix := range allowedResourceTablePrefixes {
		if strings.HasPrefix(table, prefix) {
			return true
		}
	}
	return false
}

func validateResourceTable(table string) error {
	if !isAllowedResourceTable(table) {
		return fmt.Errorf("%w: %s", ErrResourceNotAllowed, table)
	}
	return nil
}

func visibleResourceColumns(columns []ResourceColumnRow) []ResourceColumnRow {
	visible := make([]ResourceColumnRow, 0, len(columns))
	for _, col := range columns {
		if !isSensitiveResourceColumn(col.Name) {
			visible = append(visible, col)
		}
	}
	return visible
}

func normalizeResourceValue(value any) (any, error) {
	switch v := value.(type) {
	case nil, string, bool, float64:
		return v, nil
	case map[string]any, []any:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	}
}

var userColumns = []string{
	"id", "email", "full_name", "business_id", "is_console_admin",
	"is_banned", "ban_reason", "banned_at", "created_at", "password_hash",
}

// ── Generic Resource CRUD ────────────────────────────────────────────────────

func (r *Repository) ListResourceTables(ctx context.Context) ([]ResourceTableRow, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT table_name, table_comment
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []ResourceTableRow{}
	for rows.Next() {
		var table ResourceTableRow
		if err := rows.Scan(&table.Name, &table.Comment); err != nil {
			return nil, err
		}
		if !isAllowedResourceTable(table.Name) {
			continue
		}
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

func (r *Repository) DescribeResourceTable(ctx context.Context, table string) ([]ResourceColumnRow, string, error) {
	if err := validateResourceTable(table); err != nil {
		return nil, "", err
	}
	rows, err := r.pool.QueryContext(ctx, `
		SELECT column_name, data_type, is_nullable, column_key, column_default, extra, column_comment
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		ORDER BY ordinal_position ASC`, table)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	columns := []ResourceColumnRow{}
	primaryKey := ""
	for rows.Next() {
		var nullable, key string
		var col ResourceColumnRow
		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &key, &col.Default, &col.Extra, &col.Comment); err != nil {
			return nil, "", err
		}
		col.IsNullable = nullable == "YES"
		col.IsPrimary = key == "PRI"
		if col.IsPrimary && primaryKey == "" {
			primaryKey = col.Name
		}
		columns = append(columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	columns = visibleResourceColumns(columns)
	if len(columns) == 0 {
		return nil, "", sql.ErrNoRows
	}
	primaryKey = ""
	for _, col := range columns {
		if col.IsPrimary {
			primaryKey = col.Name
			break
		}
	}
	return columns, primaryKey, nil
}

func (r *Repository) ListResourceRows(ctx context.Context, table string, limit, offset int) (*ResourceRowsResult, error) {
	columns, primaryKey, err := r.DescribeResourceTable(ctx, table)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	// #nosec G201 -- table was validated by DescribeResourceTable and quoteIdentifier escapes the identifier.
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdentifier(table))
	if err := r.pool.QueryRowContext(ctx, countSQL).Scan(&total); err != nil {
		return nil, err
	}

	columnNames := make([]string, 0, len(columns))
	for _, col := range columns {
		columnNames = append(columnNames, quoteIdentifier(col.Name))
	}
	orderColumn := columns[0].Name
	if primaryKey != "" {
		orderColumn = primaryKey
	}
	// #nosec G201 -- table, columns, and orderColumn come from validated database metadata and are quoted as identifiers.
	query := fmt.Sprintf(
		"SELECT %s FROM %s ORDER BY %s DESC LIMIT ? OFFSET ?",
		strings.Join(columnNames, ", "),
		quoteIdentifier(table),
		quoteIdentifier(orderColumn),
	)
	rows, err := r.pool.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultRows, err := scanResourceRows(rows, columns)
	if err != nil {
		return nil, err
	}
	return &ResourceRowsResult{Columns: columns, PrimaryKey: primaryKey, Rows: resultRows, Total: total}, nil
}

func (r *Repository) GetResourceRow(ctx context.Context, table, id string) (map[string]any, []ResourceColumnRow, string, error) {
	columns, primaryKey, err := r.DescribeResourceTable(ctx, table)
	if err != nil {
		return nil, nil, "", err
	}
	if primaryKey == "" {
		return nil, nil, "", fmt.Errorf("table %s has no primary key", table)
	}

	columnNames := make([]string, 0, len(columns))
	for _, col := range columns {
		columnNames = append(columnNames, quoteIdentifier(col.Name))
	}
	// #nosec G201 -- table, columns, and primaryKey come from validated database metadata and are quoted as identifiers.
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ? LIMIT 1",
		strings.Join(columnNames, ", "),
		quoteIdentifier(table),
		quoteIdentifier(primaryKey),
	)
	rows, err := r.pool.QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, "", err
	}
	defer rows.Close()

	resultRows, err := scanResourceRows(rows, columns)
	if err != nil {
		return nil, nil, "", err
	}
	if len(resultRows) == 0 {
		return nil, nil, "", sql.ErrNoRows
	}
	return resultRows[0], columns, primaryKey, nil
}
