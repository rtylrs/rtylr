package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/config"
)

// The attempt budget is process-global state, so every invocation needs its own
// user id or a -count=2 run inherits the first one's exhausted budget.
var emailVerificationTestUserSeq atomic.Uint64

func nextEmailVerificationTestUser() int64 {
	return 2_000_000 + int64(emailVerificationTestUserSeq.Add(1))
}

// The pre-binding code was fmt.Sprintf("%06d", Uint32%1_000_000) stored as
// sha256(code). A 10^6 space is exhaustible inside the ten-minute code TTL at
// the 1000 req/s global ceiling, and the digest carried no user, so a guessed
// code spent whichever row it happened to match.
func TestNewEmailVerificationCodeIsTenDigitsAndDistinct(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 500; i++ {
		code, err := newEmailVerificationCode()
		if err != nil {
			t.Fatalf("newEmailVerificationCode: %v", err)
		}
		if len(code) != emailVerificationCodeDigits {
			t.Fatalf("code %q is %d characters, want %d", code, len(code), emailVerificationCodeDigits)
		}
		if _, err := strconv.ParseUint(code, 10, 64); err != nil {
			t.Fatalf("code %q is not numeric: %v", code, err)
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("code %q was generated twice in 500 draws", code)
		}
		seen[code] = struct{}{}
	}
}

// A code issued for one account must not verify another, which is only
// possible if the user id is inside the digest.
func TestEmailVerificationHashIsBoundToTheUser(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	const code = "0123456789"
	hashA, err := emailVerificationHash(101, code)
	if err != nil {
		t.Fatalf("hash for user 101: %v", err)
	}
	hashB, err := emailVerificationHash(102, code)
	if err != nil {
		t.Fatalf("hash for user 102: %v", err)
	}
	if hashA == hashB {
		t.Fatal("the same code produced the same digest for two different users; a code issued for one account would verify the other")
	}

	// A separator-free concatenation of (user, code) would let (1, 12...) and
	// (11, 2...) collide; the domain-separated, colon-delimited input must not.
	hashOne, err := emailVerificationHash(1, "1234567890")
	if err != nil {
		t.Fatalf("hash for user 1: %v", err)
	}
	hashEleven, err := emailVerificationHash(11, "2345678901")
	if err != nil {
		t.Fatalf("hash for user 11: %v", err)
	}
	if hashOne == hashEleven {
		t.Fatal("adjacent (user, code) pairs collided")
	}
}

// Without a configured secret the issue path must fail rather than fall back to
// an unkeyed digest, which is the property the MAC exists to provide.
func TestEmailVerificationHashFailsClosedWithoutSecret(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "")
	t.Setenv("JWT_SECRET", "")

	previous := config.GlobalConfig
	config.GlobalConfig = &config.AppConfig{}
	t.Cleanup(func() { config.GlobalConfig = previous })

	if _, err := emailVerificationHash(1, "0123456789"); err == nil {
		t.Fatal("expected an error with no secret configured, got a digest")
	}
}

func TestParseEmailVerificationToken(t *testing.T) {
	for _, tc := range []struct {
		name       string
		token      string
		wantUser   int64
		wantCode   string
		wantParsed bool
	}{
		{name: "valid", token: "1042.0123456789", wantUser: 1042, wantCode: "0123456789", wantParsed: true},
		{name: "leading and trailing space", token: "  1042.0123456789\n", wantUser: 1042, wantCode: "0123456789", wantParsed: true},
		{name: "legacy six digits", token: "123456"},
		{name: "no separator", token: "0123456789"},
		{name: "short code", token: "1042.012345"},
		{name: "long code", token: "1042.01234567890"},
		{name: "non numeric code", token: "1042.01234567x9"},
		{name: "zero user", token: "0.0123456789"},
		{name: "negative user", token: "-7.0123456789"},
		{name: "non numeric user", token: "abc.0123456789"},
		{name: "empty", token: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			userID, code, ok := parseEmailVerificationToken(tc.token)
			if ok != tc.wantParsed {
				t.Fatalf("parsed = %v, want %v (user %d, code %q)", ok, tc.wantParsed, userID, code)
			}
			if !tc.wantParsed {
				return
			}
			if userID != tc.wantUser || code != tc.wantCode {
				t.Fatalf("got (%d, %q), want (%d, %q)", userID, code, tc.wantUser, tc.wantCode)
			}
		})
	}
}

