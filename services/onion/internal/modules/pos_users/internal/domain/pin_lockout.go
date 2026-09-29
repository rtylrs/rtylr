package domain

import "time"

const (
	// PinVerifyMaxFailures is how many consecutive failed verifications an
	// operator absorbs before the full lockout window opens. It is the whole
	// guess budget between lockouts, so the space of PINs an attacker can
	// cover is bounded by PinVerifyMaxFailures per PinLockoutDuration.
	PinVerifyMaxFailures = 5
	// PinLockoutDuration is the wait imposed once PinVerifyMaxFailures is hit,
	// and on every attempt after that until the counter is cleared.
	PinLockoutDuration = 15 * time.Minute

	pinBackoffBase = 2 * time.Second
	pinBackoffMax  = 30 * time.Second
)

// PinBackoff returns how long a key must wait after `failures` consecutive
// failed verifications.
//
// The first failure costs nothing — a mistyped PIN has to stay cheap for the
// operator standing at the till. After that the wait doubles per failure up to
// a ceiling, and PinVerifyMaxFailures opens the full lockout. The curve is
// monotonic, so an attacker cannot make progress by pacing attempts either.
func PinBackoff(failures int) time.Duration {
	if failures <= 1 {
		return 0
	}
	if failures >= PinVerifyMaxFailures {
		return PinLockoutDuration
	}
	wait := pinBackoffBase
	for i := 2; i < failures; i++ {
		wait *= 2
		if wait >= pinBackoffMax {
			return pinBackoffMax
		}
	}
	return wait
}

// PinRetryAt returns the instant the next attempt on a key becomes acceptable,
// given the recorded failure count and when the last failure landed. The zero
// time means the key is not backing off.
func PinRetryAt(failures int, lastFailureAt time.Time) time.Time {
	wait := PinBackoff(failures)
	if wait <= 0 || lastFailureAt.IsZero() {
		return time.Time{}
	}
	return lastFailureAt.Add(wait)
}
