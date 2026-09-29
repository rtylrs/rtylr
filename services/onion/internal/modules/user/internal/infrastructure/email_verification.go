package infrastructure

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/config"
)

const (
	// The code is ten digits: 10^10 possibilities against the 10^6 the previous
	// six-digit code had. The old generator,
	// binary.BigEndian.Uint32(...)%1_000_000, was also biased — 2^32 is not a
	// multiple of 10^6, so the lowest (2^32 mod 10^6) values carried one extra
	// preimage each.
	emailVerificationCodeDigits = 10
	emailVerificationCodeSpace  = 10_000_000_000

	// Wrong codes tolerated per live code, and per target user, inside the
	// code's own ten-minute lifetime.
	//
	// The per-user cap is the one that defeats a sweep, because every distinct
	// guess is its own per-code key. The per-code cap is what stops one wrong
	// code being retried forever.
	emailVerificationCodeAttempts = 5
	emailVerificationUserAttempts = 5
	emailVerificationAttemptTTL   = 10 * time.Minute
)

var (
	// ErrInvalidEmailVerificationCode is a code that is not valid right now:
	// wrong, expired, already spent, or malformed. Deliberately the same
	// outcome for all of them so the endpoint is not an oracle.
	ErrInvalidEmailVerificationCode = errors.New("invalid or expired email verification code")

	// ErrSupersededEmailVerificationCode is a code in the pre-binding
	// six-digit format. Those rows are stored as an unkeyed sha256(code) with
	// no user in the digest, so honouring the format would keep both the
	// offline-recovery property and the cross-account confusion that the user
	// binding exists to remove — a guessed code would still spend whichever
	// account it happened to match. They get their own error so the caller is
	// told to request a new code instead of being told the code was wrong.
	ErrSupersededEmailVerificationCode = errors.New("email verification code was issued in an older format; request a new one")

	// ErrTooManyEmailVerificationAttempts is the attempt cap for a target
	// user, which is what stops an enumeration no matter how many source
	// addresses the attacker rotates through.
	ErrTooManyEmailVerificationAttempts = errors.New("too many email verification attempts")
)

// EmailVerification is what one issuance produced: Code is the number the email
// shows, Token is the self-describing value the API accepts.
type EmailVerification struct {
	UserID int64
	Code   string
	Token  string
}

// newEmailVerificationCode returns a uniformly random zero-padded code.
func newEmailVerificationCode() (string, error) {
	// Rejection sampling, because a plain modulo over 2^64 would give the
	// lowest (2^64 mod 10^10) values one extra preimage each.
	limit := ^uint64(0) - (^uint64(0) % emailVerificationCodeSpace)

	var buf [8]byte
	for {
		if _, err := rand.Read(buf[:]); err != nil {
			return "", err
		}
		if v := binary.BigEndian.Uint64(buf[:]); v < limit {
			return fmt.Sprintf("%0*d", emailVerificationCodeDigits, v%emailVerificationCodeSpace), nil
		}
	}
}

// emailVerificationSecret is the key the stored digest is MACed with. HMAC
// rather than a bare sha256 of the code: the table holds one row per pending
// verification, so an unkeyed digest would let anyone who obtains a copy of it
// (backup, replica, an over-broad SELECT grant) recover live codes by
// enumerating 10^10 values, which is hours of commodity GPU time and no
// brute force at all against a plain hash.
//
// JWT_SECRET is the deployment's existing high-entropy secret, already required
// by config validation for every runmode that serves these routes. Rotating it
// invalidates codes that are at most ten minutes old, which is the intended
// blast radius. Mirrors the MFA key resolution: explicit override first,
// JWTSecret second, and no silent unkeyed fallback.
func emailVerificationSecret() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_SECRET"))
	if secret == "" && config.GlobalConfig != nil {
		secret = strings.TrimSpace(config.GlobalConfig.JWTSecret)
	}
	if secret == "" {
		// Fail closed rather than hashing without a key, which would
		// reintroduce exactly the offline-recovery property the MAC removes.
		return nil, errors.New("no email verification secret configured")
	}

	return []byte(secret), nil
}

// emailVerificationHash binds a code to the account it was issued for. The user
// id is inside the MAC input, so the digest for (user A, code C) differs from
// the digest for (user B, code C) even for an identical code, and the
// verification query is additionally scoped by user_id. A code issued for one
// account therefore cannot spend a row belonging to another.
func emailVerificationHash(userID int64, code string) (string, error) {
	secret, err := emailVerificationSecret()
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, secret)
	// Domain-separated so this digest can never collide with another use of
	// the same secret, and ':'-delimited so no two (userID, code) splits hash
	// to the same input.
	fmt.Fprintf(mac, "email-verification:v2:%d:%s", userID, code)

	return hex.EncodeToString(mac.Sum(nil)), nil
}

