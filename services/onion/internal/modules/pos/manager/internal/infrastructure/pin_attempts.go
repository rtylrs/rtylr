package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
)

// PINAttemptStore persists failed manager-PIN counts per business so a lockout
// outlives a process restart and is shared by every replica.
//
// The key is the business and not (business, manager) because VerifyPINRequest
// carries only a PIN: the caller never names the manager, so there is no
// manager identity to key on before the match succeeds. Keying per candidate
// would be worse than useless — the key would be derived from the guess, so an
// attacker walking the PIN space would get a fresh budget on every candidate
// and the lockout would never engage. Keying per business is what actually
// bounds the guesses, because the guess is against the union of every manager
// PIN at that business.
//
// The cost is that one caller can spend the whole business's budget, disabling
// manager override for PinLockoutDuration. It does not affect logins, orders or
// anything else, and the API gate restricts who can spend it to a caller who
// already holds pos:orders:view at that business.
type PINAttemptStore struct {
	pool *sql.DB
}

var errPINAttemptStoreUnavailable = errors.New("pos manager pin attempt store unavailable")

// NewPINAttemptStore wires the persistent failed-attempt counter.
func NewPINAttemptStore(db *sql.DB) *PINAttemptStore {
	return &PINAttemptStore{pool: db}
}

// PinRetryAt reports the instant the next attempt is accepted for the business,
// or the zero time when it is not backing off.
func (s *PINAttemptStore) PinRetryAt(ctx context.Context, businessID int64) (time.Time, error) {
	if s == nil || s.pool == nil {
		return time.Time{}, errPINAttemptStoreUnavailable
	}
	var (
		failures        int
		lastFailureUnix int64
	)
	err := s.pool.QueryRowContext(ctx, `
		SELECT failures, last_failure_at
		FROM pos_manager_pin_attempts
		WHERE business_id = ?`,
		businessID).Scan(&failures, &lastFailureUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return domain.PinRetryAt(failures, time.Unix(lastFailureUnix, 0).UTC()), nil
}

// RecordPINFailure increments the counter for the business and reports the wait
// it now implies. The increment is a single atomic upsert so concurrent requests
// cannot lose a count.
func (s *PINAttemptStore) RecordPINFailure(ctx context.Context, businessID int64) (time.Time, error) {
	if s == nil || s.pool == nil {
		return time.Time{}, errPINAttemptStoreUnavailable
	}
	now := time.Now().UTC().Unix()
	if _, err := s.pool.ExecContext(ctx, `
		INSERT INTO pos_manager_pin_attempts (business_id, failures, last_failure_at)
		VALUES (?, 1, ?)
		ON DUPLICATE KEY UPDATE failures = failures + 1, last_failure_at = ?`,
		businessID, now, now); err != nil {
		return time.Time{}, err
	}
	return s.PinRetryAt(ctx, businessID)
}

// ResetPINFailures clears the counter, unblocking a business once a manager has
// typed their PIN correctly.
func (s *PINAttemptStore) ResetPINFailures(ctx context.Context, businessID int64) error {
	if s == nil || s.pool == nil {
		return errPINAttemptStoreUnavailable
	}
	_, err := s.pool.ExecContext(ctx, `
		DELETE FROM pos_manager_pin_attempts
		WHERE business_id = ?`,
		businessID)
	return err
}
