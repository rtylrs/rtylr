package application

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	testBusinessID int64 = 4242
	testLocationID int64 = 7
	testTargetUser int64 = 42
	testValidPIN         = "123456"
)

// memoryPinAttempts is a PinAttemptStore that keeps its state outside the
// Service. That is deliberate: the restart test below drops the Service and
// builds a fresh one over the same store, and it still sees the lockout — which
// is exactly the property an in-process map inside Service would fail.
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

func pinAttemptKey(businessID, locationID, userID int64) string {
	return strconv.FormatInt(businessID, 10) + "/" +
		strconv.FormatInt(locationID, 10) + "/" +
		strconv.FormatInt(userID, 10)
}

func (m *memoryPinAttempts) PinRetryAt(_ context.Context, businessID, locationID, userID int64) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readErr != nil {
		return time.Time{}, m.readErr
	}
	key := pinAttemptKey(businessID, locationID, userID)
	return domain.PinRetryAt(m.failures[key], m.last[key]), nil
}

func (m *memoryPinAttempts) RecordPINFailure(_ context.Context, businessID, locationID, userID int64) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readErr != nil {
		return time.Time{}, m.readErr
	}
	key := pinAttemptKey(businessID, locationID, userID)
	m.failures[key]++
	m.last[key] = m.clock.Now()
	return domain.PinRetryAt(m.failures[key], m.last[key]), nil
}

func (m *memoryPinAttempts) ResetPINFailures(_ context.Context, businessID, locationID, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := pinAttemptKey(businessID, locationID, userID)
	delete(m.failures, key)
	delete(m.last, key)
	return nil
}

func (m *memoryPinAttempts) ResetPINFailuresForPOSUser(context.Context, int64, int64) error {
	return nil
}

func (m *memoryPinAttempts) failureCount(businessID, locationID, userID int64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failures[pinAttemptKey(businessID, locationID, userID)]
}

// stubPinVerifier answers the credential read. hash is compared for real, so a
// success here is a genuine verification rather than a stubbed outcome.
type stubPinVerifier struct {
	user  *domain.POSUser
	hash  string
	found bool
	// lookups counts the credential reads that reached this verifier. A
	// locked-out key must not get that far, and the unknown-operator path must
	// still burn a comparison elsewhere.
	lookups int
}

func (s *stubPinVerifier) FindForVerify(_ context.Context, _, _, _ int64) (*domain.POSUser, error) {
	s.lookups++
	if !s.found {
		return nil, domain.ErrUserNotFound
	}
	return s.user, nil
}

func (s *stubPinVerifier) GetPINHash(_ context.Context, _, id int64) (string, error) {
	if s.user == nil || id != s.user.ID {
		return "", domain.ErrUserNotFound
	}
	return s.hash, nil
}

func (s *stubPinVerifier) GetAuthorizationVersion(context.Context, int64, int64) (int64, error) {
	return 7, nil
}

func newVerifiableStub(t *testing.T) *stubPinVerifier {
	t.Helper()
	hash, err := passwordutils.HashPIN(testValidPIN)
	if err != nil {
		t.Fatalf("hash test PIN: %v", err)
	}
	return &stubPinVerifier{
		found: true,
		hash:  hash,
		user: &domain.POSUser{
			ID: 900, BusinessID: testBusinessID, UserID: testTargetUser,
			LocationID: testLocationID, Name: "Manager", Role: domain.RoleManager, IsActive: true,
		},
	}
}

func newPINService(verifier pinVerifier, store PinAttemptStore) *Service {
	return &Service{pins: verifier, pinAttempts: store, now: time.Now}
}

// newScenario wires a Service and a counter store onto one shared clock, which
// is the arrangement a real deployment has. Both must read the same clock: the
// store stamps the last failure with it and the Service compares against it.
func newScenario(verifier pinVerifier) (*Service, *memoryPinAttempts, *testClock) {
	clock := newTestClock()
	store := newMemoryPinAttempts(clock)
	svc := newPINService(verifier, store)
	svc.now = clock.Now
	return svc, store, clock
}

// The Nth attempt opens the lockout: a wrong PIN is rejected PinVerifyMaxFailures
// times, and every attempt after that is refused without reaching the hash.
// The clock is stepped past each backoff window so the test measures the
// attempt budget rather than the wall-clock wait.
func TestVerifyPIN_LocksOutOnNthAttempt(t *testing.T) {
	verifier := newVerifiableStub(t)
	svc, store, clock := newScenario(verifier)
	ctx := context.Background()

	for attempt := 1; attempt <= domain.PinVerifyMaxFailures; attempt++ {
		clock.advance(domain.PinLockoutDuration)
		_, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000")
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("attempt %d: expected ErrInvalidPIN, got %v", attempt, err)
		}
		if got := store.failureCount(testBusinessID, testLocationID, testTargetUser); got != attempt {
			t.Fatalf("attempt %d: counter = %d, want %d", attempt, got, attempt)
		}
	}

	lookupsBefore := verifier.lookups
	for attempt := domain.PinVerifyMaxFailures + 1; attempt <= domain.PinVerifyMaxFailures+4; attempt++ {
		_, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000")
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("post-lockout attempt %d: expected ErrInvalidPIN, got %v", attempt, err)
		}
	}
	if verifier.lookups != lookupsBefore {
		t.Errorf("post-lockout attempts reached the credential lookup %d times; a locked-out key must short-circuit",
			verifier.lookups-lookupsBefore)
	}
}