func TestFormatAndParseRoundTrip(t *testing.T) {
	token := formatEmailVerificationToken(999_999_999, "0000000042")
	if token != "999999999.0000000042" {
		t.Fatalf("token = %q", token)
	}
	userID, code, ok := parseEmailVerificationToken(token)
	if !ok || userID != 999_999_999 || code != "0000000042" {
		t.Fatalf("round trip = (%d, %q, %v)", userID, code, ok)
	}
}

func TestIsLegacyEmailVerificationCode(t *testing.T) {
	for token, want := range map[string]bool{
		"123456":                     true,
		" 123456 ":                   true,
		"12345":                      false,
		"1234567":                    false,
		"1042.0123456789":            false,
		"abcdef":                     false,
		"":                           false,
		"1042.0123456":               false,
		"000000":                     true,
		"999999":                     true,
		"1042.0123456789\n":          false,
		"12345\n":                    false,
		"\t123456":                   true,
		"1.0123456789":               false,
		"1042.0123456789.and.more.1": false,
	} {
		if got := isLegacyEmailVerificationCode(token); got != want {
			t.Errorf("isLegacyEmailVerificationCode(%q) = %v, want %v", token, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Attempt budget
// ---------------------------------------------------------------------------

func TestVerificationAttemptsCapIsPerUserAndPerCode(t *testing.T) {
	w := &verificationAttempts{byCode: map[string][]time.Time{}, byUser: map[string][]time.Time{}}

	const userKey = "42"
	// A single wrong code cannot be retried forever.
	wrong := userKey + "|h1"
	for i := 0; i < emailVerificationCodeAttempts; i++ {
		w.recordFailure(userKey, wrong)
	}
	if w.check(userKey, wrong) {
		t.Fatalf("the same wrong code was still accepted after %d attempts", emailVerificationCodeAttempts)
	}
	// A different wrong code for the same user is also refused once the user's
	// budget is spent — this is the cap that survives an attacker rotating both
	// the code and the source address.
	if w.check(userKey, userKey+"|h2") {
		t.Fatalf("a fresh code for the same user was accepted after the per-user budget was spent")
	}
	// Another user is unaffected.
	if !w.check("43", "43|h3") {
		t.Fatal("the cap leaked across users")
	}
}

func TestVerificationAttemptsResetUserClearsBothBudgets(t *testing.T) {
	w := &verificationAttempts{byCode: map[string][]time.Time{}, byUser: map[string][]time.Time{}}

	const userKey = "42"
	for i := 0; i < emailVerificationCodeAttempts; i++ {
		w.recordFailure(userKey, userKey+"|h1")
	}
	w.recordFailure("43", "43|h9")

	w.resetUser(userKey)

	if !w.check(userKey, userKey+"|h1") {
		t.Fatal("resetUser did not restore the requesting user's budget")
	}
	if !w.check("43", "43|h9") {
		t.Fatal("resetUser cleared another user's budget")
	}
}

func TestVerificationAttemptsSweepDropsExpiredEntries(t *testing.T) {
	w := &verificationAttempts{byCode: map[string][]time.Time{}, byUser: map[string][]time.Time{}}

	stale := time.Now().Add(-2 * emailVerificationAttemptTTL)
	w.byCode["42|h1"] = []time.Time{stale}
	w.byUser["42"] = []time.Time{stale}
	w.byCode["42|h2"] = []time.Time{time.Now()}

	w.sweepLocked(time.Now(), emailVerificationAttemptTTL)

	if _, ok := w.byCode["42|h1"]; ok {
		t.Error("an expired per-code entry survived the sweep")
	}
	if _, ok := w.byUser["42"]; ok {
		t.Error("an expired per-user entry survived the sweep")
	}
	if _, ok := w.byCode["42|h2"]; !ok {
		t.Error("the sweep removed a live entry")
	}
}

func TestKeepAttemptsAfter(t *testing.T) {
	now := time.Now()
	times := []time.Time{now.Add(-2 * time.Hour), now.Add(-time.Minute), now}
	if got := len(keepAttemptsAfter(times, now.Add(-time.Hour))); got != 2 {
		t.Fatalf("kept %d timestamps, want 2", got)
	}
}

// ---------------------------------------------------------------------------
// Repository: the cross-user property, over a recording driver
// ---------------------------------------------------------------------------

// verificationProbe is a driver that records the verification SELECT and only
// answers for the exact (user_id, token_hash) pair it was primed with. A query
// for any other pair comes back with no rows, which is what a real table does
// when the code does not belong to that user.
type verificationProbe struct {
	mu sync.Mutex

	wantUser int64
	wantHash string

	// matchDigestOnly answers a one-argument (digest-only) lookup whose
	// argument equals this digest. It models the pre-fix query shape, which had
	// no user in the predicate at all.
	matchDigestOnly string

	selects   []probeQuery
	execs     []string
	committed bool
	rolled    bool
}

type probeQuery struct {
	query string
	args  []driver.Value
}

type verificationProbeConnector struct{ probe *verificationProbe }

func (c verificationProbeConnector) Connect(context.Context) (driver.Conn, error) {
	return &verificationProbeConn{probe: c.probe}, nil
}
func (c verificationProbeConnector) Driver() driver.Driver { return verificationProbeDriver{c.probe} }

type verificationProbeDriver struct{ probe *verificationProbe }

func (d verificationProbeDriver) Open(string) (driver.Conn, error) {
	return &verificationProbeConn{probe: d.probe}, nil
}

type verificationProbeConn struct{ probe *verificationProbe }

func (c *verificationProbeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported by the probe driver")
}
func (c *verificationProbeConn) Close() error              { return nil }
func (c *verificationProbeConn) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }
func (c *verificationProbeConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &verificationProbeTx{probe: c.probe}, nil
}

func (c *verificationProbeConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	p := c.probe
	p.mu.Lock()
	defer p.mu.Unlock()

	values := make([]driver.Value, 0, len(args))
	for _, a := range args {
		values = append(values, a.Value)
	}
	p.selects = append(p.selects, probeQuery{query: query, args: values})

	if len(values) == 1 {
		hash, _ := values[0].(string)
		if p.matchDigestOnly != "" && hash == p.matchDigestOnly {
			// The pre-fix query also selected user_id, so the returned row has
			// two columns.
			return &probeRows{id: 1, userID: p.wantUser, columns: 2}, nil
		}

		return &probeRows{}, nil
	}
	if len(values) != 2 {
		return &probeRows{}, nil
	}
	userID, okUser := values[0].(int64)
	hash, okHash := values[1].(string)
	if !okUser || !okHash || userID != p.wantUser || hash != p.wantHash {
		return &probeRows{}, nil
	}

	return &probeRows{id: 1}, nil
}

func (c *verificationProbeConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	p := c.probe
	p.mu.Lock()
	defer p.mu.Unlock()
	p.execs = append(p.execs, query)

	return driver.RowsAffected(1), nil
}

type verificationProbeTx struct{ probe *verificationProbe }

func (t *verificationProbeTx) Commit() error {
	t.probe.mu.Lock()
	defer t.probe.mu.Unlock()
	t.probe.committed = true

	return nil
}

func (t *verificationProbeTx) Rollback() error {
	t.probe.mu.Lock()
	defer t.probe.mu.Unlock()
	t.probe.rolled = true

	return nil
}

type probeRows struct {
	id     int64
	userID int64
	// columns is 1 for the current (id-only) query and 2 for the pre-fix shape
	// that also selected user_id.
	columns int
	served  bool
}

func (r *probeRows) Columns() []string {
	if r.columns == 2 {
		return []string{"id", "user_id"}
	}

	return []string{"id"}
}

func (r *probeRows) Close() error { return nil }

func (r *probeRows) Next(dest []driver.Value) error {
	if r.served || r.id == 0 {
		return io.EOF
	}
	r.served = true
	dest[0] = r.id
	if r.columns == 2 && len(dest) > 1 {
		dest[1] = r.userID
	}

	return nil
}

func newProbeDB(t *testing.T, probe *verificationProbe) *sql.DB {
	t.Helper()

	db := sql.OpenDB(verificationProbeConnector{probe: probe})
	t.Cleanup(func() { _ = db.Close() })

	return db
}

// A token whose code belongs to user B must not spend user B's row when it is
// presented as user A's token, and must not touch the provisioning tables at
// all. Before the binding, the lookup was by token_hash alone, so the same
// token would have verified the address it was not issued for.
func TestConsumeEmailVerificationCodeDoesNotCrossUsers(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	ownerID := nextEmailVerificationTestUser()
	otherID := nextEmailVerificationTestUser()
	const code = "0123456789"
	ownerHash, err := emailVerificationHash(ownerID, code)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	probe := &verificationProbe{wantUser: ownerID, wantHash: ownerHash}
	repo := NewRepository(newProbeDB(t, probe))

	// The attacker submits the victim's code under their own user id. The
	// digest is recomputed over (attackerID, code), which is not the stored
	// digest for (ownerID, code), so the row is not found.
	err = repo.ConsumeEmailVerificationCode(context.Background(), formatEmailVerificationToken(otherID, code))
	if !errors.Is(err, ErrInvalidEmailVerificationCode) {
		t.Fatalf("cross-user token error = %v, want ErrInvalidEmailVerificationCode", err)
	}
	if len(probe.execs) != 0 {
		t.Fatalf("a rejected cross-user token still ran %d write(s): %v", len(probe.execs), probe.execs)
	}
	if probe.committed {
		t.Fatal("a rejected cross-user token committed")
	}

	// The owner's own token verifies and does spend the row.
	if err := repo.ConsumeEmailVerificationCode(context.Background(), formatEmailVerificationToken(ownerID, code)); err != nil {
		t.Fatalf("owner's token: %v", err)
	}
	if !probe.committed {
		t.Fatal("the owner's own token did not commit")
	}
	joined := strings.Join(probe.execs, "\n")
	for _, want := range []string{"user_email_verification_tokens SET consumed_at", "user_provisioning_states", "authorization_business_members"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected a write touching %q, got:\n%s", want, joined)
		}
	}
}

// The verification query must be scoped by user id, not only by digest, so the
// row that gets spent is provably the one the code was issued for.
func TestConsumeEmailVerificationCodeScopesLookupByUser(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	userID := nextEmailVerificationTestUser()
	probe := &verificationProbe{}
	repo := NewRepository(newProbeDB(t, probe))

	_ = repo.ConsumeEmailVerificationCode(context.Background(), formatEmailVerificationToken(userID, "0000000001"))

	if len(probe.selects) == 0 {
		t.Fatal("no verification query was issued")
	}
	first := probe.selects[0]
	if !strings.Contains(first.query, "user_id = ?") {
		t.Fatalf("the verification query is not scoped by user id:\n%s", first.query)
	}
	if len(first.args) != 2 {
		t.Fatalf("verification query took %d args, want (user_id, token_hash)", len(first.args))
	}
	if got, ok := first.args[0].(int64); !ok || got != userID {
		t.Fatalf("first arg = %v, want user id %d", first.args[0], userID)
	}
	wantHash, err := emailVerificationHash(userID, "0000000001")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if got, _ := first.args[1].(string); got != wantHash {
		t.Fatalf("digest arg = %q, want the user-bound digest %q", got, wantHash)
	}
}

// A code minted before the binding is refused with its own error so the caller
// can be told to request a new one. It is never looked up by a digest-only
// match, which is the lookup that let a guess spend an unrelated account.
func TestConsumeEmailVerificationCodeRefusesLegacySixDigitCode(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	probe := &verificationProbe{}
	repo := NewRepository(newProbeDB(t, probe))

	err := repo.ConsumeEmailVerificationCode(context.Background(), "123456")
	if !errors.Is(err, ErrSupersededEmailVerificationCode) {
		t.Fatalf("legacy code error = %v, want ErrSupersededEmailVerificationCode", err)
	}
	if len(probe.selects) != 0 {
		t.Fatalf("a legacy code still queried the table %d time(s); the digest-only lookup is the vulnerability", len(probe.selects))
	}
	if len(probe.execs) != 0 {
		t.Fatalf("a legacy code ran %d write(s)", len(probe.execs))
	}
}

// Past the per-user cap the endpoint stops even answering, so a wrong code
// cannot be retried indefinitely behind the per-IP rate limit.
func TestConsumeEmailVerificationCodeStopsAtTheAttemptCap(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	userID := nextEmailVerificationTestUser()
	probe := &verificationProbe{}
	repo := NewRepository(newProbeDB(t, probe))

	for i := 0; i < emailVerificationUserAttempts; i++ {
		// A distinct wrong code each time, so only the per-user budget is
		// being charged.
		token := formatEmailVerificationToken(userID, fmt.Sprintf("%010d", i+1))
		if err := repo.ConsumeEmailVerificationCode(context.Background(), token); !errors.Is(err, ErrInvalidEmailVerificationCode) {
			t.Fatalf("attempt %d: error = %v, want ErrInvalidEmailVerificationCode", i+1, err)
		}
	}

	overCap := formatEmailVerificationToken(userID, "9999999999")
	if err := repo.ConsumeEmailVerificationCode(context.Background(), overCap); !errors.Is(err, ErrTooManyEmailVerificationAttempts) {
		t.Fatalf("past the cap: error = %v, want ErrTooManyEmailVerificationAttempts", err)
	}
	if len(probe.selects) != emailVerificationUserAttempts {
		t.Fatalf("the over-cap attempt still reached the database (%d queries)", len(probe.selects))
	}
}

// A code issued for a different user must not have charged this user's budget:
// the counters are keyed on the token's own user id, not on whatever row a
// digest happened to match.
func TestConsumeEmailVerificationCodeDoesNotSpendAnotherUsersBudget(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	victim := nextEmailVerificationTestUser()
	attacker := nextEmailVerificationTestUser()
	probe := &verificationProbe{}
	repo := NewRepository(newProbeDB(t, probe))

	// The attacker burns their own budget.
	for i := 0; i < emailVerificationUserAttempts; i++ {
		token := formatEmailVerificationToken(attacker, fmt.Sprintf("%010d", i+1))
		if err := repo.ConsumeEmailVerificationCode(context.Background(), token); !errors.Is(err, ErrInvalidEmailVerificationCode) {
			t.Fatalf("attacker attempt %d: error = %v, want ErrInvalidEmailVerificationCode", i+1, err)
		}
	}
	if err := repo.ConsumeEmailVerificationCode(context.Background(), formatEmailVerificationToken(attacker, "8888888888")); !errors.Is(err, ErrTooManyEmailVerificationAttempts) {
		t.Fatalf("attacker over cap: error = %v, want ErrTooManyEmailVerificationAttempts", err)
	}

	// The victim still has their whole budget.
	if err := repo.ConsumeEmailVerificationCode(context.Background(), formatEmailVerificationToken(victim, "1111111111")); !errors.Is(err, ErrInvalidEmailVerificationCode) {
		t.Fatalf("victim's first attempt: error = %v, want ErrInvalidEmailVerificationCode", err)
	}
}
