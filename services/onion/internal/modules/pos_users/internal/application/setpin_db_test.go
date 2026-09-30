package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// Fixture context (scripts/fixtures/20_pos.sql): owner 1001, location 3001,
// members 1002 (Cashier A) and 1003 (Server B) — both assigned via
// pos_member_locations with can_manage_pos_users = FALSE and both already
// holding a PIN.
const (
	fixtureOwnerID    int64 = 1001
	fixtureLocationID int64 = 3001
	cashierAUserID    int64 = 1002
	serverBUserID     int64 = 1003
)

func newDBService(t *testing.T) *application.Service {
	t.Helper()
	db := testutil.OpenDB(t)
	return application.NewService(context.Background(), db)
}

// A non-manager session must not set another operator's PIN — guards against
// privilege escalation (e.g. claiming a manager's identity).
func TestSetPIN_CrossUser_NonManager_Forbidden(t *testing.T) {
	s := newDBService(t)
	err := s.SetPIN(context.Background(), fixtureOwnerID,
		application.SetPINActor{UserID: serverBUserID, IsOwner: false, TeamRole: "staff"},
		domain.SetPINInput{UserID: cashierAUserID, LocationID: fixtureLocationID, Name: "Cashier A", PIN: "432165"},
	)
	if !errors.Is(err, application.ErrForbidden) {
		t.Fatalf("expected ErrForbidden setting another operator's PIN, got %v", err)
	}
}

// set-pin is onboarding only: it must never silently overwrite an existing PIN
// (guards against rotation hijack). Cashier A already has a PIN in fixtures.
func TestSetPIN_SelfRotation_Rejected(t *testing.T) {
	s := newDBService(t)
	err := s.SetPIN(context.Background(), fixtureOwnerID,
		application.SetPINActor{UserID: cashierAUserID, IsOwner: false, TeamRole: "staff"},
		domain.SetPINInput{UserID: cashierAUserID, LocationID: fixtureLocationID, Name: "Cashier A", PIN: "000042"},
	)
	if !errors.Is(err, application.ErrPINAlreadySet) {
		t.Fatalf("expected ErrPINAlreadySet overwriting an existing PIN, got %v", err)
	}
}

// An operator not assigned to the location cannot hold a PIN there.
func TestSetPIN_Unassigned_Rejected(t *testing.T) {
	s := newDBService(t)
	err := s.SetPIN(context.Background(), fixtureOwnerID,
		application.SetPINActor{UserID: 777001, IsOwner: false, TeamRole: "staff"},
		domain.SetPINInput{UserID: 777001, LocationID: fixtureLocationID, Name: "Ghost", PIN: "121242"},
	)
	if !errors.Is(err, application.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for unassigned operator, got %v", err)
	}
}
