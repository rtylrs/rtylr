package domain

import "time"

// The manager PIN is the second factor for a POS manager override: a cashier
// holds a session, and a manager types their PIN at the till to approve a void,
// refund or drawer open. It therefore has to survive guessing, and the schedule
// below is the brake.
//
// The constants and the curve are deliberately identical to the pos_users POS
// PIN lockout (5 attempts, 2s -> 4s -> 8s, then a 15 minute lockout). Two POS
// PIN surfaces with two different budgets would mean the weaker one sets the
// effective strength of both, so these values must be changed together with
// pos_users/internal/domain/pin_lockout.go.
const (
	// PinVerifyMaxFailures is how many consecutive failed verifications a
	// business absorbs before the full lockout window opens. It is the whole
	// guess budget between lockouts.
	PinVerifyMaxFailures = 5
	// PinLockoutDuration is the wait imposed once PinVerifyMaxFailures is hit,
	// and on every attempt after that until the counter is cleared.
	PinLockoutDuration = 15 * time.Minute

	pinBackoffBase = 2 * time.Second
	pinBackoffMax  = 30 * time.Second
)

// PinBackoff returns how long a business must wait after `failures`
// consecutive failed manager-PIN verifications.
//
// The first failure costs nothing so a mistyped PIN stays cheap for the manager
// standing at the till. After that the wait doubles per failure up to a
// ceiling, and PinVerifyMaxFailures opens the full lockout. The curve is
// monotonic, so pacing attempts buys no extra guesses.
func PinBackoff(failures int) time.Duration {
	switch {
	case failures <= 1:
		return 0
	case failures >= PinVerifyMaxFailures:
		return PinLockoutDuration
	}
	// failures is 2..PinVerifyMaxFailures-1 here, so this walks 2s, 4s, 8s.
	wait := pinBackoffBase << (failures - 2)
	if wait > pinBackoffMax {
		return pinBackoffMax
	}
	return wait
}

// PinRetryAt returns the instant the next attempt on a business becomes
// acceptable, given its recorded failure count and when the last failure landed.
// The zero time means the business is not backing off.
func PinRetryAt(failures int, lastFailureAt time.Time) time.Time {
	wait := PinBackoff(failures)
	if wait <= 0 || lastFailureAt.IsZero() {
		return time.Time{}
	}
	return lastFailureAt.Add(wait)
}
