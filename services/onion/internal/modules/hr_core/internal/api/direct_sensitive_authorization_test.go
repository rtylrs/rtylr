package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestDirectSensitiveAuthorization(t *testing.T) {
	employee := domain.EmployeeRecord{SalaryCents: 125000, HourlyRateCents: 2500, OvertimeRateCents: 3750}
	withoutSensitive := &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppHR, Licensed: true, HasAccess: true, Permissions: []string{"hr:employees:view"},
	}}}}
	redacted := employeeToProto(employee, authPkg.HasAppPermission(withoutSensitive, authorization.AppHR, "hr:employees:sensitive_view"))
	if redacted.Salary != 0 || redacted.HourlyRate != 0 || redacted.OvertimeRate != 0 {
		t.Fatalf("sensitive employee fields leaked without permission: %+v", redacted)
	}

	withSensitive := &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppHR, Licensed: true, HasAccess: true, Permissions: []string{"hr:employees:view", "hr:employees:sensitive_view"},
	}}}}
	visible := employeeToProto(employee, authPkg.HasAppPermission(withSensitive, authorization.AppHR, "hr:employees:sensitive_view"))
	if visible.Salary == 0 || visible.HourlyRate == 0 || visible.OvertimeRate == 0 {
		t.Fatalf("sensitive employee fields were not returned with permission: %+v", visible)
	}
}
