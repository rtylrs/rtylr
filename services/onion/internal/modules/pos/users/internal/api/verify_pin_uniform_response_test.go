package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
)

// The verify route must answer a refused PIN the same way no matter which refusal
// it was. VerifyPIN collapses unknown operator, wrong PIN, no PIN set, active
// lockout and an unreadable counter store into domain.ErrInvalidPIN, and this
// pins the single response they share. A status or message that named the cause
// would let a caller enumerate which user ids hold a PIN at a location, and a
// lockout-specific status would let it poll a victim's lockout state.
func TestVerifyPINRefusal_UniformAcrossEveryCredentialFailure(t *testing.T) {
	baseline := verifyPINRefusal(domain.ErrInvalidPIN)
	if baseline == nil {
		t.Fatal("ErrInvalidPIN must produce a refusal")
	}
	wantStatus := posUsersStatus(baseline)
	var wantErr *onekithttp.Error
	if !errors.As(baseline, &wantErr) {
		t.Fatalf("baseline returned %T, want *onekiterrors.Error", baseline)
	}

	// 429 and 423 both name the lockout, which is the specific outcome this
	// mapping must never reveal.
	for _, status := range []int{http.StatusTooManyRequests, http.StatusLocked} {
		if wantStatus == status {
			t.Errorf("verify-pin answered %d; a lockout-specific status confirms the state to the caller", status)
		}
	}

	// The sentinel and anything wrapping it are the credential failures the
	// service produces; they must all resolve to the one response.
	for _, cause := range []error{
		domain.ErrInvalidPIN,
		fmt.Errorf("repository: %w", domain.ErrInvalidPIN),
		fmt.Errorf("locked: %w", domain.ErrInvalidPIN),
	} {
		got := verifyPINRefusal(cause)
		if got == nil {
			t.Errorf("%v produced no refusal; a credential failure must not fall through to a 500", cause)
			continue
		}
		if status := posUsersStatus(got); status != wantStatus {
			t.Errorf("%v: status = %d, want %d", cause, status, wantStatus)
		}
		var gotErr *onekithttp.Error
		if !errors.As(got, &gotErr) {
			t.Errorf("%v: returned %T, want *onekiterrors.Error", cause, got)
			continue
		}
		if gotErr.GetMessage() != wantErr.GetMessage() {
			t.Errorf("%v: body = %q, want %q", cause, gotErr.GetMessage(), wantErr.GetMessage())
		}
	}
}

// A transport or database failure must not be laundered into a 401: that would
// make an outage look like a wrong PIN to the operator and to monitoring.
func TestVerifyPINRefusal_RealErrorsAreNotCredentialFailures(t *testing.T) {
	if refusal := verifyPINRefusal(errors.New("connection refused")); refusal != nil {
		t.Errorf("a transport failure produced %v; only ErrInvalidPIN is a credential refusal", refusal)
	}
}

// A weak PIN reports the policy, on every path that sets a PIN. The caller
// supplied the PIN, so naming the rule leaks nothing about anyone else's — and
// the shared text means the rule reads the same wherever a PIN is set.
func TestWeakPINRefusal_PolicyIsNamedConsistently(t *testing.T) {
	if weakPINMessage == "" {
		t.Fatal("weakPINMessage must name the policy so an operator can fix their PIN")
	}
	refusal := weakPINRefusal(domain.ErrWeakPIN)
	if refusal == nil {
		t.Fatal("a weak PIN must not be laundered into the generic credential refusal")
	}
	if got := posUsersStatus(refusal); got != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", got, http.StatusBadRequest)
	}
	var oneKitErr *onekithttp.Error
	if !errors.As(refusal, &oneKitErr) {
		t.Fatalf("returned %T, want *onekiterrors.Error", refusal)
	}
	if oneKitErr.GetMessage() != weakPINMessage {
		t.Errorf("body = %q, want %q", oneKitErr.GetMessage(), weakPINMessage)
	}
}
