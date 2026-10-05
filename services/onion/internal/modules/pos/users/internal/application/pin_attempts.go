package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

// pinVerifier is the slice of the repository VerifyPIN needs. Narrowing the
// dependency keeps the credential path testable without a database and stops it
// from growing new reads by accident.
type pinVerifier interface {
	FindForVerify(ctx context.Context, businessID, locationID, userID int64) (*domain.POSUser, error)
	GetPINHash(ctx context.Context, businessID, id int64) (string, error)
	GetAuthorizationVersion(ctx context.Context, businessID, userID int64) (int64, error)
}

// PinAttemptStore persists the failed-verification counter for one
// (business_id, location_id, user_id) key. Implementations must be shared by
// every replica and outlive the process: an in-process map would hand an
// attacker a fresh budget on every restart and on every replica behind the
// load balancer.
type PinAttemptStore interface {
	// PinRetryAt reports the instant the next attempt is accepted for the key,
	// or the zero time when the key is not backing off.
	PinRetryAt(ctx context.Context, businessID, locationID, userID int64) (time.Time, error)
	// RecordPINFailure increments the key's counter and reports the wait it now
	// implies.
	RecordPINFailure(ctx context.Context, businessID, locationID, userID int64) (time.Time, error)
	// ResetPINFailures clears the counter, unblocking a successful operator or
	// a manager rotating a stuck operator's PIN.
	ResetPINFailures(ctx context.Context, businessID, locationID, userID int64) error
	ResetPINFailuresForPOSUser(ctx context.Context, businessID, posUserID int64) error
}

// burnPINCompare is the timing equaliser: a genuine cost-12 bcrypt comparison
// against a fixed dummy hash, run on every failure path — unknown operator,
// operator with no PIN set, wrong PIN, locked out, counter store unavailable.
// Those paths all return the same error to the caller, so they must also cost
// the same ~250ms; otherwise response latency enumerates which operator ids
// exist at a location and whether a given one is currently locked out.
//
// It is a var rather than a direct call only so tests can assert that every
// failure path really does perform a comparison.
var burnPINCompare = func(pin string) bool {
	return passwordutils.CheckDummyHash(pin)
}
