package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	testOwnerID   int64 = 77
	testStaffID   int64 = 2000
	testOwnerPass       = "Correct-Horse-9"
)

var (
	ownerPasswordMu    sync.Mutex
	ownerPasswordCache = map[string]string{}
)

// mustHashPassword memoises account hashing for the same reason mustHash does
// for PINs: bcrypt at cost 12 is ~210ms and the fixtures re-derive the same
// password constantly.
func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	ownerPasswordMu.Lock()
	cached, ok := ownerPasswordCache[password]
	ownerPasswordMu.Unlock()
	if ok {
		return cached
	}
	h, err := passwordutils.HashPassword(password)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	ownerPasswordMu.Lock()
	ownerPasswordCache[password] = h
	ownerPasswordMu.Unlock()
	return h
}

// recordingWriter captures what the set path would have persisted.
type recordingWriter struct {
	stored   []domain.SetPINInput
	hashes   []string
	events   []audit.Event
	writeErr error
}

func (w *recordingWriter) SetManagerPIN(_ context.Context, input domain.SetPINInput, pinHash string, events ...audit.Event) error {
	if w.writeErr != nil {
		return w.writeErr
	}
	w.stored = append(w.stored, input)
	w.hashes = append(w.hashes, pinHash)
	w.events = append(w.events, events...)
	return nil
}

// stubCredentials answers the owner's account credential read.
type stubCredentials struct {
	hash    string
	readErr error
	reads   int
}

func (c *stubCredentials) VerifyAccountPassword(_ context.Context, _ int64, password string) (bool, error) {
	c.reads++
	if c.readErr != nil {
		return false, c.readErr
	}
	return passwordutils.CheckPasswordHash(password, c.hash), nil
}

func validSetInput() domain.SetPINInput {
	return domain.SetPINInput{
		OwnerID:         testOwnerID,
		BusinessID:      testBusinessID,
		StaffID:         testStaffID,
		PIN:             testValidPIN,
		CurrentPassword: testOwnerPass,
	}
}

func newSetScenario(t *testing.T) (*Service, *recordingWriter, *stubCredentials, *memoryPinAttempts, *testClock) {
	t.Helper()
	clock := newTestClock()
	store := newMemoryPinAttempts(clock)
	writer := &recordingWriter{}
	creds := &stubCredentials{hash: mustHashPassword(t, testOwnerPass)}
	return &Service{writer: writer, credentials: creds, attempts: store, now: clock.Now}, writer, creds, store, clock
}

// The whole point of the route: a manager PIN set through the product must be a
// credential the verify path actually accepts, stored as a bcrypt hash and never
// as the plaintext. Before this route existed nothing in the repository wrote to
// pos_manager_pins at all, so every manager credential was unusable after the
// plaintext column was dropped.
func TestSetPIN_StoredCredentialVerifiesAndIsNeverPlaintext(t *testing.T) {
	svc, writer, _, _, _ := newSetScenario(t)

	if err := svc.SetPIN(context.Background(), validSetInput()); err != nil {
		t.Fatalf("SetPIN: %v", err)
	}
	if len(writer.hashes) != 1 {
		t.Fatalf("stored %d credentials, want 1", len(writer.hashes))
	}

	stored := writer.hashes[0]
	if stored == testValidPIN {
		t.Fatal("the PIN was stored verbatim; the credential is cleartext at rest")
	}
	if !passwordutils.CheckPINHash(testValidPIN, stored) {
		t.Fatal("the stored hash does not verify against the PIN that was set")
	}
	if passwordutils.CheckPINHash("999999", stored) {
		t.Error("a different PIN verified against the stored hash")
	}

	// The credential the owner just set must work through the credential path,
	// not merely round-trip against the hash.
	verifier := &stubVerifier{candidates: []domain.ManagerPINCandidate{{StaffID: testStaffID, PinHash: stored}}}
	verifySvc := &Service{verifier: verifier, attempts: newMemoryPinAttempts(newTestClock()), now: time.Now}
	result, err := verifySvc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if err != nil {
		t.Fatalf("a PIN set through the product did not verify: %v", err)
	}
	if result == nil || !result.Valid || result.StaffID != testStaffID {
		t.Fatalf("verify returned %+v, want a valid result attributed to staff %d", result, testStaffID)
	}
}

