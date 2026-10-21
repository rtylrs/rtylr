package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

// managerPINVerifier is the slice of the repository VerifyPIN needs: a read of
// candidate hashes. Narrowing the dependency keeps the credential path
// testable without a database and stops it growing new reads by accident.
type managerPINVerifier interface {
	ListManagerPINHashes(ctx context.Context, businessID int64) ([]domain.ManagerPINCandidate, error)
}

// PinAttemptStore persists the failed-verification counter for one business.
// Implementations must be shared by every replica and outlive the process: an
// in-process map would hand an attacker a fresh budget on every restart and on
// every replica behind the load balancer.
type PinAttemptStore interface {
	// PinRetryAt reports the instant the next attempt is accepted for the
	// business, or the zero time when it is not backing off.
	PinRetryAt(ctx context.Context, businessID int64) (time.Time, error)
	// RecordPINFailure increments the counter and reports the wait it now
	// implies.
	RecordPINFailure(ctx context.Context, businessID int64) (time.Time, error)
	// ResetPINFailures clears the counter, unblocking a business once a manager
	// has typed their PIN correctly.
	ResetPINFailures(ctx context.Context, businessID int64) error
}

// checkManagerPINHash is the real credential comparison. It is a var only so
// tests can count how many candidate comparisons a path actually performed —
// the wrong-PIN path does this work instead of a dummy compare, and asserting
// that it happens at all is the difference between a timing-equalised endpoint
// and one that only looks equalised.
var checkManagerPINHash = func(pin, hash string) bool {
	return passwordutils.CheckPINHash(pin, hash)
}

// burnManagerPINCompare is the timing equaliser: a cost-12 bcrypt comparison
// against a fixed dummy hash, run on every path that performs fewer real
// comparisons than a full candidate scan would.
//
// A successful match short-circuits the scan, so without this a match at
// candidate 0 would answer measurably faster than a match at candidate 7 and
// the latency would rank a business's managers. It is a var rather than a
// direct call only so tests can assert that every path really does burn one.
var burnManagerPINCompare = func(pin string) bool {
	return passwordutils.CheckDummyHash(pin)
}
