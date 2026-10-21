package application

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	testBusinessID int64 = 4242
	testValidPIN         = "123456"
)

// testClock is the single source of "now" for a scenario. Both the Service and
// the fake counter store read it, so a test can step past a backoff window
// without sleeping and without the two disagreeing about the time.
type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func newTestClock() *testClock {
	return &testClock{now: time.Now().UTC()}
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// memoryPinAttempts keeps its state outside the Service. That is deliberate:
// the restart test drops the Service and builds a fresh one over the same
// store, and it still sees the lockout — exactly the property an in-process map
// inside Service would fail.
type memoryPinAttempts struct {
	mu       sync.Mutex
	failures map[string]int
	last     map[string]time.Time
	readErr  error
	clock    *testClock
}

func newMemoryPinAttempts(clock *testClock) *memoryPinAttempts {
	return &memoryPinAttempts{
		failures: map[string]int{},
		last:     map[string]time.Time{},
		clock:    clock,
	}
}

func businessKey(businessID int64) string {
	return strconv.FormatInt(businessID, 10)
}

func (m *memoryPinAttempts) PinRetryAt(_ context.Context, businessID int64) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readErr != nil {
		return time.Time{}, m.readErr
	}
	key := businessKey(businessID)
	return domain.PinRetryAt(m.failures[key], m.last[key]), nil
}

func (m *memoryPinAttempts) RecordPINFailure(_ context.Context, businessID int64) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readErr != nil {
		return time.Time{}, m.readErr
	}
	key := businessKey(businessID)
	m.failures[key]++
	m.last[key] = m.clock.Now()
	return domain.PinRetryAt(m.failures[key], m.last[key]), nil
}

func (m *memoryPinAttempts) ResetPINFailures(_ context.Context, businessID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := businessKey(businessID)
	delete(m.failures, key)
	delete(m.last, key)
	return nil
}

func (m *memoryPinAttempts) failureCount(businessID int64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failures[businessKey(businessID)]
}

// stubVerifier answers the credential read. Hashes are compared for real, so a
// success here is a genuine verification rather than a stubbed outcome.
type stubVerifier struct {
	candidates []domain.ManagerPINCandidate
	readErr    error
	// reads counts the credential reads that reached this verifier. A
	// locked-out business must not get that far.
	reads int
	// askedBusiness records the business the read was scoped to.
	askedBusiness int64
}

func (s *stubVerifier) ListManagerPINHashes(_ context.Context, businessID int64) ([]domain.ManagerPINCandidate, error) {
	s.reads++
	s.askedBusiness = businessID
	if s.readErr != nil {
		return nil, s.readErr
	}
	return s.candidates, nil
}

// hashCache memoises bcrypt hashing. Hashing is the same cost as a comparison
// (~210ms), and the fixtures re-hash the same handful of PINs dozens of times,
// so without this the suite spends most of its time generating fixtures.
var (
	hashMu    sync.Mutex
	hashCache = map[string]string{}
)

func mustHash(t *testing.T, pin string) string {
	t.Helper()
	hashMu.Lock()
	cached, ok := hashCache[pin]
	hashMu.Unlock()
	if ok {
		return cached
	}
	h, err := passwordutils.HashPIN(pin)
	if err != nil {
		t.Fatalf("hash test PIN: %v", err)
	}
	hashMu.Lock()
	hashCache[pin] = h
	hashMu.Unlock()
	return h
}

// candidatesFor builds n manager candidates. The last one carries validPIN, so
// a matching guess costs the full scan and an early match can be placed
// explicitly by the caller.
func candidatesFor(t *testing.T, pins ...string) []domain.ManagerPINCandidate {
	t.Helper()
	out := make([]domain.ManagerPINCandidate, 0, len(pins))
	for i, pin := range pins {
		out = append(out, domain.ManagerPINCandidate{
			OwnerID: int64(1000 + i),
			StaffID: int64(2000 + i),
			PinHash: mustHash(t, pin),
		})
	}
	return out
}

