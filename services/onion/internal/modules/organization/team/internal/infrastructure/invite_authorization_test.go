package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	hrstaff "github.com/voxire/onion/internal/modules/hr/staff"
	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

type hrMemberProvisioner struct {
	core  *hrcore.Module
	staff *hrstaff.Module
}

func (p hrMemberProvisioner) EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, firstName, lastName, email string) error {
	return p.core.EnsureEmployeeForUser(ctx, ownerID, businessID, userID, firstName, lastName, email)
}

func (p hrMemberProvisioner) EnsureStaffForUser(ctx context.Context, ownerID, businessID, userID int64, role string) error {
	return p.staff.EnsureStaffForUser(ctx, ownerID, businessID, userID, role)
}

func (p hrMemberProvisioner) DeleteProvisionedMember(ctx context.Context, ownerID, businessID, userID int64) error {
	if err := p.staff.DeleteStaffForUser(ctx, ownerID, businessID, userID); err != nil {
		return err
	}
	return p.core.DeleteEmployeeForUser(ctx, ownerID, businessID, userID)
}

func TestInviteDoesNotAssignPermissionProfilesByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	ctx := context.Background()
	service := application.NewService(ctx, infrastructure.NewRepository(db))
	provisioner := hrMemberProvisioner{
		core:  hrcore.NewModule(ctx, db),
		staff: hrstaff.NewModule(ctx, db, hrcore.NewModule(ctx, db)),
	}
	service.WithMemberProvisioner(provisioner)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("empty-profiles-invite-%d@test.local", time.Now().UnixNano())

	userID, _, err := service.InviteUser(ctx, testutil.OwnerID, testutil.BusinessID, domain.InviteUserInput{
		Email:          email,
		FullName:       "Ada Lovelace",
		LicensedAppIDs: []string{"hr", "pos"},
	})
	if err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	t.Cleanup(func() {
		if err := provisioner.DeleteProvisionedMember(ctx, testutil.OwnerID, testutil.BusinessID, userID); err != nil {
			t.Errorf("clean up HR member: %v", err)
		}
		if err := repo.DeleteInvitedUser(ctx, testutil.BusinessID, userID); err != nil {
			t.Errorf("clean up invited member: %v", err)
		}
	})

	var profiles, licenses int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authorization_role_assignments
		WHERE business_id = ? AND user_id = ?`, testutil.BusinessID, userID).Scan(&profiles); err != nil {
		t.Fatalf("count assigned profiles: %v", err)
	}
	if profiles != 0 {
		t.Fatalf("invite created %d implicit permission profiles, want none", profiles)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authorization_license_assignments
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, testutil.BusinessID, userID).Scan(&licenses); err != nil {
		t.Fatalf("count licensed apps: %v", err)
	}
	if licenses != 2 {
		t.Fatalf("pending app licenses = %d, want 2", licenses)
	}

	var employeeFirstName, employeeLastName, employeeEmail string
	if err := db.QueryRowContext(ctx, `
		SELECT first_name, last_name, email FROM hr_employees
		WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
		testutil.OwnerID, testutil.BusinessID, userID).Scan(&employeeFirstName, &employeeLastName, &employeeEmail); err != nil {
		t.Fatalf("read HR employee provisioned by the HR owner: %v", err)
	}
	if employeeFirstName != "Ada" || employeeLastName != "Lovelace" || employeeEmail != email {
		t.Fatalf("HR employee = %q %q <%s>", employeeFirstName, employeeLastName, employeeEmail)
	}
	var staffCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM hr_staff
		WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
		testutil.OwnerID, testutil.BusinessID, userID).Scan(&staffCount); err != nil {
		t.Fatalf("count HR staff directory rows: %v", err)
	}
	if staffCount != 1 {
		t.Fatalf("HR staff directory rows = %d, want 1", staffCount)
	}
}
