package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
)

// The pre-fix suite here asserted that VerifyPIN returned whatever the
// repository returned, verbatim and on every call. That pass-through *is* the
// defect: it is the behaviour that let an unbounded number of guesses reach the
// credential check. The lockout, uniform-failure and fail-closed properties are
// covered in pin_lockout_test.go, which exercises the same entry point.

func TestVerifyPIN_RejectsNonPositiveBusiness(t *testing.T) {
	for _, businessID := range []int64{0, -1} {
		verifier := &stubVerifier{candidates: candidatesFor(t, testValidPIN)}
		svc, _, _ := newScenario(verifier)

		result, err := svc.VerifyPIN(context.Background(), businessID, testValidPIN)
		if err == nil && result != nil && result.Valid {
			t.Errorf("businessID %d authenticated a manager; an unresolved tenant must never reach the credential check", businessID)
		}
		if verifier.reads != 0 {
			t.Errorf("businessID %d reached the credential read", businessID)
		}
	}
}

// A verified manager is attributed to the staff id the credential was stored
// under, which is what an override is audited against.
func TestVerifyPIN_ReturnsMatchedStaffIdentity(t *testing.T) {
	candidates := candidatesFor(t, "999111", "888222", testValidPIN)
	want := candidates[2].StaffID
	wantOwner := candidates[2].OwnerID
	svc, _, _ := newScenario(&stubVerifier{candidates: candidates})

	result, err := svc.VerifyPIN(context.Background(), testBusinessID, testValidPIN)
	if err != nil {
		t.Fatalf("VerifyPIN: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected a valid result")
	}
	if result.StaffID != want {
		t.Errorf("StaffID = %d, want %d: the override would be attributed to the wrong manager", result.StaffID, want)
	}
	if result.OwnerID != wantOwner {
		t.Errorf("OwnerID = %d, want %d", result.OwnerID, wantOwner)
	}
	if result.BusinessID != testBusinessID {
		t.Errorf("BusinessID = %d, want %d", result.BusinessID, testBusinessID)
	}
}

// The wrong manager's PIN must not authenticate even when a correct one is
// stored alongside it.
func TestVerifyPIN_RejectsAnotherManagersPIN(t *testing.T) {
	svc, _, _ := newScenario(&stubVerifier{candidates: candidatesFor(t, testValidPIN)})

	result, err := svc.VerifyPIN(context.Background(), testBusinessID, "222222")
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Fatalf("a PIN belonging to nobody returned %v / %+v, want ErrInvalidPIN", err, result)
	}
	if result != nil {
		t.Error("a failed verification returned a manager identity")
	}
}