func newScenario(verifier managerPINVerifier) (*Service, *memoryPinAttempts, *testClock) {
	clock := newTestClock()
	store := newMemoryPinAttempts(clock)
	return &Service{verifier: verifier, attempts: store, now: clock.Now}, store, clock
}

// countCompares installs counters over both comparison entry points and
// restores them when the test ends. Install it once per test and zero the
// counters between measurements — installing twice would nest the wrappers and
// make the earlier counters observe the later measurements too.
func countCompares(t *testing.T) (real, dummy *int) {
	t.Helper()
	realCount, dummyCount := 0, 0
	origCheck, origBurn := checkManagerPINHash, burnManagerPINCompare
	checkManagerPINHash = func(pin, hash string) bool {
		realCount++
		return origCheck(pin, hash)
	}
	burnManagerPINCompare = func(pin string) bool {
		dummyCount++
		return origBurn(pin)
	}
	t.Cleanup(func() { checkManagerPINHash, burnManagerPINCompare = origCheck, origBurn })
	return &realCount, &dummyCount
}

func zeroCounts(real, dummy *int) {
	*real, *dummy = 0, 0
}

// The Nth attempt opens the lockout: a wrong PIN is rejected PinVerifyMaxFailures
// times, and every attempt after that is refused without reaching the credential
// read. The clock is stepped past each backoff window so the test measures the
// attempt budget rather than the wall-clock wait.
func TestVerifyPIN_LocksOutOnNthAttempt(t *testing.T) {
	verifier := &stubVerifier{candidates: candidatesFor(t, "999111", "888222", "777333")}
	svc, store, clock := newScenario(verifier)
	ctx := context.Background()

	for attempt := 1; attempt <= domain.PinVerifyMaxFailures; attempt++ {
		clock.advance(domain.PinLockoutDuration)
		_, err := svc.VerifyPIN(ctx, testBusinessID, "000000")
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("attempt %d: expected ErrInvalidPIN, got %v", attempt, err)
		}
		if got := store.failureCount(testBusinessID); got != attempt {
			t.Fatalf("attempt %d: counter = %d, want %d", attempt, got, attempt)
		}
	}

	readsBefore := verifier.reads
	for attempt := domain.PinVerifyMaxFailures + 1; attempt <= domain.PinVerifyMaxFailures+4; attempt++ {
		_, err := svc.VerifyPIN(ctx, testBusinessID, "000000")
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("post-lockout attempt %d: expected ErrInvalidPIN, got %v", attempt, err)
		}
	}
	if verifier.reads != readsBefore {
		t.Errorf("post-lockout attempts reached the credential read %d times; a locked-out business must short-circuit",
			verifier.reads-readsBefore)
	}
}

// A correct PIN after lockout is still denied: the lockout is not a hint, and
// the success path must not be reachable until the window expires.
func TestVerifyPIN_CorrectPINAfterLockoutStillDenied(t *testing.T) {
	svc, _, clock := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})
	ctx := context.Background()

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		if _, err := svc.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}

	result, err := svc.VerifyPIN(ctx, testBusinessID, testValidPIN)
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("correct PIN during lockout returned %v, want ErrInvalidPIN", err)
	}
	if result != nil {
		t.Fatal("correct PIN during lockout returned a verified manager")
	}
}

// The lockout lives in the store, not in the Service: a brand-new Service over
// the same durable store still refuses. A process-local map would come back
// empty here and hand the attacker a fresh budget on every restart.
func TestVerifyPIN_LockoutSurvivesServiceRestart(t *testing.T) {
	ctx := context.Background()
	first, store, clock := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})
	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		if _, err := first.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}

	// Restart: a completely separate Service and verifier, same counter store.
	restarted := &Service{
		verifier: &stubVerifier{candidates: candidatesFor(t, testValidPIN)},
		attempts: store,
		now:      clock.Now,
	}
	if _, err := restarted.VerifyPIN(ctx, testBusinessID, testValidPIN); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("after restart the correct PIN returned %v, want ErrInvalidPIN: the lockout did not survive", err)
	}
}

