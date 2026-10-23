package application_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Ids sit above the fixture range and the grant tests' 9_999_930 block.
const (
	deleteTestOwnerID      int64 = 9_999_940
	deleteTestBusinessID   int64 = 9_999_941
	deleteTestMemberID     int64 = 9_999_942
	deleteTestSoloOwnerID  int64 = 9_999_943
	deleteTestSoloBizID    int64 = 9_999_944
	deleteTestAdminID      int64 = 9_999_945
	deleteTestBilledUserID int64 = 9_999_946
)

type seatRecorder struct{ owners []int64 }

func (r *seatRecorder) RecordSeatChange(_ context.Context, ownerID int64) error {
	r.owners = append(r.owners, ownerID)
	return nil
}

func TestDeleteUser_ErasesMemberAccessAndPersonalData(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	purgeDeleteUserFixtures(t, db)

	seedDeleteTestUser(t, db, deleteTestOwnerID, false)
	seedDeleteTestUser(t, db, deleteTestMemberID, false)
	mustExec(t, db, `INSERT INTO businesses (id, name, owner_id) VALUES (?, 'Delete Test Shared', ?)`, deleteTestBusinessID, deleteTestOwnerID)
	mustExec(t, db, `INSERT INTO authorization_business_members (business_id, user_id, status) VALUES (?, ?, 'active')`, deleteTestBusinessID, deleteTestMemberID)
	mustExec(t, db, `INSERT INTO authorization_license_assignments (business_id, user_id, app_id) VALUES (?, ?, 'dash')`, deleteTestBusinessID, deleteTestMemberID)
	roleID := mustInsert(t, db, `INSERT INTO authorization_roles (business_id, role_key, display_name) VALUES (?, 'delete_test', 'Delete test')`, deleteTestBusinessID)
	assignmentID := mustInsert(t, db, `INSERT INTO authorization_role_assignments (business_id, user_id, app_id, role_id) VALUES (?, ?, 'pos', ?)`, deleteTestBusinessID, deleteTestMemberID, roleID)
	mustExec(t, db, `INSERT INTO authorization_role_assignment_scopes (assignment_id, scope_type, scope_id) VALUES (?, 'pos_location', '1')`, assignmentID)
	mustExec(t, db, `INSERT INTO user_sessions (user_id, token_hash) VALUES (?, 'delete-test-session')`, deleteTestMemberID)
	mustExec(t, db, `INSERT INTO login_history (user_id, email) VALUES (?, 'member@delete.test')`, deleteTestMemberID)
	mustExec(t, db, `INSERT INTO chat_sessions (user_id, session_id) VALUES (?, 'delete-test-chat')`, deleteTestMemberID)
	mustExec(t, db, `INSERT INTO conversation_messages (session_id, user_id, role, content) VALUES ('delete-test-chat', ?, 'user', 'hello')`, deleteTestMemberID)
	mustExec(t, db, "INSERT INTO user_memory (user_id, memory_type, `key`, value) VALUES (?, 'fact', 'name', 'Ada')", deleteTestMemberID)
	mustExec(t, db, `INSERT INTO hub_notifications (owner_id, user_id, title) VALUES (?, ?, 'Hi')`, deleteTestOwnerID, deleteTestMemberID)
	mustExec(t, db, `INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email) VALUES (?, ?, ?, 'Ada', 'Lovelace', 'member@delete.test')`, deleteTestBusinessID, deleteTestBusinessID, deleteTestMemberID)
	mustExec(t, db, `INSERT INTO pos_users (owner_id, business_id, user_id, location_id, name) VALUES (?, ?, ?, 1, 'Ada')`, deleteTestOwnerID, deleteTestBusinessID, deleteTestMemberID)

	recorder := &seatRecorder{}
	svc := application.NewService(ctx, infrastructure.NewRepository(db)).WithSeatChangeRecorder(recorder)
	deleted, err := svc.DeleteUser(ctx, 0, deleteTestMemberID)
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	if len(deleted.DeactivatedBusinessIDs) != 0 {
		t.Errorf("deactivated %v, want none", deleted.DeactivatedBusinessIDs)
	}
	if want := []int64{deleteTestBusinessID}; !reflect.DeepEqual(deleted.MemberBusinessIDs, want) {
		t.Errorf("member businesses = %v, want %v", deleted.MemberBusinessIDs, want)
	}
	if want := []int64{deleteTestOwnerID}; !reflect.DeepEqual(recorder.owners, want) {
		t.Errorf("seat changes recorded for %v, want %v", recorder.owners, want)
	}
	for _, check := range []struct{ table, where string }{
		{"users", "id = ?"},
		{"authorization_business_members", "user_id = ?"},
		{"authorization_license_assignments", "user_id = ?"},
		{"authorization_role_assignments", "user_id = ?"},
		{"user_sessions", "user_id = ?"},
		{"login_history", "user_id = ?"},
		{"chat_sessions", "user_id = ?"},
		{"conversation_messages", "user_id = ?"},
		{"user_memory", "user_id = ?"},
		{"hub_notifications", "user_id = ?"},
	} {
		if n := countRows(t, db, check.table, check.where, deleteTestMemberID); n != 0 {
			t.Errorf("%s still has %d rows for the deleted user", check.table, n)
		}
	}
	if n := countRows(t, db, "authorization_role_assignment_scopes", "assignment_id = ?", assignmentID); n != 0 {
		t.Errorf("role assignment scopes left behind: %d", n)
	}
	if n := countRows(t, db, "hr_employees", "business_id = ? AND user_id IS NULL", deleteTestBusinessID); n != 1 {
		t.Errorf("hr_employees unlinked rows = %d, want the record kept without its user", n)
	}
	if n := countRows(t, db, "pos_users", "business_id = ? AND is_active = FALSE", deleteTestBusinessID); n != 1 {
		t.Errorf("deactivated pos_users rows = %d, want 1", n)
	}
	if n := countRows(t, db, "businesses", "id = ? AND is_active = TRUE", deleteTestBusinessID); n != 1 {
		t.Error("the member's deletion must not deactivate someone else's business")
	}
}

