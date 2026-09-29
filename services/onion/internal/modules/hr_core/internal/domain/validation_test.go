package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

func TestValidateEmployee_ValidRecord(t *testing.T) {
	emp := domain.EmployeeRecord{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		SalaryCents: 100000,
	}
	if err := domain.ValidateEmployee(emp); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateEmployee_EmptyFirstName(t *testing.T) {
	emp := domain.EmployeeRecord{
		FirstName: "",
		LastName:  "Smith",
		Email:     "alice@example.com",
	}
	if err := domain.ValidateEmployee(emp); !errors.Is(err, domain.ErrEmptyEmployeeName) {
		t.Errorf("expected ErrEmptyEmployeeName, got %v", err)
	}
}

func TestValidateEmployee_EmptyLastName(t *testing.T) {
	emp := domain.EmployeeRecord{
		FirstName: "Alice",
		LastName:  "",
		Email:     "alice@example.com",
	}
	if err := domain.ValidateEmployee(emp); !errors.Is(err, domain.ErrEmptyEmployeeName) {
		t.Errorf("expected ErrEmptyEmployeeName, got %v", err)
	}
}

func TestValidateEmployee_WhitespaceOnlyName(t *testing.T) {
	emp := domain.EmployeeRecord{
		FirstName: "   ",
		LastName:  "Smith",
		Email:     "alice@example.com",
	}
	if err := domain.ValidateEmployee(emp); !errors.Is(err, domain.ErrEmptyEmployeeName) {
		t.Errorf("expected ErrEmptyEmployeeName for whitespace-only name, got %v", err)
	}
}

func TestValidateEmployee_EmptyEmail(t *testing.T) {
	emp := domain.EmployeeRecord{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "",
	}
	if err := domain.ValidateEmployee(emp); !errors.Is(err, domain.ErrEmptyEmail) {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}

func TestValidateEmployee_NegativeSalary(t *testing.T) {
	emp := domain.EmployeeRecord{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		SalaryCents: -1,
	}
	if err := domain.ValidateEmployee(emp); !errors.Is(err, domain.ErrInvalidSalary) {
		t.Errorf("expected ErrInvalidSalary, got %v", err)
	}
}

func TestValidateDepartment_EmptyName(t *testing.T) {
	dept := domain.DepartmentRecord{Name: ""}
	if err := domain.ValidateDepartment(dept); !errors.Is(err, domain.ErrEmptyDeptName) {
		t.Errorf("expected ErrEmptyDeptName, got %v", err)
	}
}

func TestValidateDepartment_ValidName(t *testing.T) {
	dept := domain.DepartmentRecord{Name: "Engineering"}
	if err := domain.ValidateDepartment(dept); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateLocation_EmptyName(t *testing.T) {
	loc := domain.LocationRecord{Name: ""}
	if err := domain.ValidateLocation(loc); !errors.Is(err, domain.ErrEmptyLocationName) {
		t.Errorf("expected ErrEmptyLocationName, got %v", err)
	}
}

func TestValidateLocation_ValidName(t *testing.T) {
	loc := domain.LocationRecord{Name: "HQ"}
	if err := domain.ValidateLocation(loc); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateRole_EmptyName(t *testing.T) {
	role := domain.RoleRecord{Name: ""}
	if err := domain.ValidateRole(role); !errors.Is(err, domain.ErrEmptyRoleName) {
		t.Errorf("expected ErrEmptyRoleName, got %v", err)
	}
}

func TestValidateRole_ValidName(t *testing.T) {
	role := domain.RoleRecord{Name: "Manager"}
	if err := domain.ValidateRole(role); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateWorkSchedule_EmptyName(t *testing.T) {
	ws := domain.WorkScheduleRecord{Name: ""}
	if err := domain.ValidateWorkSchedule(ws); !errors.Is(err, domain.ErrEmptyScheduleName) {
		t.Errorf("expected ErrEmptyScheduleName, got %v", err)
	}
}

func TestValidateWorkSchedule_ValidName(t *testing.T) {
	ws := domain.WorkScheduleRecord{Name: "9-5 Standard"}
	if err := domain.ValidateWorkSchedule(ws); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSameEmailAddress_RecipientComparison(t *testing.T) {
	cases := []struct {
		name     string
		invited  string
		caller   string
		wantSame bool
	}{
		{name: "identical", invited: "invitee@example.com", caller: "invitee@example.com", wantSame: true},
		{name: "different case and padding", invited: " Invitee@Example.com ", caller: "invitee@example.com", wantSame: true},
		{name: "another mailbox", invited: "invitee@example.com", caller: "attacker@example.com"},
		{name: "invite without an address", invited: "", caller: "invitee@example.com"},
		{name: "account without an address", invited: "invitee@example.com", caller: ""},
	}
	for _, tc := range cases {
		if got := domain.SameEmailAddress(tc.invited, tc.caller); got != tc.wantSame {
			t.Errorf("%s: SameEmailAddress(%q, %q) = %v, want %v", tc.name, tc.invited, tc.caller, got, tc.wantSame)
		}
	}
}

func TestIsValidEmploymentStatus_Active(t *testing.T) {
	if !domain.IsValidEmploymentStatus(domain.EmploymentStatusActive) {
		t.Error("expected 'active' to be valid")
	}
}

func TestIsValidEmploymentStatus_Inactive(t *testing.T) {
	if !domain.IsValidEmploymentStatus(domain.EmploymentStatusInactive) {
		t.Error("expected 'inactive' to be valid")
	}
}

func TestIsValidEmploymentStatus_Invited(t *testing.T) {
	if !domain.IsValidEmploymentStatus(domain.EmploymentStatusInvited) {
		t.Error("expected 'invited' to be valid")
	}
}

func TestIsValidEmploymentStatus_Invalid(t *testing.T) {
	if domain.IsValidEmploymentStatus("fired") {
		t.Error("expected 'fired' to be invalid")
	}
}

func TestIsValidEmploymentType_AllValid(t *testing.T) {
	for _, et := range []string{
		domain.EmploymentTypeFullTime,
		domain.EmploymentTypePartTime,
		domain.EmploymentTypeContract,
		domain.EmploymentTypeFreelance,
	} {
		if !domain.IsValidEmploymentType(et) {
			t.Errorf("expected %q to be valid", et)
		}
	}
}

func TestIsValidEmploymentType_Invalid(t *testing.T) {
	if domain.IsValidEmploymentType("gig") {
		t.Error("expected 'gig' to be invalid")
	}
}