func TestNewService_UsesInjectedDependencies(t *testing.T) {
	verifier := &stubVerifier{}
	attempts := newMemoryPinAttempts(newTestClock())
	svc := NewService(verifier, nil, nil, attempts)
	if svc.verifier != verifier || svc.attempts != attempts {
		t.Fatal("NewService did not retain the injected verifier and durable-attempt capability")
	}
}

// Unknown manager (no candidates), wrong PIN and a locked-out business must be
// indistinguishable: same sentinel, and every one of them performs a bcrypt
// comparison, so neither body nor latency can separate them.
func TestVerifyPIN_FailureModesIndistinguishableAndAllCompare(t *testing.T) {
	real, dummy := countCompares(t)

	// 1. A business with no usable credential at all.
	emptySvc, _, _ := newScenario(&stubVerifier{})
	_, emptyErr := emptySvc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	emptyCost := *real + *dummy

	// 2. A wrong PIN against a business that has candidates.
	*real, *dummy = 0, 0
	wrongVerifier := &stubVerifier{candidates: candidatesFor(t, "999111", "888222", "777333")}
	wrongSvc, _, _ := newScenario(wrongVerifier)
	_, wrongErr := wrongSvc.VerifyPIN(context.Background(), testBusinessID, "000000")
	wrongCost := *real + *dummy

	for name, err := range map[string]error{"no candidates": emptyErr, "wrong pin": wrongErr} {
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Errorf("%s returned %v, want ErrInvalidPIN — a distinct error is an enumeration oracle", name, err)
		}
	}

	// An empty candidate set has no hash to compare, so its dummy compare is its
	// only KDF work. A wrong PIN does its own real comparisons instead.
	if emptyCost == 0 {
		t.Error("no-candidate path performed no hash comparison, so its latency reveals the business has no manager PIN")
	}
	if wrongCost == 0 {
		t.Error("wrong-PIN path performed no hash comparison")
	}
	if wrongVerifier.reads == 0 {
		t.Error("wrong-PIN path never reached the credential read")
	}
	if *dummy != 0 {
		t.Errorf("wrong-PIN path burned %d dummy comparisons on top of its real scan; it must cost exactly one KDF pass", *dummy)
	}

	// 3. Locked out, which must look like the first two.
	lockedSvc, _, clock := newScenario(&stubVerifier{candidates: candidatesFor(t, "999111", "888222", "777333")})
	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		_, _ = lockedSvc.VerifyPIN(context.Background(), testBusinessID, "000000")
	}
	*real, *dummy = 0, 0
	_, lockedErr := lockedSvc.VerifyPIN(context.Background(), testBusinessID, "000000")
	if !errors.Is(lockedErr, domain.ErrInvalidPIN) {
		t.Errorf("locked-out returned %v, want ErrInvalidPIN", lockedErr)
	}
	if *dummy == 0 {
		t.Error("locked-out path performed no hash comparison, so its faster latency reveals the lockout")
	}
}

// A match at the first candidate and a match at the last must cost the same
// number of comparisons. Without the burn-out loop the scan short-circuits and
// response latency ranks a business's managers.
func TestVerifyPIN_MatchCostIsIndependentOfCandidatePosition(t *testing.T) {
	others := []string{"111111", "222222", "333333", "444444", "555555"}
	real, dummy := countCompares(t)

	// The matching manager is last: the scan finds it after every other
	// candidate has been compared.
	last := candidatesFor(t, append(append([]string{}, others...), testValidPIN)...)
	lastSvc, _, _ := newScenario(&stubVerifier{candidates: last})
	if _, err := lastSvc.VerifyPIN(context.Background(), testBusinessID, testValidPIN); err != nil {
		t.Fatalf("last-candidate match: %v", err)
	}
	realLast, dummyLast := *real, *dummy

	// The matching manager is first: the scan would stop immediately, so the
	// burn-out loop has to make up the difference.
	zeroCounts(real, dummy)
	first := candidatesFor(t, append([]string{testValidPIN}, others...)...)
	firstSvc, _, _ := newScenario(&stubVerifier{candidates: first})
	if _, err := firstSvc.VerifyPIN(context.Background(), testBusinessID, testValidPIN); err != nil {
		t.Fatalf("first-candidate match: %v", err)
	}
	realFirst, dummyFirst := *real, *dummy

	if realLast == realFirst {
		t.Errorf("both positions performed %d real comparisons, so this test cannot detect short-circuiting", realLast)
	}
	if got, want := realFirst+dummyFirst, realLast+dummyLast; got != want {
		t.Errorf("total comparisons: first-position match cost %d, last-position match cost %d; "+
			"latency must not reveal which manager matched", got, want)
	}
}