func TestDeleteUser_RefusesOwnerWhileOtherMembersRemain(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	purgeDeleteUserFixtures(t, db)

	seedDeleteTestUser(t, db, deleteTestOwnerID, false)
	seedDeleteTestUser(t, db, deleteTestMemberID, false)
	mustExec(t, db, `INSERT INTO businesses (id, name, owner_id) VALUES (?, 'Delete Test Shared', ?)`, deleteTestBusinessID, deleteTestOwnerID)
	mustExec(t, db, `INSERT INTO authorization_business_members (business_id, user_id, status) VALUES (?, ?, 'active')`, deleteTestBusinessID, deleteTestMemberID)

	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	_, err := svc.DeleteUser(ctx, 0, deleteTestOwnerID)
	var inUse *application.OwnedBusinessesInUseError
	if !errors.As(err, &inUse) {
		t.Fatalf("DeleteUser error = %v, want OwnedBusinessesInUseError", err)
	}
	if len(inUse.Businesses) != 1 || inUse.Businesses[0].ID != deleteTestBusinessID {
		t.Errorf("businesses in use = %+v", inUse.Businesses)
	}
	if n := countRows(t, db, "users", "id = ?", deleteTestOwnerID); n != 1 {
		t.Error("a refused deletion must leave the owner in place")
	}
	if n := countRows(t, db, "businesses", "id = ? AND is_active = TRUE", deleteTestBusinessID); n != 1 {
		t.Error("a refused deletion must leave the business active")
	}
}

func TestDeleteUser_DeactivatesSoloBusinessAndCancelsUnbilledPlan(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	purgeDeleteUserFixtures(t, db)

	seedDeleteTestUser(t, db, deleteTestSoloOwnerID, false)
	mustExec(t, db, `INSERT INTO businesses (id, name, owner_id) VALUES (?, 'Delete Test Solo', ?)`, deleteTestSoloBizID, deleteTestSoloOwnerID)
	mustExec(t, db, `INSERT INTO authorization_business_members (business_id, user_id, member_type, status) VALUES (?, ?, 'admin', 'active')`, deleteTestSoloBizID, deleteTestSoloOwnerID)
	mustExec(t, db, `INSERT INTO subscriptions (owner_id, plan_id, status) VALUES (?, 1, 'trial')`, deleteTestSoloOwnerID)

	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	deleted, err := svc.DeleteUser(ctx, 0, deleteTestSoloOwnerID)
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if want := []int64{deleteTestSoloBizID}; !reflect.DeepEqual(deleted.DeactivatedBusinessIDs, want) {
		t.Errorf("deactivated %v, want %v", deleted.DeactivatedBusinessIDs, want)
	}
	if n := countRows(t, db, "businesses", "id = ? AND is_active = FALSE", deleteTestSoloBizID); n != 1 {
		t.Error("the solo business should be deactivated")
	}
	if n := countRows(t, db, "subscriptions", "owner_id = ? AND status = 'cancelled'", deleteTestSoloOwnerID); n != 1 {
		t.Error("the unbilled trial should be cancelled")
	}
	if n := countRows(t, db, "users", "id = ?", deleteTestSoloOwnerID); n != 0 {
		t.Error("the owner should be deleted")
	}
}

