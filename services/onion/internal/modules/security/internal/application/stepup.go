package application

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrStepUpRequired means the caller holds an enrolled second factor but has
	// not proved it recently enough to change the account's authentication
	// posture.
	ErrStepUpRequired = errors.New("recent MFA verification required")
	// ErrStepUpUnavailable means the factor state could not be read. It is
	// separate from ErrStepUpRequired so the outage is visible in logs, but it
	// is denied identically: an unknown state is never a satisfied check.
	ErrStepUpUnavailable = errors.New("MFA factor state unavailable")
	// ErrMFAAttemptsExceeded means too many factor proofs were rejected inside
	// the attempt window.
	ErrMFAAttemptsExceeded = errors.New("too many MFA code attempts")
)

// stepUpWindow matches the platform's MFA-verified window so a user who has
// just cleared the MFA gate is not immediately asked to re-prove the same
// factor before reaching the settings screen.
const stepUpWindow = 5 * time.Minute

// requireStepUp is the single rule guarding every second-factor mutation: a
// user who already holds a factor must have proved it within stepUpWindow.
//
// A user with no factor is not required to prove anything, because there is
// nothing to rotate, disable or remove — and demanding a proof there would make
// it impossible for any account to ever enroll one. First enrollment is
// self-proving instead: enroll/confirm requires a TOTP derived from the secret
// that enrollment just issued.
//
// Every failure path denies. A state read that errors, a missing provisioning
// row, a missing timestamp, a stale timestamp and a timestamp in the future all
// mean "not proved", never "skip the check".
func (s *Service) requireStepUp(ctx context.Context, userID int64) error {
	state, err := s.repo.GetMFAFactorState(ctx, userID)
	if err != nil {
		return errors.Join(ErrStepUpUnavailable, err)
	}
	if !state.FactorEnrolled {
		return nil
	}
	if state.LastVerifiedAt == nil {
		return ErrStepUpRequired
	}
	age := time.Since(*state.LastVerifiedAt)
	if age < 0 || age > stepUpWindow {
		return ErrStepUpRequired
	}
	return nil
}

const (
	mfaAttemptLimit  = 5
	mfaAttemptWindow = time.Minute
	// maxTrackedMFAUsers bounds the map. Sweeping on overflow keeps the limiter
	// from growing with the number of distinct callers that ever guessed a code.
	maxTrackedMFAUsers = 10000
)

// mfaAttemptLimiter counts rejected factor proofs per user. A 6-digit TOTP
// with a +/-1 step window is ~3% per 30 seconds to guess, and a success opens
// a five-minute verified window, so a single unbounded route hands an attacker
// the whole second factor.
//
// It is process-local: with N replicas a caller gets N * mfaAttemptLimit
// guesses per window. That is why /hub/security/v1/mfa/verify still needs a
// per-route limit in rate_limit.go, which is the only shared, IP-keyed layer
// the module cannot reach.
type mfaAttemptLimiter struct {
	mu       sync.Mutex
	failures map[int64][]time.Time
}

var mfaAttempts = &mfaAttemptLimiter{failures: make(map[int64][]time.Time)}

func (l *mfaAttemptLimiter) check(userID int64, now time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.recentLocked(userID, now)) >= mfaAttemptLimit {
		return ErrMFAAttemptsExceeded
	}
	return nil
}

func (l *mfaAttemptLimiter) recordFailure(userID int64, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := append(l.recentLocked(userID, now), now)
	if len(kept) > mfaAttemptLimit {
		kept = kept[len(kept)-mfaAttemptLimit:]
	}
	l.failures[userID] = kept
	if len(l.failures) > maxTrackedMFAUsers {
		l.sweepLocked(now)
	}
}

func (l *mfaAttemptLimiter) reset(userID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, userID)
}

func (l *mfaAttemptLimiter) recentLocked(userID int64, now time.Time) []time.Time {
	kept := l.failures[userID][:0]
	for _, at := range l.failures[userID] {
		if at.After(now.Add(-mfaAttemptWindow)) {
			kept = append(kept, at)
		}
	}
	if len(kept) == 0 {
		delete(l.failures, userID)
	}
	return kept
}

func (l *mfaAttemptLimiter) sweepLocked(now time.Time) {
	cutoff := now.Add(-mfaAttemptWindow)
	for userID, times := range l.failures {
		fresh := times[:0]
		for _, at := range times {
			if at.After(cutoff) {
				fresh = append(fresh, at)
			}
		}
		if len(fresh) == 0 {
			delete(l.failures, userID)
		} else {
			l.failures[userID] = fresh
		}
	}
}
