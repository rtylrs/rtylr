package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/departments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
)

type fakeScopeResolver struct{}

func (fakeScopeResolver) ActiveBusinessOwner(_ context.Context, businessID int64) (int64, error) {
	return businessID, nil
}

func (fakeScopeResolver) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

type fakeDepartmentService struct {
	created domain.DepartmentInput
}

func (f *fakeDepartmentService) ListDepartments(_ context.Context, _ int64) ([]domain.Department, error) {
	return nil, nil
}

func (f *fakeDepartmentService) CreateDepartment(_ context.Context, input domain.DepartmentInput) (int64, error) {
	f.created = input
	return 55, nil
}

func (f *fakeDepartmentService) UpdateDepartment(_ context.Context, _ int64, _ domain.DepartmentInput) error {
	return nil
}

func (f *fakeDepartmentService) DeleteDepartment(_ context.Context, _, _ int64) error {
	return nil
}

type fakeDeptPermissions struct {
	empID   int64
	granted map[string]bool
}

func (f *fakeDeptPermissions) GetEmployeeIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	return f.empID, nil
}

func (f *fakeDeptPermissions) HasPermission(_ context.Context, _, _ int64, permissionKey string) (bool, error) {
	return f.granted[permissionKey], nil
}

func hrUserContext(userID, businessID int64, perms []string) context.Context {
	user := &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{
				{
					AppID:       authorization.AppHR,
					Licensed:    true,
					HasAccess:   true,
					Permissions: perms,
				},
			},
		},
	}
	dev := &authPkg.Device{User: user}
	return dev.Ctx(context.Background())
}

func TestCreateDepartment_UsesBusinessIDAsOwnerIDAndAllowsHRAdmin(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(fakeScopeResolver{}))
	svc := &fakeDepartmentService{}
	api := NewDepartmentAPI(svc, &fakeDeptPermissions{
		empID:   101,
		granted: map[string]bool{"manage_departments": true},
	})

	ctx := hrUserContext(34, 10, []string{"hr:employees:view"})
	resp, err := api.CreateDepartment(ctx, &hrdepartmentv1.CreateDepartmentRequest{Name: "Roastery"})
	if err != nil {
		t.Fatalf("CreateDepartment for HR-internal admin failed: %v", err)
	}
	if resp.GetId() != 55 {
		t.Fatalf("expected id 55, got %d", resp.GetId())
	}
	if svc.created.OwnerID != 10 || svc.created.BusinessID != 10 {
		t.Fatalf("expected owner_id=10 business_id=10, got owner_id=%d business_id=%d", svc.created.OwnerID, svc.created.BusinessID)
	}
}

func TestCreateDepartment_DeniesOrdinaryEmployee(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(fakeScopeResolver{}))
	svc := &fakeDepartmentService{}
	api := NewDepartmentAPI(svc, &fakeDeptPermissions{empID: 102, granted: map[string]bool{}})

	ctx := hrUserContext(35, 10, []string{"hr:employees:view", "hr:time_off:request"})
	if _, err := api.CreateDepartment(ctx, &hrdepartmentv1.CreateDepartmentRequest{Name: "Roastery"}); err == nil {
		t.Fatal("expected permission denied for ordinary employee")
	}
}
