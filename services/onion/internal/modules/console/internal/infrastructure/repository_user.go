package infrastructure

import (
	"context"
	"database/sql"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

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

func nullableString(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return raw
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
