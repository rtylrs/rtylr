package license

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// LicenseStatusActive / LicenseStatusRevoked mirror the self_host_licenses.status
// ENUM. The authority looks these up to honor revocation/expiry from the DB.
const (
	LicenseStatusActive  = "active"
	LicenseStatusRevoked = "revoked"
)

// ErrLicenseNotFound is returned by LookupByKeyHash when no row matches the hash.
var ErrLicenseNotFound = errors.New("license not found")

var (
	ErrSeatLimitExceeded = errors.New("license seat limit exceeded")
	ErrInstanceIDMissing = errors.New("instance id is required")
)

// LicenseRecord is the subset of self_host_licenses the authority needs to apply
// revocation and DB-side expiry on top of the signature/expiry verification.
//
//nolint:revive // Keep the exported name explicit for package consumers and tests.
type LicenseRecord struct {
	Status     string
	ExpiresAt  *time.Time
	GraceUntil *time.Time
}

// LookupByKeyHash fetches the revocation/expiry state for a license by its
// SHA-256 key hash (see HashKey). Returns ErrLicenseNotFound when absent so
// callers can fall back to the signature-based result.
func LookupByKeyHash(ctx context.Context, db *sql.DB, keyHash string) (*LicenseRecord, error) {
	if db == nil {
		return nil, sql.ErrConnDone
	}

	const query = `
		SELECT status, expires_at, grace_until
		FROM self_host_licenses
		WHERE key_hash = ?
		LIMIT 1
	`

	var (
		status     string
		expiresAt  sql.NullTime
		graceUntil sql.NullTime
	)
	err := db.QueryRowContext(ctx, query, keyHash).Scan(&status, &expiresAt, &graceUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}

	record := &LicenseRecord{Status: status}
	if expiresAt.Valid {
		t := expiresAt.Time
		record.ExpiresAt = &t
	}
	if graceUntil.Valid {
		t := graceUntil.Time
		record.GraceUntil = &t
	}
	return record, nil
}

// ApplyRevocation overlays DB revocation/expiry state onto a signature-valid
// StatusReport. A revoked row forces an invalid "revoked" report; a row expired
// past its grace window forces an "expired" report. Otherwise the input report
// is returned unchanged. Callers only invoke this for signature-valid reports.
func ApplyRevocation(report StatusReport, record *LicenseRecord, now time.Time) StatusReport {
	if record == nil {
		return report
	}

	if record.Status == LicenseStatusRevoked {
		report.Valid = false
		report.Status = StatusRevoked
		report.Reason = "license has been revoked"
		return report
	}

	expiresAt := record.ExpiresAt
	graceUntil := record.GraceUntil
	if expiresAt != nil && now.After(*expiresAt) {
		if graceUntil == nil || now.After(*graceUntil) {
			report.Valid = false
			report.Status = StatusExpired
			report.Reason = ErrExpired.Error()
		}
	}
	return report
}

// EnsureActivationTable creates the small authority-side activation ledger.
// It is intentionally initialized by the license service because this table
// belongs to the authority database, not to an application module schema.
func EnsureActivationTable(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return sql.ErrConnDone
	}
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS self_host_license_activations (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			key_hash CHAR(64) NOT NULL,
			instance_id VARCHAR(255) NOT NULL,
			activated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uq_license_activation (key_hash, instance_id),
			KEY idx_license_activation_key (key_hash)
		)`)
	return err
}

// Activate records an instance once and enforces the signed seat limit. A
// repeated activation of the same instance is idempotent.
func Activate(ctx context.Context, db *sql.DB, keyHash, instanceID string, seatLimit int) error {
	if db == nil {
		return sql.ErrConnDone
	}
	if instanceID == "" {
		return ErrInstanceIDMissing
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existing int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM self_host_license_activations
		WHERE key_hash = ? AND instance_id = ?`, keyHash, instanceID).Scan(&existing)
	if err != nil {
		return err
	}
	if existing > 0 {
		return tx.Commit()
	}

	if seatLimit > 0 {
		var count int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM self_host_license_activations
			WHERE key_hash = ? FOR UPDATE`, keyHash).Scan(&count); err != nil {
			return err
		}
		if count >= seatLimit {
			return ErrSeatLimitExceeded
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO self_host_license_activations (key_hash, instance_id)
		VALUES (?, ?)`, keyHash, instanceID); err != nil {
		return err
	}
	return tx.Commit()
}