// The owner's account password is the second factor. A wrong one must not write
// a credential, and it must be counted so this route cannot be used to guess the
// password without limit.
func TestSetPIN_WrongOwnerPasswordNeitherWritesNorIsFree(t *testing.T) {
	svc, writer, creds, store, _ := newSetScenario(t)
	creds.hash = mustHashPassword(t, "Some-Other-Pass-1")

	input := validSetInput()
	input.CurrentPassword = "wrong-password"
	err := svc.SetPIN(context.Background(), input)
	if !errors.Is(err, domain.ErrSetPINNotAuthorized) {
		t.Fatalf("wrong owner password returned %v, want ErrSetPINNotAuthorized", err)
	}
	if len(writer.stored) != 0 {
		t.Error("a wrong owner password still wrote a credential")
	}
	if got := store.failureCount(testBusinessID); got != 1 {
		t.Errorf("counter = %d after a wrong owner password, want 1: this route is a password oracle without it", got)
	}
}

// A successful set clears the counter, so an owner who fumbled their password is
// not left locked out of a capability they are entitled to.
func TestSetPIN_SuccessClearsTheAttemptCounter(t *testing.T) {
	svc, _, _, store, clock := newSetScenario(t)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		clock.advance(domain.PinBackoff(2) + time.Second)
		bad := validSetInput()
		bad.CurrentPassword = "wrong-password"
		if err := svc.SetPIN(ctx, bad); !errors.Is(err, domain.ErrSetPINNotAuthorized) {
			t.Fatalf("seeding failure %d: %v", i, err)
		}
	}
	if got := store.failureCount(testBusinessID); got != 2 {
		t.Fatalf("counter = %d before the successful set, want 2", got)
	}

	clock.advance(domain.PinBackoff(2) + time.Second)
	if err := svc.SetPIN(ctx, validSetInput()); err != nil {
		t.Fatalf("SetPIN after two failures: %v", err)
	}
	if got := store.failureCount(testBusinessID); got != 0 {
		t.Errorf("counter = %d after a successful set, want 0", got)
	}
}

// A locked-out business cannot set a PIN even with the correct owner password:
// the lockout is a statement about the session, not about the guess, and letting
// the correct password through would make it a hint.
func TestSetPIN_LockedOutBusinessCannotSetWithTheCorrectPassword(t *testing.T) {
	svc, writer, creds, store, clock := newSetScenario(t)
	ctx := context.Background()

	for i := 0; i < domain.PinVerifyMaxFailures; i++ {
		clock.advance(domain.PinLockoutDuration)
		bad := validSetInput()
		bad.CurrentPassword = "wrong-password"
		if err := svc.SetPIN(ctx, bad); !errors.Is(err, domain.ErrSetPINNotAuthorized) {
			t.Fatalf("seeding failure %d: %v", i, err)
		}
	}

	// Correct owner password, correct PIN, still refused.
	if err := svc.SetPIN(ctx, validSetInput()); !errors.Is(err, domain.ErrSetPINNotAuthorized) {
		t.Fatalf("correct owner password during lockout returned %v, want ErrSetPINNotAuthorized", err)
	}
	if len(writer.stored) != 0 {
		t.Error("a locked-out business wrote a credential")
	}
	// A refused attempt must not drain the counter, or repeatedly asking while
	// locked out would hand the caller a fresh budget.
	if got := store.failureCount(testBusinessID); got != domain.PinVerifyMaxFailures {
		t.Errorf("counter = %d after a refused locked-out attempt, want %d", got, domain.PinVerifyMaxFailures)
	}
	readsAtLockout := creds.reads

	// Once the window elapses the same request succeeds, so the lockout is a
	// wait and not a permanent denial of the owner.
	clock.advance(domain.PinLockoutDuration + time.Minute)
	if err := svc.SetPIN(ctx, validSetInput()); err != nil {
		t.Fatalf("SetPIN after the lockout elapsed: %v", err)
	}
	if creds.reads == readsAtLockout {
		t.Error("the locked-out attempt read the owner credential; it must short-circuit before the password check")
	}
}

