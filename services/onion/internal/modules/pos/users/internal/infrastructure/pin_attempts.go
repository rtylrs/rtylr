package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
)

// PINAttemptStore persists failed PIN-verification counts per
// (business_id, location_id, user_id) so a lockout outlives a process restart
// and is shared by every replica.
type PINAttemptStore struct {
	pool *sql.DB
}

var errPINAttemptStoreUnavailable = errors.New("pos pin attempt store unavailable")

// NewPINAttemptStore wires the persistent failed-attempt counter.
func NewPINAttemptStore(db *sql.DB) *PINAttemptStore {
	return &PINAttemptStore{pool: db}
}

// PinRetryAt reports the instant the next attempt is accepted for the key, or
// the zero time when the key is not backing off.
func (s *PINAttemptStore) PinRetryAt(ctx context.Context, businessID, locationID, userID int64) (time.Time, error) {
	if s == nil || s.pool == nil {
		return time.Time{}, errPINAttemptStoreUnavailable
	}
	var (
		failures        int
		lastFailureUnix int64
	)
	err := s.pool.QueryRowContext(ctx, `
		SELECT failures, last_failure_at
		FROM pos_user_pin_attempts
		WHERE business_id = ? AND location_id = ? AND user_id = ?`,
		businessID, locationID, userID).Scan(&failures, &lastFailureUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return domain.PinRetryAt(failures, time.Unix(lastFailureUnix, 0).UTC()), nil
}

// RecordPINFailure increments the counter for the key and reports the wait it
// now implies. The increment is a single atomic upsert so concurrent requests
// against the same operator cannot lose a count.
func (s *PINAttemptStore) RecordPINFailure(ctx context.Context, businessID, locationID, userID int64) (time.Time, error) {
	if s == nil || s.pool == nil {
		return time.Time{}, errPINAttemptStoreUnavailable
	}
	now := time.Now().UTC().Unix()
	if _, err := s.pool.ExecContext(ctx, `
		INSERT INTO pos_user_pin_attempts (business_id, location_id, user_id, failures, last_failure_at)
		VALUES (?, ?, ?, 1, ?)
		ON DUPLICATE KEY UPDATE failures = failures + 1, last_failure_at = ?`,
		businessID, locationID, userID, now, now); err != nil {
		return time.Time{}, err
	}
	return s.PinRetryAt(ctx, businessID, locationID, userID)
}

// ResetPINFailures clears the counter for the key. Called after a successful
// verification and after a PIN rotation, so a successful operator and a
// manager resetting a stuck operator are both unstuck.
func (s *PINAttemptStore) ResetPINFailures(ctx context.Context, businessID, locationID, userID int64) error {
	if s == nil || s.pool == nil {
		return errPINAttemptStoreUnavailable
	}
	_, err := s.pool.ExecContext(ctx, `
		DELETE FROM pos_user_pin_attempts
		WHERE business_id = ? AND location_id = ? AND user_id = ?`,
		businessID, locationID, userID)
	return err
}

// ResetPINFailuresForPOSUser clears the counter addressed by a pos_users row
// id, which is what the manager-facing update path holds. Deleting through the
// join keeps the key derivation in one place instead of making callers look up
// user_id and location_id first.
func (s *PINAttemptStore) ResetPINFailuresForPOSUser(ctx context.Context, businessID, posUserID int64) error {
	if s == nil || s.pool == nil {
		return errPINAttemptStoreUnavailable
	}
	_, err := s.pool.ExecContext(ctx, `
		DELETE a FROM pos_user_pin_attempts a
		JOIN pos_users p
		  ON p.business_id = a.business_id
		 AND p.user_id = a.user_id
		 AND p.location_id = a.location_id
		WHERE p.business_id = ? AND p.id = ?`, businessID, posUserID)
	return err
}
