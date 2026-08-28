package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
	"github.com/voxire/onion/internal/testutil"
)

// Regression for #938: the handler used to pass the authenticated user's id as
// the owner argument, but the repository update is scoped by
// hr_employees.business_id. With distinct user and business ids the UPDATE
// matched zero rows and the endpoint failed for every normal employee.
func TestSetClockInPIN_ScopesByBusinessNotUserID(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(915001)
		businessID = int64(915002) // deliberately different from the user id
		employeeID = int64(915003)
	)

	testutil.GrantBusinessAccess(t, db, businessID, userID)
	seedAttendanceEmployee(t, db, employeeID, businessID, userID)

	handler := &SetClockInPINHandler{Service: application.NewService(ctx, db)}
	resp, err := handler.Serve(
		authPkg.WithAuthUser(ctx, &authPkg.AuthUser{UserID: userID, BusinessID: businessID}),
		&SetClockInPINRequest{EmployeeID: employeeID, Pin: "1234"},
	)
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success")
	}

	var hash *string
	if err := db.QueryRowContext(ctx,
		`SELECT clock_in_pin_hash FROM hr_employees WHERE id = ? AND business_id = ?`,
		employeeID, businessID,
	).Scan(&hash); err != nil {
		t.Fatalf("read back pin hash: %v", err)
	}
	if hash == nil || *hash == "" {
		t.Fatal("clock_in_pin_hash was not written")
	}
	if !passwordutils.CheckPINHash("1234", *hash) {
		t.Fatal("stored hash does not verify against the submitted PIN")
	}
}

// A caller whose token carries no usable business scope is denied rather than
// writing under a zeroed tenant bucket.
func TestSetClockInPIN_DeniesWithoutBusinessScope(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(916001)
		businessID = int64(916002)
	)

	testutil.GrantBusinessAccess(t, db, businessID, userID)

	handler := &SetClockInPINHandler{Service: application.NewService(ctx, db)}
	_, err := handler.Serve(
		authPkg.WithAuthUser(ctx, &authPkg.AuthUser{UserID: userID, BusinessID: 0}),
		&SetClockInPINRequest{EmployeeID: 1, Pin: "1234"},
	)

	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusForbidden {
		t.Fatalf("expected 403, got %v", err)
	}
}

// The PIN write targets hr_core's table from this module's repository; seed the
// row the update is scoped by with only the NOT NULL columns set.
func seedAttendanceEmployee(tb testing.TB, db *sql.DB, id, businessID, userID int64) {
	tb.Helper()

	_, err := db.Exec(
		`INSERT INTO hr_employees (id, owner_id, business_id, user_id, first_name, last_name, email)
		 VALUES (?, ?, ?, ?, 'Pin', 'Test', ?)`,
		id, businessID, businessID, userID, "pin-test@example.invalid",
	)
	if err != nil {
		tb.Fatalf("seed hr_employees row (%d): %v", id, err)
	}
	tb.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM hr_employees WHERE id = ?`, id)
	})
}
