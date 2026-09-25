package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// nonOwnerService is the real service with business ownership forced off, so
// these tests exercise the permission lookups against MySQL rather than the
// owner bypass. The bypass needs the platform connection pool, which
// integration tests do not initialise.
type nonOwnerService struct{ *Service }

func (nonOwnerService) IsBusinessOwner(context.Context, int64) (bool, error) { return false, nil }

// openGatedService returns a service wired to the integration database, plus
// the handle the seed helper needs to clean its rows up.
func openGatedService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	db := testutil.OpenDB(t)
	return &Service{repo: infrastructure.NewRepository(db)}, db
}

// seedGatedEmployee creates an employee linked to userID and gives it roleName,
// returning the employee id. Rows are removed when the test finishes.
func seedGatedEmployee(t *testing.T, svc *Service, db *sql.DB, userID int64, firstName, roleName string) int64 {
	t.Helper()
	ctx := context.Background()

	if err := svc.SeedDefaultRoles(ctx, testutil.OwnerID); err != nil {
		t.Fatalf("SeedDefaultRoles: %v", err)
	}

	created, err := svc.CreateEmployee(ctx, testutil.OwnerID, domain.EmployeeRecord{
		UserID: userID, FirstName: firstName, LastName: "Gated",
		Email: firstName + ".gated@example.com", Status: "active", EmploymentType: "full_time",
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", created.ID)
	})

	roles, err := svc.ListRoles(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	for _, role := range roles {
		if role.Name != roleName {
			continue
		}
		if assignErr := svc.AssignRole(ctx, testutil.OwnerID, created.ID, role.ID); assignErr != nil {
			t.Fatalf("AssignRole: %v", assignErr)
		}
		return created.ID
	}

	t.Fatalf("seeded role %q not found", roleName)
	return 0
}

// The reproduction from issue #62, resolved against real role rows: a caller
// holding only the Employee role must not be able to write employee records.
func TestCanManage_AgainstMySQL_PlainEmployeeIsDenied(t *testing.T) {
	svc, db := openGatedService(t)
	const plainUserID int64 = 900201

	seedGatedEmployee(t, svc, db, plainUserID, "Plain", "Employee")

	if canManage(context.Background(), nonOwnerService{svc}, testutil.OwnerID, plainUserID, permissionManageEmployees) {
		t.Error("expected a caller holding only the Employee role to be denied")
	}
}

func TestCanManage_AgainstMySQL_HRAdminIsAllowed(t *testing.T) {
	svc, db := openGatedService(t)
	const adminUserID int64 = 900202

	seedGatedEmployee(t, svc, db, adminUserID, "Admin", "HR Admin")

	if !canManage(context.Background(), nonOwnerService{svc}, testutil.OwnerID, adminUserID, permissionManageEmployees) {
		t.Error("expected a caller holding HR Admin to be allowed")
	}
}

// The Employee role carries neither key, so the role gate that protects against
// self-promotion denies it too.
func TestCanManage_AgainstMySQL_PlainEmployeeCannotManageRoles(t *testing.T) {
	svc, db := openGatedService(t)
	const plainUserID int64 = 900203

	seedGatedEmployee(t, svc, db, plainUserID, "Norole", "Employee")

	if canManage(context.Background(), nonOwnerService{svc}, testutil.OwnerID, plainUserID, permissionManageRoles) {
		t.Error("expected a caller holding only the Employee role to be denied role management")
	}
}
