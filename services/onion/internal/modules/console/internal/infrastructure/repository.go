package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type UserRow struct {
	ID             int64
	Email          string
	FullName       sql.NullString
	BusinessID     sql.NullInt64
	IsConsoleAdmin bool
	IsBanned       bool
	BanReason      sql.NullString
	BannedAt       sql.NullTime
	CreatedAt      sql.NullTime
	PasswordHash   string
}

type BusinessRow struct {
	ID        int64
	Name      string
	OwnerID   int64
	CreatedAt sql.NullTime
}

type CreateBusinessOwnerParams struct {
	Email              string
	PasswordHash       string
	FullName           string
	BusinessName       string
	Industry           string
	Country            string
	Timezone           string
	Currency           string
	GrantPlan          bool
	PlanID             int64
	SubscriptionStatus string
	GrantDays          int
}

type CreateBusinessOwnerRow struct {
	OwnerID        int64
	BusinessID     int64
	SubscriptionID sql.NullInt64
}

type SubscriptionRow struct {
	ID                   int64
	OwnerID              int64
	OwnerName            sql.NullString
	PlanID               int64
	PlanName             string
	Status               string
	SeatCount            int
	ActiveSeatCount      int
	IncludedSeatCount    int
	BillableSeatCount    int
	TrialEndsAt          sql.NullTime
	NextBillingDate      sql.NullTime
	CurrentPeriodEndsAt  sql.NullTime
	PaddleSubscriptionID sql.NullString
	PaddleCustomerID     sql.NullString
	CreatedAt            sql.NullTime
	UpdatedAt            sql.NullTime
	OwnerEmail           sql.NullString
	BusinessID           sql.NullInt64
	BusinessName         sql.NullString
}

type SubscriptionWriteParams struct {
	OwnerID              int64
	PlanID               int64
	Status               string
	SeatCount            int
	TrialEndsAt          sql.NullTime
	NextBillingDate      sql.NullTime
	CurrentPeriodEndsAt  sql.NullTime
	PaddleSubscriptionID sql.NullString
	PaddleCustomerID     sql.NullString
}

type PlanRow struct {
	ID                      int64
	Name                    string
	Slug                    string
	PricePerSeatCents       int
	AnnualPricePerSeatCents int
	TrialDays               int
	MaxSeats                sql.NullInt64
	MaxLocations            sql.NullInt64
	IsActive                bool
	IsSolo                  bool
	IsFeatured              bool
	TierOrder               int
	Tagline                 string
	CreatedAt               sql.NullTime
}

type GrantRow struct {
	ID              int64
	UserID          int64
	UserName        sql.NullString
	PlanID          int64
	PlanName        string
	GrantedBy       int64
	GranterName     sql.NullString
	Note            sql.NullString
	ExpiresAt       sql.NullTime
	RevokedAt       sql.NullTime
	CreatedAt       sql.NullTime
	UserEmail       sql.NullString
	BusinessID      sql.NullInt64
	BusinessName    sql.NullString
	ActiveSeatCount int
}

type AddonCatalogRow struct {
	ID          string
	Name        string
	Description sql.NullString
	Icon        sql.NullString
	Price       sql.NullString
	IsFree      sql.NullBool
	SortOrder   sql.NullInt64
}

type AddonGrantRow struct {
	ID             int64
	UserID         int64
	UserName       sql.NullString
	UserEmail      sql.NullString
	BusinessID     int64
	BusinessName   sql.NullString
	AddonID        string
	AddonName      string
	AddonPrice     sql.NullString
	GrantedBy      int64
	GranterName    sql.NullString
	Note           sql.NullString
	ExpiresAt      sql.NullTime
	RevokedAt      sql.NullTime
	EnabledByGrant bool
	CreatedAt      sql.NullTime
}

type AddonGrantScopeRow struct {
	UserID         int64
	BusinessID     int64
	AddonID        string
	EnabledByGrant bool
}