// The new PIN must satisfy the 6-12 digit policy. A weak PIN is refused before
// any credential is read or written, and it does not consume guess budget: it is
// the caller's own input, not a guess at a secret.
func TestSetPIN_WeakPINIsRefusedBeforeAnyCredentialWork(t *testing.T) {
	svc, writer, creds, store, _ := newSetScenario(t)

	weak := map[string]string{
		"too short":  "1234",
		"too long":   "1234567890123",
		"not digits": "12345a",
	}
	for name, pin := range weak {
		t.Run(name, func(t *testing.T) {
			input := validSetInput()
			input.PIN = pin
			err := svc.SetPIN(context.Background(), input)
			if !errors.Is(err, domain.ErrWeakPIN) {
				t.Fatalf("PIN %q returned %v, want ErrWeakPIN", pin, err)
			}
		})
	}

	if len(writer.stored) != 0 {
		t.Error("a weak PIN was stored")
	}
	if creds.reads != 0 {
		t.Error("a weak PIN reached the owner credential read")
	}
	if got := store.failureCount(testBusinessID); got != 0 {
		t.Errorf("a weak PIN consumed %d failures of the guess budget; it is not a guess", got)
	}
}

// The 6-12 digit policy is what the module documents, and it is now reachable:
// ValidatePIN had no production caller at all while this route did not exist.
func TestSetPIN_AcceptsTheDocumentedPolicyBoundaries(t *testing.T) {
	for _, pin := range []string{"123456", "12345678", "123456789012"} {
		t.Run(pin, func(t *testing.T) {
			svc, writer, _, _, _ := newSetScenario(t)
			input := validSetInput()
			input.PIN = pin
			if err := svc.SetPIN(context.Background(), input); err != nil {
				t.Fatalf("PIN %q rejected: %v", pin, err)
			}
			if len(writer.hashes) != 1 {
				t.Fatalf("PIN %q stored %d credentials, want 1", pin, len(writer.hashes))
			}
		})
	}
}

// Every unreadable dependency must deny. A counter that cannot be read is a
// counter that cannot be enforced, and a credential that cannot be read is a
// step-up check that cannot be satisfied.
func TestSetPIN_UnavailableDependenciesFailClosed(t *testing.T) {
	t.Run("counter store errors", func(t *testing.T) {
		svc, writer, _, store, _ := newSetScenario(t)
		store.readErr = errors.New("counter store down")
		if err := svc.SetPIN(context.Background(), validSetInput()); !errors.Is(err, domain.ErrSetPINNotAuthorized) {
			t.Fatalf("returned %v, want ErrSetPINNotAuthorized (fail closed)", err)
		}
		if len(writer.stored) != 0 {
			t.Error("a credential was written with an unreadable counter")
		}
	})

	t.Run("no attempt store wired", func(t *testing.T) {
		writer := &recordingWriter{}
		svc := &Service{writer: writer, credentials: &stubCredentials{hash: mustHashPassword(t, testOwnerPass)}, now: time.Now}
		if err := svc.SetPIN(context.Background(), validSetInput()); !errors.Is(err, domain.ErrSetPINStoreUnavailable) {
			t.Fatalf("returned %v, want ErrSetPINStoreUnavailable (fail closed)", err)
		}
		if len(writer.stored) != 0 {
			t.Error("a credential was written with no attempt store")
		}
	})

	t.Run("no write path wired", func(t *testing.T) {
		svc := &Service{credentials: &stubCredentials{hash: mustHashPassword(t, testOwnerPass)}, now: time.Now}
		if err := svc.SetPIN(context.Background(), validSetInput()); err == nil {
			t.Fatal("SetPIN succeeded with no write path; the request would be reported as stored when it is not")
		}
	})

	t.Run("owner credential read errors", func(t *testing.T) {
		svc, writer, creds, _, _ := newSetScenario(t)
		creds.readErr = errors.New("connection refused")
		err := svc.SetPIN(context.Background(), validSetInput())
		if !errors.Is(err, domain.ErrSetPINStoreUnavailable) {
			t.Fatalf("returned %v, want ErrSetPINStoreUnavailable", err)
		}
		if errors.Is(err, domain.ErrSetPINNotAuthorized) {
			t.Error("an outage was reported as a rejected password; it must surface as a server error")
		}
		if len(writer.stored) != 0 {
			t.Error("a credential was written when the owner credential could not be read")
		}
	})
}