// The counter is keyed by business. Burning one business's budget must not deny
// another business, otherwise one caller could shut out an unrelated tenant.
//
// This is deliberately coarser than the pos_users (business, location, user)
// key, and the reason is structural: VerifyPINRequest carries only a PIN, so no
// manager identity exists to key on before the match succeeds. See the note on
// infrastructure.PINAttemptStore.
func TestVerifyPIN_LockoutIsScopedToOneBusiness(t *testing.T) {
	ctx := context.Background()

	for _, victimBusiness := range []int64{testBusinessID + 1, 9999} {
		t.Run("business "+strconv.FormatInt(victimBusiness, 10), func(t *testing.T) {
			attackerSvc, store, clock := newScenario(&stubVerifier{candidates: candidatesFor(t, "999111")})
			for i := 0; i < domain.PinVerifyMaxFailures; i++ {
				clock.advance(domain.PinLockoutDuration)
				if _, err := attackerSvc.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
					t.Fatalf("seeding failure %d: got %v", i, err)
				}
			}
			// Without this the scenario would pass vacuously: a build that
			// records nothing leaves every bystander verifiable, which is exactly
			// what the assertion below is meant to contrast against.
			if got := store.failureCount(testBusinessID); got != domain.PinVerifyMaxFailures {
				t.Fatalf("attacker recorded %d failures, want %d: the budget was not actually consumed",
					got, domain.PinVerifyMaxFailures)
			}

			victim := &stubVerifier{candidates: candidatesFor(t, testValidPIN)}
			victimSvc := &Service{verifier: victim, attempts: store, now: clock.Now}
			result, err := victimSvc.VerifyPIN(ctx, victimBusiness, testValidPIN)
			if err != nil {
				t.Fatalf("business %d was denied by another business's lockout: %v", victimBusiness, err)
			}
			if result == nil || !result.Valid {
				t.Fatalf("verify returned %+v, want a verified manager", result)
			}
			if victim.askedBusiness != victimBusiness {
				t.Errorf("credential read scoped to business %d, want %d: a tenant boundary was crossed",
					victim.askedBusiness, victimBusiness)
			}
		})
	}
}

// A counter that cannot be read must deny, not fall through: fail-open restores
// the unbounded guess budget the lockout exists to remove.
func TestVerifyPIN_CounterStoreErrorFailsClosed(t *testing.T) {
	svc, store, _ := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})
	store.readErr = errors.New("counter store down")

	_, err := svc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("correct PIN with an unreadable counter returned %v, want ErrInvalidPIN (fail closed)", err)
	}
}

// A Service with no attempt store at all must also deny. This is the exact
// shape of the pre-fix build, where nothing bounded the guess budget.
func TestVerifyPIN_MissingAttemptStoreFailsClosed(t *testing.T) {
	verifier := &stubVerifier{candidates: candidatesFor(t, testValidPIN)}
	svc := &Service{verifier: verifier, now: time.Now}

	_, err := svc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("VerifyPIN without an attempt store returned %v, want ErrInvalidPIN (fail closed)", err)
	}
	if verifier.reads != 0 {
		t.Error("the credential read was reached with no attempt store; the lockout can be bypassed entirely")
	}
}

// A success clears the counter, so a manager who eventually types their PIN
// correctly is not left permanently locked out.
func TestVerifyPIN_SuccessClearsCounter(t *testing.T) {
	svc, store, clock := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		clock.advance(domain.PinBackoff(2) + time.Second)
		if _, err := svc.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}
	if got := store.failureCount(testBusinessID); got != 2 {
		t.Fatalf("counter = %d before success, want 2", got)
	}
	clock.advance(domain.PinBackoff(2) + time.Second)
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testValidPIN); err != nil {
		t.Fatalf("correct PIN: %v", err)
	}
	if got := store.failureCount(testBusinessID); got != 0 {
		t.Errorf("counter = %d after a successful verification, want 0", got)
	}
}