// A correct PIN after lockout is still denied: the lockout is not a hint, and
// the success path must not be reachable until the window expires.
func TestVerifyPIN_CorrectPINAfterLockoutStillDenied(t *testing.T) {
	verifier := newVerifiableStub(t)
	svc, _, clock := newScenario(verifier)
	ctx := context.Background()

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "999999"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}

	result, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, testValidPIN)
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("correct PIN during lockout returned %v, want ErrInvalidPIN", err)
	}
	if result != nil {
		t.Fatal("correct PIN during lockout returned a verified user")
	}
}

// The lockout lives in the store, not in the Service: a brand-new Service over
// the same durable store still refuses. A process-local map would come back
// empty here and hand the attacker a fresh budget on every restart.
func TestVerifyPIN_LockoutSurvivesServiceRestart(t *testing.T) {
	ctx := context.Background()
	first, store, clock := newScenario(newVerifiableStub(t))
	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		if _, err := first.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "999999"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}

	// Restart: a completely separate Service and verifier, same counter store
	// and same durable state.
	restarted := newPINService(newVerifiableStub(t), store)
	restarted.now = clock.Now
	if _, err := restarted.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, testValidPIN); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("after restart the correct PIN returned %v, want ErrInvalidPIN: the lockout did not survive", err)
	}
}

func TestNewServiceUsesInjectedAttemptStore(t *testing.T) {
	store := newMemoryPinAttempts(newTestClock())
	svc := NewService(nil, store)
	if svc.pinAttempts != store {
		t.Fatal("NewService did not retain its injected attempt store")
	}
}

// Unknown operator, wrong PIN, operator with no PIN set and a locked-out key must
// be indistinguishable: same sentinel, and every one of them performs a hash
// comparison, so neither body nor latency can separate them.
func TestVerifyPIN_FailureModesIndistinguishableAndAllCompare(t *testing.T) {
	burns := 0
	original := burnPINCompare
	burnPINCompare = func(pin string) bool { burns++; return false }
	t.Cleanup(func() { burnPINCompare = original })

	wrongPINVerifier := newVerifiableStub(t)
	wrongSvc, _, clock := newScenario(wrongPINVerifier)
	unknownSvc, _, _ := newScenario(&stubPinVerifier{found: false})
	noPINSvc, _, _ := newScenario(
		&stubPinVerifier{found: true, hash: "", user: &domain.POSUser{ID: 901, UserID: testTargetUser}},
	)
	ctx := context.Background()

	burns = 0
	_, unknownErr := unknownSvc.VerifyPIN(ctx, testBusinessID, testLocationID, 999999, testValidPIN)
	unknownBurns := burns

	burns = 0
	_, wrongErr := wrongSvc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000")
	wrongBurns := burns

	burns = 0
	_, noPINErr := noPINSvc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, testValidPIN)
	noPINBurns := burns

	for name, err := range map[string]error{"unknown operator": unknownErr, "wrong pin": wrongErr, "no pin set": noPINErr} {
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Errorf("%s returned %v, want ErrInvalidPIN — a distinct error is an enumeration oracle", name, err)
		}
	}
	// Unknown operator and no-PIN-set have no real hash to compare, so the
	// dummy comparison is their only KDF work. Wrong PIN does its own real
	// comparison, which is why it is measured through the verifier instead.
	if unknownBurns == 0 {
		t.Error("unknown-operator path performed no hash comparison, so its latency reveals that the user id does not exist")
	}
	if noPINBurns == 0 {
		t.Error("no-PIN-set path performed no hash comparison, so its latency reveals the operator has no PIN")
	}
	if wrongBurns != 0 {
		t.Errorf("wrong-PIN path burned %d dummy comparisons on top of its real one; it must cost exactly one KDF", wrongBurns)
	}
	if wrongPINVerifier.lookups == 0 {
		t.Error("wrong-PIN path never reached the credential lookup")
	}

	// Locked out is the fourth mode and must look like the first three.
	burns = 0
	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		_, _ = wrongSvc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000")
	}
	_, lockedErr := wrongSvc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000")
	if !errors.Is(lockedErr, domain.ErrInvalidPIN) {
		t.Errorf("locked-out returned %v, want ErrInvalidPIN", lockedErr)
	}
	if burns == 0 {
		t.Error("locked-out path performed no hash comparison, so its faster latency reveals the lockout")
	}
}

