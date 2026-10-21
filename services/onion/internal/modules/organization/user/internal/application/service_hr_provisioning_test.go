package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateBusinessForUserProvisionsHRRecords(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := newTestService(db, "")
	hrModule := hrcore.NewModule(ctx, db)
	service.WithBusinessProvisioner(hrModule)
	service.WithEmployeeProvisioner(hrModule)

	email := fmt.Sprintf("hr-provisioning+%d@example.com", time.Now().UnixNano())
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: "hash",
		FullName:     "HR Provisioning Owner",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	businessName := fmt.Sprintf("HR Provisioning Business %d", time.Now().UnixNano())
	var businessID int64
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		if _, cleanupErr := db.ExecContext(cleanupCtx,
			"DELETE FROM authorization_business_members WHERE business_id = ? AND user_id = ?",
			businessID, userID); cleanupErr != nil {
			t.Errorf("delete authorization membership: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(cleanupCtx,
			"DELETE FROM hr_employee_roles WHERE business_id = ?", businessID); cleanupErr != nil {
			t.Errorf("delete HR employee roles: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(cleanupCtx,
			"DELETE FROM hr_employees WHERE business_id = ?", businessID); cleanupErr != nil {
			t.Errorf("delete HR employees: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(cleanupCtx,
			"DELETE FROM hr_roles WHERE business_id = ?", businessID); cleanupErr != nil {
			t.Errorf("delete HR roles: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(cleanupCtx,
			"DELETE FROM hr_company_config WHERE business_id = ?", businessID); cleanupErr != nil {
			t.Errorf("delete HR company config: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(cleanupCtx,
			"DELETE FROM businesses WHERE id = ?", businessID); cleanupErr != nil {
			t.Errorf("delete business: %v", cleanupErr)
		}
		if cleanupErr := service.repo.DeleteUser(cleanupCtx, userID); cleanupErr != nil {
			t.Errorf("delete user: %v", cleanupErr)
		}
	})

	if err := service.CreateBusinessForUser(ctx, userID, businessName); err != nil {
		t.Fatalf("CreateBusinessForUser: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		"SELECT business_id FROM users WHERE id = ?", userID).Scan(&businessID); err != nil {
		t.Fatalf("read created business ID: %v", err)
	}

	var companyName string
	if err := db.QueryRowContext(ctx,
		"SELECT name FROM businesses WHERE id = ?", businessID).Scan(&companyName); err != nil {
		t.Fatalf("read organization business name: %v", err)
	}
	if companyName != businessName {
		t.Errorf("organization business name = %q, want %q", companyName, businessName)
	}

	var currency, timezone string
	if err := db.QueryRowContext(ctx, `
		SELECT default_currency, default_timezone
		FROM hr_company_config
		WHERE business_id = ?`, businessID).Scan(&currency, &timezone); err != nil {
		t.Fatalf("read provisioned HR config: %v", err)
	}
	if currency != "USD" {
		t.Errorf("HR default currency = %q, want USD", currency)
	}
	if timezone != "UTC" {
		t.Errorf("HR default timezone = %q, want UTC", timezone)
	}

	var roleCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hr_roles WHERE business_id = ?", businessID).Scan(&roleCount); err != nil {
		t.Fatalf("count provisioned HR roles: %v", err)
	}
	if roleCount < 5 {
		t.Fatalf("provisioned HR role count = %d, want at least 5", roleCount)
	}

	var employeeFirstName, employeeLastName, employeeEmail string
	if err := db.QueryRowContext(ctx, `
		SELECT first_name, last_name, email
		FROM hr_employees
		WHERE business_id = ? AND user_id = ?`, businessID, userID).Scan(&employeeFirstName, &employeeLastName, &employeeEmail); err != nil {
		t.Fatalf("read provisioned owner employee: %v", err)
	}
	if employeeFirstName != "HR" || employeeLastName != "Provisioning Owner" || employeeEmail != email {
		t.Fatalf("owner employee = %q %q <%s>, want %q %q <%s>", employeeFirstName, employeeLastName, employeeEmail,
			"HR", "Provisioning Owner", email)
	}

	var ownerAdminCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM hr_employee_roles er
		JOIN hr_roles r ON r.id = er.role_id
		JOIN hr_employees e ON e.id = er.employee_id
		WHERE er.business_id = ? AND e.user_id = ? AND r.name = 'HR Admin'`,
		businessID, userID).Scan(&ownerAdminCount); err != nil {
		t.Fatalf("count owner HR Admin assignments: %v", err)
	}
	if ownerAdminCount != 1 {
		t.Fatalf("owner HR Admin assignments = %d, want 1", ownerAdminCount)
	}
}