func TestDeleteUser_RefusesWhilePaddleCanStillBill(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	purgeDeleteUserFixtures(t, db)

	seedDeleteTestUser(t, db, deleteTestBilledUserID, false)
	mustExec(t, db, `INSERT INTO subscriptions (owner_id, plan_id, status, paddle_subscription_id) VALUES (?, 1, 'active', 'sub_delete_test')`, deleteTestBilledUserID)

	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	if _, err := svc.DeleteUser(ctx, 0, deleteTestBilledUserID); !errors.Is(err, application.ErrUserHasBilledPlan) {
		t.Fatalf("DeleteUser error = %v, want ErrUserHasBilledPlan", err)
	}
	if n := countRows(t, db, "users", "id = ?", deleteTestBilledUserID); n != 1 {
		t.Error("a refused deletion must leave the user in place")
	}
}

func TestDeleteUser_RefusesSelfAndConsoleAdmins(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	purgeDeleteUserFixtures(t, db)

	seedDeleteTestUser(t, db, deleteTestAdminID, true)
	svc := application.NewService(ctx, infrastructure.NewRepository(db))

	if _, err := svc.DeleteUser(ctx, deleteTestAdminID, deleteTestAdminID); !errors.Is(err, application.ErrDeleteSelf) {
		t.Errorf("self delete error = %v, want ErrDeleteSelf", err)
	}
	if _, err := svc.DeleteUser(ctx, 0, deleteTestAdminID); !errors.Is(err, application.ErrDeleteConsoleAdmin) {
		t.Errorf("admin delete error = %v, want ErrDeleteConsoleAdmin", err)
	}
	if _, err := svc.DeleteUser(ctx, 0, 9_999_999); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("missing user error = %v, want sql.ErrNoRows", err)
	}
}

func seedDeleteTestUser(t *testing.T, db *sql.DB, id int64, consoleAdmin bool) {
	t.Helper()
	mustExec(t, db, `INSERT INTO users (id, email, password_hash, is_console_admin) VALUES (?, ?, 'x', ?)`,
		id, fmt.Sprintf("delete-test-%d@example.test", id), consoleAdmin)
}

// purgeDeleteUserFixtures clears leftovers before a test and everything it
// seeded afterwards, whether or not the deletion ran.
func purgeDeleteUserFixtures(t *testing.T, db *sql.DB) {
	t.Helper()
	users := []any{deleteTestOwnerID, deleteTestMemberID, deleteTestSoloOwnerID, deleteTestAdminID, deleteTestBilledUserID}
	businesses := []any{deleteTestBusinessID, deleteTestSoloBizID}
	purge := func() {
		ctx := context.Background()
		for _, statement := range []string{
			`DELETE scope FROM authorization_role_assignment_scopes scope JOIN authorization_role_assignments assignment ON assignment.id = scope.assignment_id WHERE assignment.business_id IN (?, ?)`,
			`DELETE FROM authorization_role_assignments WHERE business_id IN (?, ?)`,
			`DELETE FROM authorization_roles WHERE business_id IN (?, ?)`,
			`DELETE FROM authorization_license_assignments WHERE business_id IN (?, ?)`,
			`DELETE FROM authorization_business_members WHERE business_id IN (?, ?)`,
			`DELETE FROM hr_employees WHERE business_id IN (?, ?)`,
			`DELETE FROM pos_users WHERE business_id IN (?, ?)`,
			`DELETE FROM businesses WHERE id IN (?, ?)`,
		} {
			_, _ = db.ExecContext(ctx, statement, businesses...)
		}
		for _, statement := range []string{
			`DELETE FROM subscriptions WHERE owner_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM user_sessions WHERE user_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM login_history WHERE user_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM conversation_messages WHERE user_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM chat_sessions WHERE user_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM user_memory WHERE user_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM hub_notifications WHERE user_id IN (?, ?, ?, ?, ?)`,
			`DELETE FROM users WHERE id IN (?, ?, ?, ?, ?)`,
		} {
			_, _ = db.ExecContext(ctx, statement, users...)
		}
	}
	purge()
	t.Cleanup(purge)
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("seed %q: %v", query, err)
	}
}

func mustInsert(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()
	result, err := db.ExecContext(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("seed %q: %v", query, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("seed %q: %v", query, err)
	}
	return id
}

func countRows(t *testing.T, db *sql.DB, table, where string, args ...any) int {
	t.Helper()
	var n int
	// #nosec G201 -- table and where are test constants.
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}