type StatsRow struct {
	TotalUsers        int64
	BannedUsers       int64
	TotalBusinesses   int64
	ActiveSubs        int64
	TrialSubs         int64
	ManualGrants      int64
	ManualAddonGrants int64
	ConsoleAdmins     int64
}

type ResourceTableRow struct {
	Name    string
	Comment sql.NullString
}

type ResourceColumnRow struct {
	Name       string
	DataType   string
	IsNullable bool
	IsPrimary  bool
	Default    sql.NullString
	Extra      string
	Comment    sql.NullString
}

type ResourceRowsResult struct {
	Columns    []ResourceColumnRow
	PrimaryKey string
	Rows       []map[string]any
	Total      int64
}

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

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
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

func (r *Repository) DescribeResourceTable(ctx context.Context, table string) ([]ResourceColumnRow, string, error) {
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
	query, args, err := sq.
		Select(userColumns...).
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
		&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

// ── Users CRUD ────────────────────────────────────────────────────────────────

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
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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
	query, args, err := sq.
		Select(userColumns...).
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
		&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string, isConsoleAdmin bool) (int64, error) {
	query, args, err := sq.
		Insert("users").
		Columns("email", "password_hash", "full_name", "is_console_admin").
		Values(email, passwordHash, fullName, isConsoleAdmin).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateBusinessOwner(ctx context.Context, params CreateBusinessOwnerParams) (*CreateBusinessOwnerRow, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	userRes, err := tx.ExecContext(ctx, `
INSERT INTO users (email, password_hash, full_name, onboarded, auth_provider)
VALUES (?, ?, ?, TRUE, 'password')`,
		params.Email,
		params.PasswordHash,
		nullableString(params.FullName),
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	ownerID, err := userRes.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	businessRes, err := tx.ExecContext(ctx, `
INSERT INTO businesses (name, owner_id, industry, country, timezone, currency, is_active)
VALUES (?, ?, ?, ?, ?, ?, TRUE)`,
		params.BusinessName,
		ownerID,
		nullableString(params.Industry),
		nullableString(params.Country),
		params.Timezone,
		params.Currency,
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	businessID, err := businessRes.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if _, err = tx.ExecContext(ctx, `
UPDATE users
SET business_id = ?, onboarded = TRUE, updated_at = NOW()
WHERE id = ?`,
		businessID,
		ownerID,
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	appsJSON := `["dash","pos","erp","hr","crm","finance","flow","insights"]`
	if params.PlanID > 0 {
		if err = tx.QueryRowContext(ctx, `
SELECT apps
FROM plans
WHERE id = ? AND is_active = TRUE
LIMIT 1`,
			params.PlanID,
		).Scan(&appsJSON); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if !json.Valid([]byte(appsJSON)) {
			_ = tx.Rollback()
			return nil, fmt.Errorf("plan %d has invalid apps JSON", params.PlanID)
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO user_app_permissions (business_id, owner_id, user_id, apps, role, is_active)
VALUES (?, ?, ?, ?, 'owner', TRUE)
ON DUPLICATE KEY UPDATE
  business_id = VALUES(business_id),
  apps = VALUES(apps),
  role = VALUES(role),
  is_active = TRUE,
  updated_at = NOW()`,
		businessID,
		ownerID,
		ownerID,
		appsJSON,
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	result := &CreateBusinessOwnerRow{
		OwnerID:    ownerID,
		BusinessID: businessID,
	}
	if params.GrantPlan {
		var trialEndsAt any
		var periodEndsAt any
		if params.GrantDays > 0 {
			endsAt := time.Now().UTC().AddDate(0, 0, params.GrantDays)
			trialEndsAt = endsAt
			periodEndsAt = endsAt
		}

		subRes, subErr := tx.ExecContext(ctx, `
INSERT INTO subscriptions (owner_id, plan_id, status, seat_count, trial_ends_at, current_period_ends_at)
VALUES (?, ?, ?, 1, ?, ?)
ON DUPLICATE KEY UPDATE
  plan_id = VALUES(plan_id),
  status = VALUES(status),
  seat_count = VALUES(seat_count),
  trial_ends_at = VALUES(trial_ends_at),
  current_period_ends_at = VALUES(current_period_ends_at),
  updated_at = NOW()`,
			ownerID,
			params.PlanID,
			params.SubscriptionStatus,
			trialEndsAt,
			periodEndsAt,
		)
		if subErr != nil {
			_ = tx.Rollback()
			return nil, subErr
		}
		if subID, idErr := subRes.LastInsertId(); idErr == nil && subID > 0 {
			result.SubscriptionID = sql.NullInt64{Int64: subID, Valid: true}
		} else {
			var existingID int64
			if scanErr := tx.QueryRowContext(ctx, "SELECT id FROM subscriptions WHERE owner_id = ? LIMIT 1", ownerID).Scan(&existingID); scanErr != nil {
				_ = tx.Rollback()
				return nil, scanErr
			}
			result.SubscriptionID = sql.NullInt64{Int64: existingID, Valid: true}
		}
	}

	_, _ = tx.ExecContext(ctx, `
INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email, status)
SELECT ?, COALESCE(u.business_id, 0), u.id,
       COALESCE(NULLIF(SUBSTRING_INDEX(u.full_name, ' ', 1), ''), u.email),
       CASE WHEN LOCATE(' ', u.full_name) > 0 THEN SUBSTRING(u.full_name, LOCATE(' ', u.full_name) + 1) ELSE '' END,
       u.email, 'active'
FROM users u
WHERE u.id = ?
  AND NOT EXISTS (SELECT 1 FROM hr_employees e WHERE e.user_id = u.id)`,
		ownerID,
		ownerID,
	)

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id int64, fullName, email string) error {
	query, args, err := sq.
		Update("users").
		Set("full_name", fullName).
		Set("email", email).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) BanUser(ctx context.Context, id int64, reason string) error {
	query, args, err := sq.
		Update("users").
		Set("is_banned", true).
		Set("ban_reason", reason).
		Set("banned_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) UnbanUser(ctx context.Context, id int64) error {
	query, args, err := sq.
		Update("users").
		Set("is_banned", false).
		Set("ban_reason", nil).
		Set("banned_at", nil).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error {
	query, args, err := sq.
		Update("users").
		Set("is_console_admin", isAdmin).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Businesses CRUD ───────────────────────────────────────────────────────────

func (r *Repository) ListBusinesses(
	ctx context.Context,
	limit, offset int,
	id int64,
	name string,
	ownerID int64,
) ([]BusinessRow, int64, error) {
	conds := sq.And{}
	if id > 0 {
		conds = append(conds, sq.Eq{"id": id})
	}
	if name != "" {
		conds = append(conds, sq.Eq{"name": name})
	}
	if ownerID > 0 {
		conds = append(conds, sq.Eq{"owner_id": ownerID})
	}

	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("businesses").
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
		Select("id", "name", "owner_id", "created_at").
		From("businesses").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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

	businesses := make([]BusinessRow, 0, limit)
	for rows.Next() {
		var b BusinessRow
		if err := rows.Scan(&b.ID, &b.Name, &b.OwnerID, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		businesses = append(businesses, b)
	}
	return businesses, total, rows.Err()
}

func (r *Repository) GetBusinessByID(ctx context.Context, id int64) (*BusinessRow, error) {
	query, args, err := sq.
		Select("id", "name", "owner_id", "created_at").
		From("businesses").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var b BusinessRow
	if err := row.Scan(&b.ID, &b.Name, &b.OwnerID, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) GetActiveBusinessByOwner(ctx context.Context, ownerID int64) (*BusinessRow, error) {
	query, args, err := sq.
		Select("id", "name", "owner_id", "created_at").
		From("businesses").
		Where(sq.Eq{"owner_id": ownerID, "is_active": true}).
		OrderBy("created_at ASC").
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var b BusinessRow
	if err := row.Scan(&b.ID, &b.Name, &b.OwnerID, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) UpdateBusiness(ctx context.Context, id int64, name string) error {
	query, args, err := sq.
		Update("businesses").
		Set("name", name).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteBusiness(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("businesses").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) CountActiveSeatsForOwner(ctx context.Context, ownerID int64) (int, error) {
	const query = `
SELECT COUNT(DISTINCT user_id)
FROM (
	SELECT b.owner_id AS user_id
	FROM businesses b
	JOIN users u ON u.id = b.owner_id
	WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
	UNION
	SELECT p.user_id AS user_id
	FROM user_app_permissions p
	JOIN users u ON u.id = p.user_id
	WHERE p.owner_id = ? AND COALESCE(p.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
	UNION
	SELECT u.id AS user_id
	FROM users u
	JOIN businesses b ON b.id = u.business_id
	WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
) active_users`

	var count int
	if err := r.pool.QueryRowContext(ctx, query, ownerID, ownerID, ownerID).Scan(&count); err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return count, nil
}

func (r *Repository) decorateSubscriptionSeats(ctx context.Context, sub *SubscriptionRow) error {
	activeSeats, err := r.CountActiveSeatsForOwner(ctx, sub.OwnerID)
	if err != nil {
		return err
	}
	sub.ActiveSeatCount = activeSeats
	sub.IncludedSeatCount = activeSeats
	if sub.IncludedSeatCount > 10 {
		sub.IncludedSeatCount = 10
	}
	sub.BillableSeatCount = activeSeats - sub.IncludedSeatCount
	return nil
}

// ── Subscriptions CRUD ────────────────────────────────────────────────────────

func (r *Repository) ListSubscriptions(ctx context.Context, limit, offset int, status string) ([]SubscriptionRow, int64, error) {
	conds := sq.And{}
	if status != "" {
		conds = append(conds, sq.Eq{"s.status": status})
	}

	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
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
		Select("s.id", "s.owner_id", "COALESCE(u.full_name, '')", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.next_billing_date", "s.current_period_ends_at", "s.paddle_subscription_id", "s.paddle_customer_id",
			"s.created_at", "s.updated_at", "u.email", "b.id", "b.name").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		LeftJoin("businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE").
		Where(conds).
		OrderBy("s.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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

	subs := make([]SubscriptionRow, 0, limit)
	for rows.Next() {
		var s SubscriptionRow
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.OwnerName, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
			&s.TrialEndsAt, &s.NextBillingDate, &s.CurrentPeriodEndsAt, &s.PaddleSubscriptionID, &s.PaddleCustomerID,
			&s.CreatedAt, &s.UpdatedAt, &s.OwnerEmail, &s.BusinessID, &s.BusinessName); err != nil {
			return nil, 0, err
		}
		if err := r.decorateSubscriptionSeats(ctx, &s); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, total, rows.Err()
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id int64) (*SubscriptionRow, error) {
	query, args, err := sq.
		Select("s.id", "s.owner_id", "COALESCE(u.full_name, '')", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.next_billing_date", "s.current_period_ends_at", "s.paddle_subscription_id", "s.paddle_customer_id",
			"s.created_at", "s.updated_at", "u.email", "b.id", "b.name").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		LeftJoin("businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE").
		Where(sq.Eq{"s.id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var s SubscriptionRow
	if err := row.Scan(&s.ID, &s.OwnerID, &s.OwnerName, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
		&s.TrialEndsAt, &s.NextBillingDate, &s.CurrentPeriodEndsAt, &s.PaddleSubscriptionID, &s.PaddleCustomerID,
		&s.CreatedAt, &s.UpdatedAt, &s.OwnerEmail, &s.BusinessID, &s.BusinessName); err != nil {
		return nil, err
	}
	if err := r.decorateSubscriptionSeats(ctx, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionRow, error) {
	query, args, err := sq.
		Select("s.id", "s.owner_id", "COALESCE(u.full_name, '')", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.next_billing_date", "s.current_period_ends_at", "s.paddle_subscription_id", "s.paddle_customer_id",
			"s.created_at", "s.updated_at", "u.email", "b.id", "b.name").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		LeftJoin("businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE").
		Where(sq.Eq{"s.owner_id": ownerID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var s SubscriptionRow
	if err := row.Scan(&s.ID, &s.OwnerID, &s.OwnerName, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
		&s.TrialEndsAt, &s.NextBillingDate, &s.CurrentPeriodEndsAt, &s.PaddleSubscriptionID, &s.PaddleCustomerID,
		&s.CreatedAt, &s.UpdatedAt, &s.OwnerEmail, &s.BusinessID, &s.BusinessName); err != nil {
		return nil, err
	}
	if err := r.decorateSubscriptionSeats(ctx, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func nullableTimeValue(value sql.NullTime) any {
	if value.Valid {
		return value.Time
	}
	return nil
}

func nullableStringValue(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func (r *Repository) UpsertSubscription(ctx context.Context, params SubscriptionWriteParams) error {
	query, args, err := sq.
		Insert("subscriptions").
		Columns(
			"owner_id",
			"plan_id",
			"status",
			"seat_count",
			"trial_ends_at",
			"next_billing_date",
			"current_period_ends_at",
			"paddle_subscription_id",
			"paddle_customer_id",
		).
		Values(
			params.OwnerID,
			params.PlanID,
			params.Status,
			params.SeatCount,
			nullableTimeValue(params.TrialEndsAt),
			nullableTimeValue(params.NextBillingDate),
			nullableTimeValue(params.CurrentPeriodEndsAt),
			nullableStringValue(params.PaddleSubscriptionID),
			nullableStringValue(params.PaddleCustomerID),
		).
		Suffix(`ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), status = VALUES(status),
		   seat_count = VALUES(seat_count), trial_ends_at = VALUES(trial_ends_at),
		   next_billing_date = VALUES(next_billing_date), current_period_ends_at = VALUES(current_period_ends_at),
		   paddle_subscription_id = VALUES(paddle_subscription_id), paddle_customer_id = VALUES(paddle_customer_id),
		   updated_at = NOW()`).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) UpdateSubscription(ctx context.Context, id int64, params SubscriptionWriteParams) error {
	query, args, err := sq.
		Update("subscriptions").
		Set("plan_id", params.PlanID).
		Set("status", params.Status).
		Set("seat_count", params.SeatCount).
		Set("trial_ends_at", nullableTimeValue(params.TrialEndsAt)).
		Set("next_billing_date", nullableTimeValue(params.NextBillingDate)).
		Set("current_period_ends_at", nullableTimeValue(params.CurrentPeriodEndsAt)).
		Set("paddle_subscription_id", nullableStringValue(params.PaddleSubscriptionID)).
		Set("paddle_customer_id", nullableStringValue(params.PaddleCustomerID)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteSubscription(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Plans CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListPlans(ctx context.Context) ([]PlanRow, error) {
	query, args, err := sq.
		Select("id", "name", "slug", "price_per_seat_cents", "annual_price_per_seat_cents", "trial_days",
			"max_seats", "max_locations", "is_active", "is_solo", "is_featured", "tier_order", "tagline", "created_at").
		From("plans").
		OrderBy("tier_order ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []PlanRow
	for rows.Next() {
		var p PlanRow
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.PricePerSeatCents, &p.AnnualPricePerSeatCents,
			&p.TrialDays, &p.MaxSeats, &p.MaxLocations, &p.IsActive, &p.IsSolo,
			&p.IsFeatured, &p.TierOrder, &p.Tagline, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *Repository) UpdatePlan(ctx context.Context, id int64, name string, pricePerSeatCents, annualPricePerSeatCents int, isActive bool) error {
	query, args, err := sq.
		Update("plans").
		Set("name", name).
		Set("price_per_seat_cents", pricePerSeatCents).
		Set("annual_price_per_seat_cents", annualPricePerSeatCents).
		Set("is_active", isActive).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Manual Access Grants CRUD ─────────────────────────────────────────────────

func (r *Repository) ListGrants(ctx context.Context, limit, offset int) ([]GrantRow, int64, error) {
	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("manual_access_grants").
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
		Select("g.id", "g.user_id", "COALESCE(u.full_name, '')", "g.plan_id", "p.name", "g.granted_by", "adm.full_name",
			"g.note", "g.expires_at", "g.revoked_at", "g.created_at", "u.email", "b.id", "b.name").
		From("manual_access_grants g").
		Join("plans p ON p.id = g.plan_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.owner_id = g.user_id AND COALESCE(b.is_active, TRUE) = TRUE").
		LeftJoin("users adm ON adm.id = g.granted_by").
		OrderBy("g.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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

	grants := make([]GrantRow, 0, limit)
	for rows.Next() {
		var g GrantRow
		if err := rows.Scan(&g.ID, &g.UserID, &g.UserName, &g.PlanID, &g.PlanName, &g.GrantedBy, &g.GranterName,
			&g.Note, &g.ExpiresAt, &g.RevokedAt, &g.CreatedAt, &g.UserEmail, &g.BusinessID, &g.BusinessName); err != nil {
			return nil, 0, err
		}
		if g.BusinessID.Valid {
			seatCount, err := r.CountActiveSeatsForOwner(ctx, g.UserID)
			if err != nil {
				return nil, 0, err
			}
			g.ActiveSeatCount = seatCount
		}
		grants = append(grants, g)
	}
	return grants, total, rows.Err()
}

func (r *Repository) CreateGrant(ctx context.Context, userID, planID, grantedBy int64, note string, expiresAt *time.Time) (int64, error) {
	var expiresSQL interface{}
	if expiresAt != nil {
		expiresSQL = *expiresAt
	}
	query, args, err := sq.
		Insert("manual_access_grants").
		Columns("user_id", "plan_id", "granted_by", "note", "expires_at").
		Values(userID, planID, grantedBy, note, expiresSQL).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetGrantUserID(ctx context.Context, id int64) (int64, error) {
	query, args, err := sq.
		Select("user_id").
		From("manual_access_grants").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	var userID int64
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *Repository) RevokeGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Update("manual_access_grants").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("manual_access_grants").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Manual Addon Grants CRUD ─────────────────────────────────────────────────

func (r *Repository) ListAddons(ctx context.Context) ([]AddonCatalogRow, error) {
	query, args, err := sq.
		Select("id", "name", "description", "icon", "price", "is_free", "sort_order").
		From("addons").
		OrderBy("sort_order ASC", "id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addons := make([]AddonCatalogRow, 0)
	for rows.Next() {
		var a AddonCatalogRow
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon, &a.Price, &a.IsFree, &a.SortOrder); err != nil {
			return nil, err
		}
		addons = append(addons, a)
	}
	return addons, rows.Err()
}

func (r *Repository) AddonExists(ctx context.Context, addonID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*)").
		From("addons").
		Where(sq.Eq{"id": addonID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ListAddonGrants(ctx context.Context, limit, offset int) ([]AddonGrantRow, int64, error) {
	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("manual_addon_grants").
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
		Select(
			"g.id", "g.user_id", "COALESCE(u.full_name, '')", "u.email",
			"g.business_id", "COALESCE(b.name, '')",
			"g.addon_id", "a.name", "a.price",
			"g.granted_by", "adm.full_name", "g.note", "g.expires_at", "g.revoked_at",
			"g.enabled_by_grant", "g.created_at",
		).
		From("manual_addon_grants g").
		Join("addons a ON a.id = g.addon_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.id = g.business_id").
		LeftJoin("users adm ON adm.id = g.granted_by").
		OrderBy("g.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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

	grants := make([]AddonGrantRow, 0, limit)
	for rows.Next() {
		var g AddonGrantRow
		if err := rows.Scan(&g.ID, &g.UserID, &g.UserName, &g.UserEmail, &g.BusinessID, &g.BusinessName,
			&g.AddonID, &g.AddonName, &g.AddonPrice, &g.GrantedBy, &g.GranterName, &g.Note,
			&g.ExpiresAt, &g.RevokedAt, &g.EnabledByGrant, &g.CreatedAt); err != nil {
			return nil, 0, err
		}
		grants = append(grants, g)
	}
	return grants, total, rows.Err()
}

func (r *Repository) CreateAddonGrant(ctx context.Context, userID, businessID int64, addonID string, grantedBy int64, note string, expiresAt *time.Time, enabledByGrant bool) (int64, error) {
	var expiresSQL interface{}
	if expiresAt != nil {
		expiresSQL = *expiresAt
	}
	query, args, err := sq.
		Insert("manual_addon_grants").
		Columns("user_id", "business_id", "addon_id", "granted_by", "note", "expires_at", "enabled_by_grant").
		Values(userID, businessID, addonID, grantedBy, note, expiresSQL, enabledByGrant).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetAddonGrantScope(ctx context.Context, id int64) (*AddonGrantScopeRow, error) {
	query, args, err := sq.
		Select("user_id", "business_id", "addon_id", "enabled_by_grant").
		From("manual_addon_grants").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	var scope AddonGrantScopeRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&scope.UserID, &scope.BusinessID, &scope.AddonID, &scope.EnabledByGrant); err != nil {
		return nil, err
	}
	return &scope, nil
}

func (r *Repository) RevokeAddonGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Update("manual_addon_grants").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteAddonGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("manual_addon_grants").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) IsBusinessAddonEnabled(ctx context.Context, businessID int64, addonID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*)").
		From("business_addons").
		Where(sq.Eq{"business_id": businessID, "addon_id": addonID, "is_enabled": true}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) EnableBusinessAddon(ctx context.Context, businessID int64, addonID string) error {
	query := `
INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)
VALUES (?, ?, true, NOW())
ON DUPLICATE KEY UPDATE is_enabled = true, enabled_at = NOW(), disabled_at = NULL`
	_, err := r.pool.ExecContext(ctx, query, businessID, addonID)
	return err
}

func (r *Repository) DisableBusinessAddon(ctx context.Context, businessID int64, addonID string) error {
	query, args, err := sq.
		Update("business_addons").
		Set("is_enabled", false).
		Set("disabled_at", sq.Expr("NOW()")).
		Where(sq.Eq{"business_id": businessID, "addon_id": addonID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) HasActiveAddonGrant(ctx context.Context, businessID int64, addonID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*)").
		From("manual_addon_grants").
		Where(sq.Eq{"business_id": businessID, "addon_id": addonID, "revoked_at": nil}).
		Where("(expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// ── Stats ─────────────────────────────────────────────────────────────────────

func (r *Repository) GetStats(ctx context.Context) (*StatsRow, error) {
	query, args, err := sq.
		Select(
			"(SELECT COUNT(*) FROM users) AS total_users",
			"(SELECT COUNT(*) FROM users WHERE is_banned = TRUE) AS banned_users",
			"(SELECT COUNT(*) FROM businesses) AS total_businesses",
			"(SELECT COUNT(*) FROM subscriptions WHERE status = 'active') AS active_subs",
			"(SELECT COUNT(*) FROM subscriptions WHERE status = 'trial') AS trial_subs",
			"(SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL) AS manual_grants",
			"(SELECT COUNT(*) FROM manual_addon_grants WHERE revoked_at IS NULL) AS manual_addon_grants",
			"(SELECT COUNT(*) FROM users WHERE is_console_admin = TRUE) AS console_admins",
		).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}

	var s StatsRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&s.TotalUsers, &s.BannedUsers, &s.TotalBusinesses, &s.ActiveSubs, &s.TrialSubs, &s.ManualGrants, &s.ManualAddonGrants, &s.ConsoleAdmins,
	); err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	return &s, nil
}
