package application

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/scopes"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestAcceptInvite_PreservesInvitedEmployeeFields(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := &Service{repo: infrastructure.NewRepository(db).WithAccountEmailReader(scopes.NewReader(db).AccountEmail)}
	ctx := context.Background()

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	invited, err := svc.InviteEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		FirstName:             "Invited",
		LastName:              "Preserved",
		Email:                 "invite.preserved.847@example.com",
		Phone:                 "+1555000847",
		EmploymentType:        "part_time",
		StartDate:             startDate,
		SalaryCents:           725000,
		HourlyRateCents:       4200,
		Address:               "847 Preserve Street",
		EmergencyContactName:  "Emergency Contact",
		EmergencyContactPhone: "+1555000848",
		Notes:                 "Keep every field",
	})
	if err != nil {
		t.Fatalf("InviteEmployee: %v", err)
	}
	invite, err := svc.repo.GetInviteByToken(ctx, invited)
	if err != nil {
		t.Fatalf("GetInviteByToken: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employee_invites WHERE token = ?", invited)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", invite.EmployeeID)
	})

	const acceptedUserID int64 = 900847
	createAcceptingAccount(t, db, acceptedUserID, "invite.preserved.847@example.com")
	if err := svc.AcceptInviteForCaller(ctx, invited, acceptedUserID); err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}

	got, err := svc.repo.GetEmployee(ctx, testutil.OwnerID, invite.EmployeeID)
	if err != nil {
		t.Fatalf("GetEmployee after acceptance: %v", err)
	}
	if got.UserID != acceptedUserID || got.Status != "active" {
		t.Fatalf("accepted employee link/status = user %d, status %q; want user %d and active", got.UserID, got.Status, acceptedUserID)
	}
	if got.FirstName != "Invited" || got.LastName != "Preserved" || got.Email != "invite.preserved.847@example.com" || got.Phone != "+1555000847" {
		t.Fatalf("accepted employee identity fields were changed: %+v", got)
	}
	if got.EmploymentType != "part_time" || got.StartDate.Format("2006-01-02") != startDate.Format("2006-01-02") || got.SalaryCents != 725000 || got.HourlyRateCents != 4200 {
		t.Fatalf("accepted employee employment fields were changed: %+v", got)
	}
	if got.Address != "847 Preserve Street" || got.EmergencyContactName != "Emergency Contact" || got.EmergencyContactPhone != "+1555000848" || got.Notes != "Keep every field" {
		t.Fatalf("accepted employee contact fields were changed: %+v", got)
	}
}

func TestDeactivateEmployee_PreservesEmployeeFields(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := &Service{repo: infrastructure.NewRepository(db).WithAccountEmailReader(scopes.NewReader(db).AccountEmail)}
	ctx := context.Background()

	created, err := svc.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		UserID:                900848,
		FirstName:             "Terminated",
		LastName:              "Preserved",
		Email:                 "termination.preserved.847@example.com",
		Phone:                 "+1555000849",
		EmploymentType:        "full_time",
		Status:                "active",
		SalaryCents:           810000,
		Address:               "847 Offboard Street",
		EmergencyContactName:  "Emergency Contact",
		EmergencyContactPhone: "+1555000850",
		Notes:                 "Keep this too",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", created.ID)
	})

	if err := svc.DeactivateEmployee(ctx, testutil.OwnerID, created.ID); err != nil {
		t.Fatalf("DeactivateEmployee: %v", err)
	}

	got, err := svc.repo.GetEmployee(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetEmployee after deactivation: %v", err)
	}
	if got.Status != "terminated" || got.EndDate.IsZero() {
		t.Fatalf("termination state = status %q, end date %v; want terminated with an end date", got.Status, got.EndDate)
	}
	if got.UserID != 900848 || got.FirstName != "Terminated" || got.LastName != "Preserved" || got.Email != "termination.preserved.847@example.com" || got.Phone != "+1555000849" {
		t.Fatalf("termination changed employee identity fields: %+v", got)
	}
	if got.EmploymentType != "full_time" || got.SalaryCents != 810000 || got.Address != "847 Offboard Street" || got.EmergencyContactName != "Emergency Contact" || got.EmergencyContactPhone != "+1555000850" || got.Notes != "Keep this too" {
		t.Fatalf("termination changed employee details: %+v", got)
	}
}
