package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/searchutil"
)

var ErrSelfHostLicenseNotFound = errors.New("self-host license not found")

type SelfHostLicenseRow struct {
	ID            int64
	OwnerEmail    string
	CustomerName  sql.NullString
	Plan          string
	Features      sql.NullString
	SeatLimit     sql.NullInt64
	ExpiresAt     sql.NullTime
	GraceUntil    sql.NullTime
	Status        string
	LicenseKey    string
	KeyHash       string
	CreatedBy     sql.NullInt64
	UpdatedBy     sql.NullInt64
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
	RegeneratedAt sql.NullTime
	RevokedAt     sql.NullTime
}

type CreateSelfHostLicenseParams struct {
	OwnerEmail   string
	CustomerName string
	Plan         string
	FeaturesJSON string
	SeatLimit    int
	ExpiresAt    *time.Time
	GraceUntil   *time.Time
	LicenseKey   string
	KeyHash      string
	CreatedBy    int64
}

const selfHostLicenseColumns = "id, owner_email, customer_name, plan, features, seat_limit, expires_at, grace_until, status, license_key, key_hash, created_by, updated_by, created_at, updated_at, regenerated_at, revoked_at"

func scanSelfHostLicense(scanner interface {
	Scan(dest ...any) error
}) (*SelfHostLicenseRow, error) {
	var row SelfHostLicenseRow
	if err := scanner.Scan(
		&row.ID,
		&row.OwnerEmail,
		&row.CustomerName,
		&row.Plan,
		&row.Features,
		&row.SeatLimit,
		&row.ExpiresAt,
		&row.GraceUntil,
		&row.Status,
		&row.LicenseKey,
		&row.KeyHash,
		&row.CreatedBy,
		&row.UpdatedBy,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.RegeneratedAt,
		&row.RevokedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSelfHostLicenseNotFound
		}
		return nil, err
	}
	return &row, nil
}

func nullableString(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return raw
}

func nullableInt(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}

func (r *Repository) ListSelfHostLicenses(ctx context.Context, limit, offset int, email string) ([]SelfHostLicenseRow, int64, error) {
	listQ := sq.Select(selfHostLicenseColumns).
		From("self_host_licenses").
		OrderBy("created_at DESC", "id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
	countQ := sq.Select("COUNT(*)").From("self_host_licenses")

	if email = strings.TrimSpace(strings.ToLower(email)); email != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(email)
		if booleanQuery != "" {
			filter := sq.Expr(
				"(owner_email = ? OR MATCH(owner_email, customer_name) AGAINST (? IN BOOLEAN MODE))",
				email,
				booleanQuery,
			)
			listQ = listQ.Where(filter)
			countQ = countQ.Where(filter)
		} else {
			filter := sq.Eq{"owner_email": email}
			listQ = listQ.Where(filter)
			countQ = countQ.Where(filter)
		}
	}

	query, args, err := listQ.ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]SelfHostLicenseRow, 0, limit)
	for rows.Next() {
		row, err := scanSelfHostLicense(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQuery, countArgs, err := countQ.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.pool.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (r *Repository) GetSelfHostLicenseByID(ctx context.Context, id int64) (*SelfHostLicenseRow, error) {
	query, args, err := sq.Select(selfHostLicenseColumns).
		From("self_host_licenses").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanSelfHostLicense(r.pool.QueryRowContext(ctx, query, args...))
}

func (r *Repository) GetSelfHostLicenseByOwnerEmail(ctx context.Context, email string) (*SelfHostLicenseRow, error) {
	query, args, err := sq.Select(selfHostLicenseColumns).
		From("self_host_licenses").
		Where("LOWER(owner_email) = ?", strings.ToLower(strings.TrimSpace(email))).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanSelfHostLicense(r.pool.QueryRowContext(ctx, query, args...))
}

func (r *Repository) CreateSelfHostLicense(ctx context.Context, params CreateSelfHostLicenseParams) (int64, error) {
	query, args, err := sq.Insert("self_host_licenses").
		Columns(
			"owner_email",
			"customer_name",
			"plan",
			"features",
			"seat_limit",
			"expires_at",
			"grace_until",
			"license_key",
			"key_hash",
			"created_by",
			"updated_by",
		).
		Values(
			params.OwnerEmail,
			nullableString(params.CustomerName),
			params.Plan,
			nullableString(params.FeaturesJSON),
			nullableInt(params.SeatLimit),
			params.ExpiresAt,
			params.GraceUntil,
			params.LicenseKey,
			params.KeyHash,
			params.CreatedBy,
			params.CreatedBy,
		).
		ToSql()
	if err != nil {
		return 0, err
	}
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) RegenerateSelfHostLicenseKey(ctx context.Context, id int64, licenseKey, keyHash string, updatedBy int64) error {
	query, args, err := sq.Update("self_host_licenses").
		Set("license_key", licenseKey).
		Set("key_hash", keyHash).
		Set("status", "active").
		Set("updated_by", updatedBy).
		Set("regenerated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Set("revoked_at", nil).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrSelfHostLicenseNotFound
	}
	return nil
}

func (r *Repository) RevokeSelfHostLicense(ctx context.Context, id int64, updatedBy int64) error {
	query, args, err := sq.Update("self_host_licenses").
		Set("status", "revoked").
		Set("updated_by", updatedBy).
		Set("revoked_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrSelfHostLicenseNotFound
	}
	return nil
}