// The backoff must bite before the full lockout, otherwise the Nth attempt is
// the only brake and an attacker can pace around it.
func TestVerifyPIN_IntermediateBackoffRefusesRetry(t *testing.T) {
	svc, _, _ := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})
	ctx := context.Background()

	if _, err := svc.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("first failure: %v", err)
	}
	if _, err := svc.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("second failure: %v", err)
	}
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testValidPIN); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("correct PIN inside the backoff window returned %v, want ErrInvalidPIN", err)
	}
}

// With the clock past the window, the manager is served again.
func TestVerifyPIN_SucceedsOnceBackoffWindowHasElapsed(t *testing.T) {
	svc, _, clock := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})
	ctx := context.Background()

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		if _, err := svc.VerifyPIN(ctx, testBusinessID, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}

	clock.advance(domain.PinLockoutDuration + time.Minute)
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testValidPIN); err != nil {
		t.Fatalf("correct PIN after the lockout window elapsed: %v", err)
	}
}

// The credential must be compared as a bcrypt hash. A stored value that is not
// a hash must never authenticate, which is what a legacy plaintext row would be
// after the migration drops the plaintext column.
func TestVerifyPIN_PlaintextStoredValueNeverAuthenticates(t *testing.T) {
	legacy := &stubVerifier{candidates: []domain.ManagerPINCandidate{{
		StaffID: 2000,
		PinHash: testValidPIN, // what a pre-migration plaintext row looks like
	}}}
	svc, _, _ := newScenario(legacy)

	result, err := svc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if err == nil && result != nil && result.Valid {
		t.Fatal("a plaintext stored PIN authenticated; the column is not actually hashed")
	}
}

// A business holding more candidates than the verification budget fails closed.
// Truncating the scan would let an attacker park a decoy above the cap and have
// the real manager's PIN skipped.
func TestVerifyPIN_OverCandidateBudgetFailsClosed(t *testing.T) {
	pins := make([]string, 0, domain.MaxManagerPINCandidates+1)
	for i := 0; i <= domain.MaxManagerPINCandidates; i++ {
		pins = append(pins, strconv.Itoa(100000+i*7))
	}
	verifier := &stubVerifier{candidates: candidatesFor(t, pins...)}
	svc, _, _ := newScenario(verifier)

	result, err := svc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("over-budget business returned %v / %+v, want ErrInvalidPIN (fail closed)", err, result)
	}
}

// An infrastructure fault must not be laundered into a credential verdict: a 500
// is honest, and it is not an existence oracle because it does not vary with
// the guess.
func TestVerifyPIN_ReadFailureIsNotACredentialRefusal(t *testing.T) {
	want := errors.New("connection refused")
	svc, _, _ := newScenario(&stubVerifier{readErr: want})

	_, err := svc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if !errors.Is(err, want) {
		t.Fatalf("read failure returned %v, want the underlying error so the handler can answer 500", err)
	}
	if errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatal("a read failure was reported as an invalid PIN; an outage would look like a wrong PIN")
	}
}

// An empty or whitespace PIN must be refused without reaching the credential
// store, and must still cost a comparison.
func TestVerifyPIN_EmptyPINRefusedWithComparison(t *testing.T) {
	_, dummy := countCompares(t)
	verifier := &stubVerifier{candidates: candidatesFor(t, testValidPIN)}
	svc, _, _ := newScenario(verifier)

	for _, pin := range []string{"", "   "} {
		_, err := svc.VerifyPIN(context.Background(), testBusinessID, pin)
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Errorf("PIN %q returned %v, want ErrInvalidPIN", pin, err)
		}
	}
	if *dummy == 0 {
		t.Error("empty-PIN path performed no hash comparison")
	}
	if verifier.reads != 0 {
		t.Error("empty PIN reached the credential read")
	}
}
