package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

// recordingTerminationRevoker captures the values the service forwards when an
// employee is terminated, so the test can prove the linked user id survived the
// delete-shaped write and immediate access revocation was not skipped.
type recordingTerminationRevoker struct {
	businessIDs []int64
	userIDs     []int64
}

type testBusinessOwner struct {
	userID int64
}

func (r testBusinessOwner) IsOwner(_ context.Context, userID int64) (bool, error) {
	return userID == r.userID, nil
}

func (r *recordingTerminationRevoker) RevokeAccessForTerminated(_ context.Context, businessID, userID int64) error {
	r.businessIDs = append(r.businessIDs, businessID)
	r.userIDs = append(r.userIDs, userID)
	return nil
}

// deleteEmployeeContext builds the request context DELETE /hr/employees/{id}
// runs under: an authenticated owner holding hr:employees:manage with a
// resolved HR app scope, and the chi route param the handler reads.
func deleteEmployeeContext(businessID, userID, employeeID int64) context.Context {
	user := &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
		ProvisioningStatus: &authorization.ProvisioningStatus{
			State:         authorization.ProvisioningActive,
			EmailVerified: true,
		},
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           businessID,
			IsOwner:              true,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"hr:employees:manage"},
			}},
		},
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(employeeID, 10))
	return authPkg.WithAuthUser(context.WithValue(context.Background(), chi.RouteCtxKey, rctx), user)
}

// seedTerminationEmployee inserts a fully populated employee row so the
// termination write has every column to lose.
func seedTerminationEmployee(t *testing.T, db *sql.DB, id, businessID, userID int64) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO hr_employees
			(id, owner_id, business_id, user_id, department_id, location_id, work_schedule_id,
			 first_name, last_name, email, phone, employment_type, status,
			 salary, hourly_rate, overtime_rate, start_date, end_date,
			 address, emergency_contact_name, emergency_contact_phone, notes)
		 VALUES (?, ?, ?, ?, 12, 34, 56,
			 'Grace', 'Hopper', ?, '+1555000983', 'part_time', 'active',
			 9876.54, 43.21, 65.43, '2024-03-01', NULL,
			 '1 Compiler Way', 'Ada Lovelace', '+1555000984', 'Preserve every column')`,
		id, businessID, businessID, userID, "grace.hopper.983@example.invalid",
	)
	if err != nil {
		t.Fatalf("seed hr_employees row (%d): %v", id, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM hr_employees WHERE id = ?`, id)
	})
}

// Regression: DeleteEmployeeHandler used to call the full-row
// UpdateEmployeeForWriter with a zero-filled EmployeeRecord{ID, Status}, which
// blanked every other column and nulled user_id before the termination
// revocation could resolve it. Terminating must change only status and
// end_date, and the linked user must still be revoked immediately.
func TestDeleteEmployeeHandler_TerminationPreservesEmployeeFields(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(983001)
		businessID = int64(983002)
		employeeID = int64(983003)
		linkedUser = int64(983004)
	)

	testutil.GrantBusinessAccess(t, db, businessID, userID)
	seedTerminationEmployee(t, db, employeeID, businessID, linkedUser)

	svc := application.NewService(ctx, infrastructure.NewRepository(db)).WithBusinessOwnerResolver(testBusinessOwner{userID: userID})
	revoker := &recordingTerminationRevoker{}
	svc.SetTerminationRevoker(revoker)

	handler := &DeleteEmployeeHandler{Service: svc}
	resp, err := handler.Serve(deleteEmployeeContext(businessID, userID, employeeID), &DeleteEmployeeRequest{})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected Success")
	}

	got, err := infrastructure.NewRepository(db).GetEmployee(ctx, businessID, employeeID)
	if err != nil {
		t.Fatalf("GetEmployee after termination: %v", err)
	}
	if got.Status != "terminated" {
		t.Fatalf("Status = %q, want terminated", got.Status)
	}
	if got.EndDate.IsZero() {
		t.Fatal("EndDate was not set on termination")
	}
	if got.UserID != linkedUser {
		t.Fatalf("UserID = %d, want %d; termination must not unlink the user before access revocation", got.UserID, linkedUser)
	}
	if got.FirstName != "Grace" || got.LastName != "Hopper" || got.Email != "grace.hopper.983@example.invalid" || got.Phone != "+1555000983" {
		t.Fatalf("termination changed identity fields: %+v", got)
	}
	if got.DepartmentID != 12 || got.LocationID != 34 || got.WorkScheduleID != 56 {
		t.Fatalf("termination changed assignment fields: %+v", got)
	}
	if got.EmploymentType != "part_time" || got.SalaryCents != 987654 || got.HourlyRateCents != 4321 || got.OvertimeRateCents != 6543 {
		t.Fatalf("termination changed employment/compensation fields: %+v", got)
	}
	if got.StartDate.Format("2006-01-02") != "2024-03-01" {
		t.Fatalf("StartDate = %q, want 2024-03-01", got.StartDate.Format("2006-01-02"))
	}
	if got.Address != "1 Compiler Way" || got.EmergencyContactName != "Ada Lovelace" || got.EmergencyContactPhone != "+1555000984" || got.Notes != "Preserve every column" {
		t.Fatalf("termination changed contact fields: %+v", got)
	}

	if len(revoker.userIDs) != 1 || revoker.userIDs[0] != linkedUser {
		t.Fatalf("immediate access revocation user ids = %v, want [%d]", revoker.userIDs, linkedUser)
	}
	if len(revoker.businessIDs) != 1 || revoker.businessIDs[0] != businessID {
		t.Fatalf("immediate access revocation business ids = %v, want [%d]", revoker.businessIDs, businessID)
	}
}

// A delete for an employee id that does not exist must report 404 instead of
// the success the zero-row full-row update used to return.
func TestDeleteEmployeeHandler_MissingEmployeeReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(983011)
		businessID = int64(983012)
	)

	testutil.GrantBusinessAccess(t, db, businessID, userID)

	svc := application.NewService(ctx, infrastructure.NewRepository(db)).WithBusinessOwnerResolver(testBusinessOwner{userID: userID})
	handler := &DeleteEmployeeHandler{Service: svc}
	resp, err := handler.Serve(deleteEmployeeContext(businessID, userID, 999999983), &DeleteEmployeeRequest{})

	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusNotFound {
		t.Fatalf("expected 404 for a missing employee, got response=%+v err=%v", resp, err)
	}
	if resp.Success {
		t.Fatal("a missing employee must not report success")
	}
}