// formatEmailVerificationToken builds the "<userID>.<code>" token the endpoint
// accepts. The user id rides along so the stored digest can be recomputed from
// the request alone — the endpoint is anonymous, so the binding cannot come from
// a session.
//
// Client consequence: the token is no longer a bare six-digit string, so a
// client that validates the typed value as /^\d{6}$/ rejects it. The auth web
// app's VerifyEmail input does exactly that and has to be widened to the token
// shape; the emailed link is unaffected.
func formatEmailVerificationToken(userID int64, code string) string {
	return strconv.FormatInt(userID, 10) + "." + code
}

// parseEmailVerificationToken splits "<userID>.<10 digits>" into its parts.
func parseEmailVerificationToken(token string) (int64, string, bool) {
	userPart, code, found := strings.Cut(strings.TrimSpace(token), ".")
	if !found {
		return 0, "", false
	}
	userID, err := strconv.ParseInt(userPart, 10, 64)
	if err != nil || userID <= 0 || len(code) != emailVerificationCodeDigits {
		return 0, "", false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return 0, "", false
		}
	}

	return userID, code, true
}

// isLegacyEmailVerificationCode reports whether a token is exactly the
// pre-binding six-digit format.
func isLegacyEmailVerificationCode(token string) bool {
	token = strings.TrimSpace(token)
	if len(token) != 6 {
		return false
	}
	for _, c := range token {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// verificationAttempts counts wrong email-verification codes, keyed by the
// target user and by the digest of the submitted code — never by the code
// itself, so a heap dump of this map yields no live credential.
//
// Keying by user rather than by IP is the point: the budget holds when the
// attacker rotates source addresses, which a per-IP budget does not. The state
// is in-process, so behind N replicas the attacker gets N times the budget —
// the same limitation the HTTP rate limiters have.
type verificationAttempts struct {
	mu        sync.Mutex
	byCode    map[string][]time.Time
	byUser    map[string][]time.Time
	lastSweep time.Time
}

var emailVerificationAttempts = &verificationAttempts{
	byCode: map[string][]time.Time{},
	byUser: map[string][]time.Time{},
}

// check reports whether another attempt against this user and code is within
// budget, without recording anything. A correct code presented while the budget
// is spent is still refused: past the cap the answer stops being about the
// code, and the caller is told to request a new one.
func (w *verificationAttempts) check(userKey, codeKey string) bool {
	now := time.Now()

	w.mu.Lock()
	defer w.mu.Unlock()

	w.sweepLocked(now, emailVerificationAttemptTTL)

	return len(w.byCode[codeKey]) < emailVerificationCodeAttempts &&
		len(w.byUser[userKey]) < emailVerificationUserAttempts
}

// recordFailure charges one wrong code against both budgets.
func (w *verificationAttempts) recordFailure(userKey, codeKey string) {
	now := time.Now()

	w.mu.Lock()
	defer w.mu.Unlock()

	w.byCode[codeKey] = append(w.byCode[codeKey], now)
	w.byUser[userKey] = append(w.byUser[userKey], now)
}

// resetUser clears the budget once a code verifies, and when a fresh code is
// issued. Asking for a new code is the intended recovery from a spent cap, so
// it must not inherit it.
func (w *verificationAttempts) resetUser(userKey string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.byUser, userKey)
	prefix := userKey + "|"
	for k := range w.byCode {
		if strings.HasPrefix(k, prefix) {
			delete(w.byCode, k)
		}
	}
}

// sweepLocked drops timestamps older than the window. Without it a guesser
// picking fresh user ids and codes would grow the maps without bound, since an
// entry is only pruned when that exact key is touched again.
func (w *verificationAttempts) sweepLocked(now time.Time, ttl time.Duration) {
	if !w.lastSweep.IsZero() && now.Sub(w.lastSweep) < ttl {
		return
	}
	w.lastSweep = now

	cutoff := now.Add(-ttl)
	for k, times := range w.byCode {
		if kept := keepAttemptsAfter(times, cutoff); len(kept) == 0 {
			delete(w.byCode, k)
		} else {
			w.byCode[k] = kept
		}
	}
	for k, times := range w.byUser {
		if kept := keepAttemptsAfter(times, cutoff); len(kept) == 0 {
			delete(w.byUser, k)
		} else {
			w.byUser[k] = kept
		}
	}
}

func keepAttemptsAfter(times []time.Time, cutoff time.Time) []time.Time {
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	return kept
}
