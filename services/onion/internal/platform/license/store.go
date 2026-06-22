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

// LicenseRecord is the subset of self_host_licenses the authority needs to apply
// revocation and DB-side expiry on top of the signature/expiry verification.
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