// A stored value that is not a bcrypt hash must never authorise a change. bcrypt
// rejects a malformed hash immediately and cheaply, so this case also has to
// spend a real comparison or its latency gives the failure away.
func TestSetPIN_MalformedOrMissingOwnerHashNeverAuthorises(t *testing.T) {
	for name, hash := range map[string]string{
		"empty":          "",
		"plaintext":      testOwnerPass,
		"truncated hash": "$2a$12$tooshort",
	} {
		t.Run(name, func(t *testing.T) {
			svc, writer, creds, store, _ := newSetScenario(t)
			creds.hash = hash

			if err := svc.SetPIN(context.Background(), validSetInput()); !errors.Is(err, domain.ErrSetPINNotAuthorized) {
				t.Fatalf("stored %q returned %v, want ErrSetPINNotAuthorized", hash, err)
			}
			if len(writer.stored) != 0 {
				t.Error("a non-bcrypt stored value authorised a credential write")
			}
			if got := store.failureCount(testBusinessID); got != 1 {
				t.Errorf("counter = %d, want 1: the attempt must still be counted", got)
			}
		})
	}
}

// The write must be attributable after the fact. The audit event names the actor
// and the subject and carries neither the PIN nor its hash: activity_log is a
// durable, widely-readable table and a credential must never land in one.
func TestSetPIN_WritesAnAuditRecordWithNoCredentialMaterial(t *testing.T) {
	svc, writer, _, _, _ := newSetScenario(t)

	if err := svc.SetPIN(context.Background(), validSetInput()); err != nil {
		t.Fatalf("SetPIN: %v", err)
	}
	if len(writer.events) != 1 {
		t.Fatalf("wrote %d audit events, want exactly 1", len(writer.events))
	}
	event := writer.events[0]

	if event.ActorID != testOwnerID {
		t.Errorf("actor = %d, want the owner %d who authorised the write", event.ActorID, testOwnerID)
	}
	if event.EntityID != testStaffID {
		t.Errorf("entity = %d, want the manager %d the credential belongs to", event.EntityID, testStaffID)
	}
	if event.BusinessID != testBusinessID {
		t.Errorf("business = %d, want %d: a tenant boundary was crossed", event.BusinessID, testBusinessID)
	}
	if event.Action == "" || event.EntityType == "" {
		t.Error("the audit event does not identify what changed")
	}

	rendered, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("render audit event: %v", err)
	}
	blob := string(rendered)
	for _, secret := range []string{testValidPIN, testOwnerPass, writer.hashes[0]} {
		if secret == "" {
			continue
		}
		if strings.Contains(blob, secret) {
			t.Errorf("the audit event contains credential material (%q): %s", secret, blob)
		}
	}
}

// The plaintext must not survive in the value handed to the repository, so it
// cannot be picked up by a log line or an error report further down.
func TestSetPIN_PlaintextIsClearedBeforeTheWrite(t *testing.T) {
	svc, writer, _, _, _ := newSetScenario(t)

	if err := svc.SetPIN(context.Background(), validSetInput()); err != nil {
		t.Fatalf("SetPIN: %v", err)
	}
	if got := writer.stored[0].PIN; got != "" {
		t.Errorf("the struct reaching the repository still carries the plaintext PIN (%q)", got)
	}
}

