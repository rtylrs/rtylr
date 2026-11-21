package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/departments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
)

type departmentService interface {
	ListDepartments(ctx context.Context, businessID int64) ([]domain.Department, error)
	CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error)
	UpdateDepartment(ctx context.Context, businessID int64, input domain.DepartmentInput) error
	DeleteDepartment(ctx context.Context, businessID, departmentID int64) error
}

type DepartmentAPI struct {
	Service     departmentService
	permissions PermissionProvider
}

func NewDepartmentAPI(service departmentService, permissions ...PermissionProvider) *DepartmentAPI {
	var p PermissionProvider
	if len(permissions) > 0 {
		p = permissions[0]
	}
	return &DepartmentAPI{Service: service, permissions: p}
}

func (a *DepartmentAPI) ListDepartments(ctx context.Context, _ *hrdepartmentv1.ListDepartmentsRequest) (*hrdepartmentv1.ListDepartmentsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	departments, err := a.Service.ListDepartments(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list departments")
	}

	payloads := make([]*hrdepartmentv1.DepartmentPayload, 0, len(departments))
	for _, d := range departments {
		payloads = append(payloads, toProtoDepartment(d))
	}
	return &hrdepartmentv1.ListDepartmentsResponse{Departments: payloads}, nil
}

func (a *DepartmentAPI) CreateDepartment(ctx context.Context, req *hrdepartmentv1.CreateDepartmentRequest) (*hrdepartmentv1.CreateDepartmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := requireDepartmentManagePermission(ctx, a.permissions, businessID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	id, err := a.Service.CreateDepartment(ctx, domain.DepartmentInput{
		OwnerID:        businessID,
		BusinessID:     businessID,
		Name:           req.GetName(),
		Description:    req.GetDescription(),
		ManagerStaffID: req.GetManagerStaffId(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create department")
	}
	return &hrdepartmentv1.CreateDepartmentResponse{Id: id}, nil
}

func (a *DepartmentAPI) UpdateDepartment(ctx context.Context, req *hrdepartmentv1.UpdateDepartmentRequest) (*hrdepartmentv1.UpdateDepartmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := requireDepartmentManagePermission(ctx, a.permissions, businessID); err != nil {
		return nil, err
	}
	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	if err := a.Service.UpdateDepartment(ctx, businessID, domain.DepartmentInput{
		ID:             req.GetId(),
		Name:           body.GetName(),
		Description:    body.GetDescription(),
		ManagerStaffID: body.GetManagerStaffId(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update department")
	}
	return &hrdepartmentv1.UpdateDepartmentResponse{Success: true}, nil
}

func (a *DepartmentAPI) DeleteDepartment(ctx context.Context, req *hrdepartmentv1.DeleteDepartmentRequest) (*hrdepartmentv1.DeleteDepartmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := requireDepartmentManagePermission(ctx, a.permissions, businessID); err != nil {
		return nil, err
	}
	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "id is required")
	}

	if err := a.Service.DeleteDepartment(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete department")
	}
	return &hrdepartmentv1.DeleteDepartmentResponse{Success: true}, nil
}

func toProtoDepartment(d domain.Department) *hrdepartmentv1.DepartmentPayload {
	return &hrdepartmentv1.DepartmentPayload{
		Id:             d.ID,
		OwnerId:        d.OwnerID,
		BusinessId:     d.BusinessID,
		Name:           d.Name,
		Description:    d.Description,
		ManagerStaffId: d.ManagerStaffID,
		CreatedAt:      d.CreatedAt.Unix(),
		UpdatedAt:      d.UpdatedAt.Unix(),
	}
}
