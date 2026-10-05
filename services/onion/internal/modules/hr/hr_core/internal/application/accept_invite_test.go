package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/scopes"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// createAcceptingAccount inserts the account an acceptance links to. An invite
// is addressed to an hr_employees email and only that account may claim it, so
// the account has to exist under the same address.
func createAcceptingAccount(t *testing.T, db *sql.DB, userID int64, email string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, 'not-a-real-hash')", userID, email); err != nil {
		t.Fatalf("create accepting account %d: %v", userID, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", userID)
	})
}

// Regression: AcceptInvite consumed the invite token before attempting to link
// the employee. When the link failed - the user_id already belonged to another
// employee row - the token stayed consumed while the employee stayed unlinked,
// so every retry with that token answered "Invalid invite token".
func TestAcceptInvite_FailedLinkLeavesTokenUsable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithAccountEmailReader(scopes.NewReader(db).AccountEmail)
	svc := &Service{repo: repo}
	ctx := context.Background()

	const (
		recipientUserID int64 = 981002
		blockingUserID  int64 = 981003
	)

	// The recipient's account is already on another employee row, so the link
	// must fail with ErrUserAlreadyLinked.
	blocking, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		UserID: recipientUserID, FirstName: "Already", LastName: "Linked",
		Email: "already.linked.981002@example.invalid", Status: "invited",
	})
	if err != nil {
		t.Fatalf("CreateEmployee (linked): %v", err)
	}

	target, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName: "Invite", LastName: "Target",
		Email: "invite.target.981002@example.invalid", Status: "invited",
	})
	if err != nil {
		t.Fatalf("CreateEmployee (target): %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employee_invites WHERE employee_id IN (?, ?)", blocking.ID, target.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id IN (?, ?)", blocking.ID, target.ID)
	})

	createAcceptingAccount(t, db, recipientUserID, "invite.target.981002@example.invalid")
	createAcceptingAccount(t, db, blockingUserID, "blocking.981003@example.invalid")

	token, err := repo.CreateInviteToken(ctx, testutil.OwnerID, target.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateInviteToken: %v", err)
	}

	if err := svc.AcceptInviteForCaller(ctx, token, recipientUserID); !errors.Is(err, domain.ErrUserAlreadyLinked) {
		t.Fatalf("AcceptInvite with an already-linked account = %v, want ErrUserAlreadyLinked", err)
	}

	// The failed accept must not burn the invite.
	invite, err := repo.GetInviteByToken(ctx, token)
	if err != nil {
		t.Fatalf("invite token was consumed by the failed link: %v", err)
	}
	if invite.EmployeeID != target.ID {
		t.Fatalf("invite employee id = %d, want %d", invite.EmployeeID, target.ID)
	}

	// The same token is still usable once the conflicting link is gone.
	if _, err := db.ExecContext(ctx, "UPDATE hr_employees SET user_id = NULL WHERE id = ?", blocking.ID); err != nil {
		t.Fatalf("release the blocking account: %v", err)
	}
	if err := svc.AcceptInviteForCaller(ctx, token, recipientUserID); err != nil {
		t.Fatalf("AcceptInvite retry with the preserved token: %v", err)
	}
	accepted, err := repo.GetEmployee(ctx, testutil.OwnerID, target.ID)
	if err != nil {
		t.Fatalf("GetEmployee after accept: %v", err)
	}
	if accepted.UserID != recipientUserID || accepted.Status != "active" {
		t.Fatalf("accepted employee user_id=%d status=%q, want user_id=%d status=active", accepted.UserID, accepted.Status, recipientUserID)
	}

	// A successful accept still consumes the token exactly once.
	if err := svc.AcceptInviteForCaller(ctx, token, recipientUserID); !errors.Is(err, domain.ErrInviteTokenInvalid) {
		t.Fatalf("second accept error = %v, want ErrInviteTokenInvalid", err)
	}
}

// The other link failure the fix covers: the employee row is gone, so the link
// matches no row. The invite must survive that too.
func TestAcceptInvite_MissingEmployeeLeavesTokenUsable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithAccountEmailReader(scopes.NewReader(db).AccountEmail)
	svc := &Service{repo: repo}
	ctx := context.Background()

	const userID int64 = 981011

	emp, err := repo.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName: "Vanished", LastName: "Employee",
		Email: "vanished.employee.981011@example.invalid", Status: "invited",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employee_invites WHERE employee_id = ?", emp.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id = ?", emp.ID)
	})

	createAcceptingAccount(t, db, userID, "vanished.employee.981011@example.invalid")

	token, err := repo.CreateInviteToken(ctx, testutil.OwnerID, emp.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateInviteToken: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id = ?", emp.ID); err != nil {
		t.Fatalf("delete employee row: %v", err)
	}

	if err := svc.AcceptInviteForCaller(ctx, token, userID); !errors.Is(err, domain.ErrEmployeeNotFound) {
		t.Fatalf("AcceptInvite for a deleted employee = %v, want ErrEmployeeNotFound", err)
	}

	if _, err := repo.GetInviteByToken(ctx, token); err != nil {
		t.Fatalf("invite token was consumed by the failed link: %v", err)
	}
}
