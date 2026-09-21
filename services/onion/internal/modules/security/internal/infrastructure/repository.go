package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/modules/security/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

func (r *Repository) GetSecuritySettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	// cross-schema read from users — 2FA flag is owned by the user/auth module
	twoFAEnabled, err := r.queries.GetTwoFactorEnabled(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SecuritySettings{
				TwoFactorEnabled: false,
				ActiveSessions:   []domain.ActiveSession{},
			}, nil
		}
		logger.Ctx(ctx).Error("Failed to get security settings", logger.Error(err))
		return domain.SecuritySettings{}, err
	}

	settings := domain.SecuritySettings{
		TwoFactorEnabled: twoFAEnabled,
		ActiveSessions:   []domain.ActiveSession{},
	}
	var required, enrolled bool
	stateErr := r.db.QueryRowContext(ctx, `
		SELECT mfa_required, mfa_enrolled
		FROM user_provisioning_states
		WHERE user_id = ?`, userID).Scan(&required, &enrolled)
	if stateErr == nil {
		settings.MFARequired = required
		settings.MFAEnrolled = enrolled
	} else if !authorizationSchemaUnavailable(stateErr) && !errors.Is(stateErr, sql.ErrNoRows) {
		return domain.SecuritySettings{}, stateErr
	}
	return settings, nil
}

func (r *Repository) SetTwoFactor(ctx context.Context, userID int64, enabled bool) error {
	// cross-schema write to users — two_factor_enabled is owned by the user/auth module.
	// Both flags must flip together: the auth gate reads mfa_required from
	// user_provisioning_states while users.two_factor_enabled drives the security
	// settings, so a second-statement failure must not leave them disagreeing.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		return cause
	}
	if err := sqlc.New(tx).SetTwoFactor(ctx, sqlc.SetTwoFactorParams{
		TwoFactorEnabled: enabled,
		ID:               userID,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to set 2FA", logger.Error(err))
		return rollback(err)
	}
	_, stateErr := tx.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET mfa_required = ?, updated_at = NOW()
		WHERE user_id = ?`, enabled, userID)
	if stateErr != nil && !authorizationSchemaUnavailable(stateErr) {
		return rollback(stateErr)
	}
	return tx.Commit()
}

func (r *Repository) BeginMFAEnrollment(ctx context.Context, userID int64, encryptedSecret []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		return cause
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user_mfa_enrollments
		WHERE user_id = ? AND method = 'totp' AND is_active = FALSE`, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_mfa_enrollments (user_id, method, secret_ciphertext, is_active)
		VALUES (?, 'totp', ?, FALSE)`, userID, encryptedSecret); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

func (r *Repository) GetMFASecret(ctx context.Context, userID int64, active bool) ([]byte, error) {
	var secret []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT secret_ciphertext
		FROM user_mfa_enrollments
		WHERE user_id = ? AND method = 'totp' AND is_active = ?
		ORDER BY id DESC LIMIT 1`, userID, active).Scan(&secret)
	return secret, err
}

func (r *Repository) CompleteMFAEnrollment(ctx context.Context, userID int64, recoveryHashes [][]byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		return cause
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE user_mfa_enrollments
		SET is_active = TRUE, enrolled_at = NOW()
		WHERE user_id = ? AND method = 'totp' AND is_active = FALSE`, userID)
	if err != nil {
		return rollback(err)
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		if err == nil {
			err = sql.ErrNoRows
		}
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user_mfa_enrollments SET is_active = FALSE, revoked_at = NOW() WHERE user_id = ? AND method = 'recovery_code' AND is_active = TRUE`, userID); err != nil {
		return rollback(err)
	}
	for _, hash := range recoveryHashes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_mfa_enrollments (user_id, method, recovery_code_hash, is_active)
			VALUES (?, 'recovery_code', ?, TRUE)`, userID, hash); err != nil {
			return rollback(err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET mfa_required = TRUE, mfa_enrolled = TRUE, last_mfa_verified_at = NOW(), updated_at = NOW()
		WHERE user_id = ?`, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET two_factor_enabled = TRUE WHERE id = ?`, userID); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

func (r *Repository) MarkMFAVerified(ctx context.Context, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET last_mfa_verified_at = NOW(), updated_at = NOW()
		WHERE user_id = ? AND mfa_required = TRUE AND mfa_enrolled = TRUE`, userID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}

type RecoveryCodeRow struct {
	ID   int64
	Hash string
}

func (r *Repository) ListRecoveryCodes(ctx context.Context, userID int64) ([]RecoveryCodeRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, recovery_code_hash
		FROM user_mfa_enrollments
		WHERE user_id = ? AND method = 'recovery_code' AND is_active = TRUE
		ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]RecoveryCodeRow, 0)
	for rows.Next() {
		var code RecoveryCodeRow
		if err := rows.Scan(&code.ID, &code.Hash); err != nil {
			return nil, err
		}
		result = append(result, code)
	}
	return result, rows.Err()
}

func (r *Repository) ConsumeRecoveryCode(ctx context.Context, userID, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE user_mfa_enrollments
		SET is_active = FALSE, revoked_at = NOW()
		WHERE id = ? AND user_id = ? AND method = 'recovery_code' AND is_active = TRUE`, id, userID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		_ = tx.Rollback()
		if err == nil {
			err = sql.ErrNoRows
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE user_provisioning_states SET last_mfa_verified_at = NOW(), updated_at = NOW() WHERE user_id = ?`, userID); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func authorizationSchemaUnavailable(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "user_provisioning_states") || strings.Contains(message, "user_mfa_enrollments")
}

func (r *Repository) GetLoginHistory(ctx context.Context, userID int64, limit, offset int) ([]domain.LoginEntry, int64, error) {
	total, err := r.queries.CountLoginHistory(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count login history", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListLoginHistory(ctx, sqlc.ListLoginHistoryParams{
		UserID: userID,
		Limit:  int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset: int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get login history", logger.Error(err))
		return nil, 0, err
	}

	entries := make([]domain.LoginEntry, 0, len(rows))
	for _, row := range rows {
		entry := domain.LoginEntry{
			ID:        row.ID,
			Email:     row.Email,
			Method:    row.Method,
			IP:        row.IpAddress,
			UserAgent: row.UserAgent,
		}
		if row.Success.Valid {
			entry.Success = row.Success.Bool
		}
		if row.CreatedAt.Valid {
			entry.CreatedAt = row.CreatedAt.Time
		}
		entries = append(entries, entry)
	}
	return entries, total, nil
}

func (r *Repository) InvalidateAllSessions(ctx context.Context, userID int64) error {
	// cross-schema write to user_sessions — sessions are owned by the auth module
	err := r.queries.InvalidateAllSessions(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to invalidate sessions", logger.Error(err))
	}
	return err
}

func (r *Repository) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error {
	err := r.queries.RecordLogin(ctx, sqlc.RecordLoginParams{
		UserID:    userID,
		Email:     email,
		Method:    method,
		IpAddress: sql.NullString{String: ip, Valid: ip != ""},
		UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		Success:   sql.NullBool{Bool: success, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record login", logger.Error(err))
	}
	return err
}
