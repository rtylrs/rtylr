package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure/persistence/sqlc"
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
	queries *sqlc.Queries
	pool    *sql.DB
}

var ErrResourceNotAllowed = errors.New("console resource is not allowlisted")

// Application-owned schema prefixes are intentionally narrow; system tables
// and unprefixed tables must be added explicitly to the exact-name set below.
var allowedResourceTablePrefixes = []string{
	"booking_", "careers_", "crm_", "erp_", "finance_", "flow_", "hr_",
	"oauth_", "pos_", "store_", "support_", "voxire_", "rtylr_",
}

var allowedResourceTables = map[string]struct{}{
	"access_requests": {}, "activity_log": {}, "addons": {}, "businesses": {},
	"chat_sessions": {}, "clients": {}, "config": {}, "conversation_messages": {},
	"conversation_summaries": {}, "custom_domains": {}, "email_jobs": {},
	"expenses": {}, "faq": {}, "feature_flags": {}, "users": {},
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
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

// dynamic-query: squirrel
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

	// cross-schema write to users (user module) — create the owner account during provisioning
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

	// cross-schema write to businesses (config/business module) — create the business during provisioning
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

	// cross-schema write to users (user module) — link the owner account to its new business
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
		// cross-schema read from plans (subscriptions module) — plan app entitlements for provisioning
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

	// cross-schema write to user_app_permissions (team/hub module) — grant owner role on the new business
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

		// cross-schema write to subscriptions (subscriptions module) — grant the plan during provisioning
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
			// cross-schema read from subscriptions (subscriptions module) — resolve existing subscription ID after upsert
			if scanErr := tx.QueryRowContext(ctx, "SELECT id FROM subscriptions WHERE owner_id = ? LIMIT 1", ownerID).Scan(&existingID); scanErr != nil {
				_ = tx.Rollback()
				return nil, scanErr
			}
			result.SubscriptionID = sql.NullInt64{Int64: existingID, Valid: true}
		}
	}

	// cross-schema write to hr_employees (hr_core module), reading from users (user module) — seed the owner's employee record
	if _, err = tx.ExecContext(ctx, `
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
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

// dynamic-query: squirrel
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
	return r.queries.DeleteUser(ctx, id)
}

func (r *Repository) BanUser(ctx context.Context, id int64, reason string) error {
	return r.queries.BanUser(ctx, sqlc.BanUserParams{
		BanReason: sql.NullString{String: reason, Valid: true},
		ID:        id,
	})
}

func (r *Repository) UnbanUser(ctx context.Context, id int64) error {
	return r.queries.UnbanUser(ctx, id)
}

func (r *Repository) SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error {
	return r.queries.SetConsoleAdmin(ctx, sqlc.SetConsoleAdminParams{
		IsConsoleAdmin: isAdmin,
		ID:             id,
	})
}

// ── Businesses CRUD ───────────────────────────────────────────────────────────

// dynamic-query: squirrel
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
	row, err := r.queries.GetBusinessByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return businessFromRow(row.ID, row.Name, row.OwnerID, row.CreatedAt), nil
}

func (r *Repository) GetActiveBusinessByOwner(ctx context.Context, ownerID int64) (*BusinessRow, error) {
	row, err := r.queries.GetActiveBusinessByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return businessFromRow(row.ID, row.Name, row.OwnerID, row.CreatedAt), nil
}

func (r *Repository) UpdateBusiness(ctx context.Context, id int64, name string) error {
	return r.queries.UpdateBusiness(ctx, sqlc.UpdateBusinessParams{
		Name: name,
		ID:   id,
	})
}

func (r *Repository) DeleteBusiness(ctx context.Context, id int64) error {
	return r.queries.DeleteBusiness(ctx, id)
}

func (r *Repository) CountActiveSeatsForOwner(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountActiveSeatsForOwner(ctx, sqlc.CountActiveSeatsForOwnerParams{
		OwnerID:   ownerID,
		OwnerID_2: ownerID,
		OwnerID_3: ownerID,
	})
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return int(count), nil
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

// dynamic-query: squirrel
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
	row, err := r.queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	sub := subscriptionFromIDRow(row)
	if err := r.decorateSubscriptionSeats(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (r *Repository) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionRow, error) {
	row, err := r.queries.GetSubscriptionByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	sub := subscriptionFromOwnerRow(row)
	if err := r.decorateSubscriptionSeats(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
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
	return r.queries.UpsertSubscription(ctx, upsertSubscriptionParams(params))
}

// dynamic-query: squirrel
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
	return r.queries.DeleteSubscription(ctx, id)
}

// ── Plans CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListPlans(ctx context.Context) ([]PlanRow, error) {
	rows, err := r.queries.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	plans := make([]PlanRow, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, planFromListRow(row))
	}
	return plans, nil
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

// dynamic-query: squirrel
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

// GetGrant returns the display fields for one grant without loading the full
// grant history. CreateGrant uses it to build its response efficiently.
func (r *Repository) GetGrant(ctx context.Context, id int64) (*GrantRow, error) {
	query, args, err := sq.
		Select("g.id", "g.user_id", "COALESCE(u.full_name, '')", "g.plan_id", "p.name", "g.granted_by", "adm.full_name",
			"g.note", "g.expires_at", "g.revoked_at", "g.created_at", "u.email", "b.id", "b.name").
		From("manual_access_grants g").
		Join("plans p ON p.id = g.plan_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.owner_id = g.user_id AND COALESCE(b.is_active, TRUE) = TRUE").
		LeftJoin("users adm ON adm.id = g.granted_by").
		Where(sq.Eq{"g.id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	var grant GrantRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&grant.ID, &grant.UserID, &grant.UserName, &grant.PlanID, &grant.PlanName, &grant.GrantedBy, &grant.GranterName,
		&grant.Note, &grant.ExpiresAt, &grant.RevokedAt, &grant.CreatedAt, &grant.UserEmail, &grant.BusinessID, &grant.BusinessName,
	); err != nil {
		return nil, err
	}
	if grant.BusinessID.Valid {
		seatCount, err := r.CountActiveSeatsForOwner(ctx, grant.UserID)
		if err != nil {
			return nil, err
		}
		grant.ActiveSeatCount = seatCount
	}
	return &grant, nil
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

// dynamic-query: squirrel
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

// ── Stats ─────────────────────────────────────────────────────────────────────

func (r *Repository) GetStats(ctx context.Context) (*StatsRow, error) {
	row, err := r.queries.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	return &StatsRow{
		TotalUsers:        row.TotalUsers,
		BannedUsers:       row.BannedUsers,
		TotalBusinesses:   row.TotalBusinesses,
		ActiveSubs:        row.ActiveSubs,
		TrialSubs:         row.TrialSubs,
		ManualGrants:      row.ManualGrants,
		ManualAddonGrants: row.ManualAddonGrants,
		ConsoleAdmins:     row.ConsoleAdmins,
	}, nil
}