// The counter is keyed by (business, location, operator). Burning one operator's
// budget must not deny a colleague at the same location, and the same operator
// at another location or business must be unaffected — otherwise one attacker
// could shut a whole store out of PIN verification.
func TestVerifyPIN_LockoutIsScopedToOneOperatorAndLocation(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name                                       string
		victimBusiness, victimLocation, victimUser int64
	}{
		{"different operator", testBusinessID, testLocationID, 43},
		{"different location", testBusinessID, 8, testTargetUser},
		{"different business", 9999, testLocationID, testTargetUser},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			attackerSvc, store, clock := newScenario(newVerifiableStub(t))
			for i := 0; i < domain.PinVerifyMaxFailures; i++ {
				clock.advance(domain.PinLockoutDuration)
				if _, err := attackerSvc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
					t.Fatalf("seeding failure %d: got %v", i, err)
				}
			}
			// Without this the scenario would pass vacuously: a build that
			// records nothing leaves every bystander verifiable, which is exactly
			// what the assertion below is meant to contrast against.
			if got := store.failureCount(testBusinessID, testLocationID, testTargetUser); got != domain.PinVerifyMaxFailures {
				t.Fatalf("attacker recorded %d failures, want %d: the target's budget was not actually consumed",
					got, domain.PinVerifyMaxFailures)
			}

			victimHash, err := passwordutils.HashPIN(testValidPIN)
			if err != nil {
				t.Fatalf("hash: %v", err)
			}
			bystander := &stubPinVerifier{
				found: true, hash: victimHash,
				user: &domain.POSUser{ID: 901, UserID: tc.victimUser, LocationID: tc.victimLocation, IsActive: true},
			}
			bystanderSvc := newPINService(bystander, store)
			bystanderSvc.now = clock.Now
			result, err := bystanderSvc.VerifyPIN(ctx, tc.victimBusiness, tc.victimLocation, tc.victimUser, testValidPIN)
			if err != nil {
				t.Fatalf("%s was denied by another operator's lockout: %v", tc.name, err)
			}
			if result == nil || result.UserID != tc.victimUser {
				t.Fatalf("%s verify returned %+v, want a verified user %d", tc.name, result, tc.victimUser)
			}
		})
	}
}

// A counter that cannot be read must deny, not fall through: fail-open restores
// the unbounded guess budget the lockout exists to remove.
func TestVerifyPIN_CounterStoreErrorFailsClosed(t *testing.T) {
	svc, store, _ := newScenario(newVerifiableStub(t))
	store.readErr = errors.New("counter store down")

	_, err := svc.VerifyPIN(context.Background(), testBusinessID, testLocationID, testTargetUser, testValidPIN)
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("correct PIN with an unreadable counter returned %v, want ErrInvalidPIN (fail closed)", err)
	}
}

// A success clears the counter, so an operator who eventually types their PIN
// correctly is not left permanently locked out.
func TestVerifyPIN_SuccessClearsCounter(t *testing.T) {
	svc, store, clock := newScenario(newVerifiableStub(t))
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		clock.advance(domain.PinBackoff(2) + time.Second)
		if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}
	if got := store.failureCount(testBusinessID, testLocationID, testTargetUser); got != 2 {
		t.Fatalf("counter = %d before success, want 2", got)
	}
	clock.advance(domain.PinBackoff(2) + time.Second)
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, testValidPIN); err != nil {
		t.Fatalf("correct PIN: %v", err)
	}
	if got := store.failureCount(testBusinessID, testLocationID, testTargetUser); got != 0 {
		t.Errorf("counter = %d after a successful verification, want 0", got)
	}
}

// The backoff must bite before the full lockout, otherwise the Nth attempt is
// the only brake and an attacker can pace around it.
func TestVerifyPIN_IntermediateBackoffRefusesRetry(t *testing.T) {
	svc, _, _ := newScenario(newVerifiableStub(t))
	ctx := context.Background()

	if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("first failure: %v", err)
	}
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("second failure: %v", err)
	}
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, testValidPIN); !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("correct PIN inside the backoff window returned %v, want ErrInvalidPIN", err)
	}
}

// With the clock past the window, the operator is served again.
func TestVerifyPIN_SucceedsOnceBackoffWindowHasElapsed(t *testing.T) {
	svc, _, clock := newScenario(newVerifiableStub(t))
	ctx := context.Background()

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, "000000"); !errors.Is(err, domain.ErrInvalidPIN) {
			t.Fatalf("seeding failure %d: got %v", i, err)
		}
	}

	clock.advance(domain.PinLockoutDuration + time.Minute)
	if _, err := svc.VerifyPIN(ctx, testBusinessID, testLocationID, testTargetUser, testValidPIN); err != nil {
		t.Fatalf("correct PIN after the lockout window elapsed: %v", err)
	}
}
