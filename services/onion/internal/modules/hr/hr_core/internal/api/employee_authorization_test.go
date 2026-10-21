package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
)

func TestEmployeeToProtoRedactsCompensationWithoutSensitivePermission(t *testing.T) {
	employee := domain.EmployeeRecord{
		ID:                41,
		FirstName:         "Ada",
		LastName:          "Lovelace",
		SalaryCents:       125000,
		HourlyRateCents:   2500,
		OvertimeRateCents: 3750,
	}

	public := employeeToProto(employee, false)
	if public.Salary != 0 || public.HourlyRate != 0 || public.OvertimeRate != 0 {
		t.Fatalf("compensation leaked without sensitive permission: %+v", public)
	}

	sensitive := employeeToProto(employee, true)
	if sensitive.Salary != 1250 || sensitive.HourlyRate != 25 || sensitive.OvertimeRate != 37.5 {
		t.Fatalf("compensation was not returned with sensitive permission: %+v", sensitive)
	}
}