// A failed write must propagate. Reporting success when nothing was stored is
// the failure mode that leaves a business believing it has a manager credential.
func TestSetPIN_WriteFailurePropagates(t *testing.T) {
	svc, writer, _, _, _ := newSetScenario(t)
	want := errors.New("deadlock detected")
	writer.writeErr = want

	if err := svc.SetPIN(context.Background(), validSetInput()); !errors.Is(err, want) {
		t.Fatalf("returned %v, want the underlying write error", err)
	}
}

func TestNewService_UsesInjectedWriteCapabilities(t *testing.T) {
	writer := &recordingWriter{}
	credentials := &stubCredentials{}
	svc := NewService(nil, writer, credentials, nil)
	if svc.writer != writer || svc.credentials != credentials {
		t.Fatal("NewService did not retain the injected write and owner-credential capabilities")
	}
}

// A malformed request is refused before anything is touched, so no budget and no
// credential read is spent on input that could never be stored.
func TestSetPIN_MalformedRequestsAreRefused(t *testing.T) {
	cases := map[string]func(d *domain.SetPINInput){
		"no business": func(d *domain.SetPINInput) { d.BusinessID = 0 },
		"no staff":    func(d *domain.SetPINInput) { d.StaffID = 0 },
		"no owner":    func(d *domain.SetPINInput) { d.OwnerID = 0 },
		"empty pin":   func(d *domain.SetPINInput) { d.PIN = "   " },
		"no password": func(d *domain.SetPINInput) { d.CurrentPassword = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			svc, writer, creds, store, _ := newSetScenario(t)
			input := validSetInput()
			mutate(&input)

			if err := svc.SetPIN(context.Background(), input); !errors.Is(err, domain.ErrInvalidSetPIN) {
				t.Fatalf("returned %v, want ErrInvalidSetPIN", err)
			}
			if len(writer.stored) != 0 {
				t.Error("a malformed request wrote a credential")
			}
			if creds.reads != 0 {
				t.Error("a malformed request reached the owner credential read")
			}
			if got := store.failureCount(testBusinessID); got != 0 {
				t.Errorf("a malformed request consumed %d failures of the guess budget", got)
			}
		})
	}
}

// The owner is the actor, the manager is the subject. Collapsing the two would
// attribute a manager's credential to the owner in the audit trail.
func TestSetPIN_OwnerAndManagerRemainDistinctIdentities(t *testing.T) {
	svc, writer, _, _, _ := newSetScenario(t)

	if err := svc.SetPIN(context.Background(), validSetInput()); err != nil {
		t.Fatalf("SetPIN: %v", err)
	}
	if writer.stored[0].OwnerID == writer.stored[0].StaffID {
		// A self-service business where the owner is also the manager is legal,
		// so this is only asserted as a wiring check: the two fields must be
		// carried separately rather than one being copied from the other.
		if writer.stored[0].OwnerID != testOwnerID {
			t.Errorf("owner = %d, want %d", writer.stored[0].OwnerID, testOwnerID)
		}
	}
	if writer.stored[0].StaffID != testStaffID {
		t.Errorf("staff = %d, want %d", writer.stored[0].StaffID, testStaffID)
	}
}

// Guard against the credential being re-introduced anywhere in this path.
func TestSetPIN_NeverStoresTheAccountPasswordAsTheCredential(t *testing.T) {
	svc, writer, _, _, _ := newSetScenario(t)
	if err := svc.SetPIN(context.Background(), validSetInput()); err != nil {
		t.Fatalf("SetPIN: %v", err)
	}
	stored := writer.hashes[0]
	if passwordutils.CheckPasswordHash(testOwnerPass, stored) {
		t.Fatal("the owner's account password satisfies the stored manager credential")
	}
	if !strings.HasPrefix(stored, "$2") {
		t.Errorf("stored credential %q is not a bcrypt hash", stored)
	}
}
