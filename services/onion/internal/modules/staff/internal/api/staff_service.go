package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/staff/internal/application"
	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
)

type StaffAPI struct {
	Service     *application.Service
	permissions PermissionProvider
}

func NewStaffAPI(service *application.Service, permissions PermissionProvider) *StaffAPI {
	return &StaffAPI{Service: service, permissions: permissions}
}

func (a *StaffAPI) ListStaff(ctx context.Context, req *hrstaffv1.ListStaffRequest) (*hrstaffv1.ListStaffResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	device := authPkg.DeviceCtx(ctx)

	if !canViewDirectory(ctx, a.permissions, businessID, device.User.UserID) {
		return nil, onekiterrors.PermissionDenied()
	}

	staffList, listErr := a.Service.ListStaff(ctx, businessID, req.GetSearch(), req.GetDepartmentId(), req.GetStatus())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list staff")
	}

	hasSensitiveDetailsPermission := canViewSensitiveDetails(ctx, a.permissions, businessID, device.User.UserID)
	payloads := make([]*hrstaffv1.Staff, 0, len(staffList))
	for _, staff := range staffList {
		payloads = append(payloads, newStaffPayload(staff, hasSensitiveDetailsPermission, device.User.UserID))
	}

	return &hrstaffv1.ListStaffResponse{Staff: payloads}, nil
}

// maxDirectoryPageSize bounds how many rows a single directory page can
// request, regardless of what the caller asks for - the generated request
// only validates page_size > 0, so an unbounded value would defeat pagination.
const maxDirectoryPageSize = 100

// ListStaffDirectory serves the company-wide staff roster to any employee
// whose business has turned the directory on and whose role holds
// view_directory - both real checks against hr_core, not decorative ones.
// Unlike ListStaff, it can never expose salary, address, or emergency-contact
// data: DirectoryStaffEntry has no fields for them at the contract level.
func (a *StaffAPI) ListStaffDirectory(ctx context.Context, req *hrstaffv1.ListStaffDirectoryRequest) (*hrstaffv1.ListStaffDirectoryResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	device := authPkg.DeviceCtx(ctx)

	if !canViewDirectory(ctx, a.permissions, businessID, device.User.UserID) {
		return nil, onekiterrors.PermissionDenied()
	}

	pageSize := req.GetPageSize()
	if pageSize > maxDirectoryPageSize {
		pageSize = maxDirectoryPageSize
	}
	offset := (req.GetPage() - 1) * pageSize

	staffList, total, listErr := a.Service.ListStaffDirectory(ctx, businessID, pageSize, offset)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list staff directory")
	}

	entries := make([]*hrstaffv1.DirectoryStaffEntry, 0, len(staffList))
	for _, staff := range staffList {
		entries = append(entries, newDirectoryStaffEntry(staff))
	}

	return &hrstaffv1.ListStaffDirectoryResponse{
		Staff:    entries,
		Total:    total,
		Page:     req.GetPage(),
		PageSize: pageSize,
	}, nil
}

func (a *StaffAPI) CreateStaff(ctx context.Context, req *hrstaffv1.CreateStaffRequest) (*hrstaffv1.CreateStaffResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, createErr := a.Service.CreateStaff(ctx, domain.StaffInput{
		OwnerID:               device.User.UserID,
		BusinessID:            businessID,
		UserID:                req.GetUserId(),
		DepartmentID:          req.GetDepartmentId(),
		ManagerStaffID:        req.GetManagerStaffId(),
		FirstName:             req.GetFirstName(),
		LastName:              req.GetLastName(),
		Email:                 req.GetEmail(),
		Phone:                 req.GetPhone(),
		Role:                  req.GetRole(),
		EmploymentType:        req.GetEmploymentType(),
		Status:                req.GetStatus(),
		Salary:                req.GetSalary(),
		HourlyRate:            req.GetHourlyRate(),
		OvertimeRate:          req.GetOvertimeRate(),
		StartDate:             req.GetStartDate(),
		EndDate:               req.GetEndDate(),
		Address:               req.GetAddress(),
		EmergencyContactName:  req.GetEmergencyContactName(),
		EmergencyContactPhone: req.GetEmergencyContactPhone(),
		Apps:                  req.GetApps(),
		Notes:                 req.GetNotes(),
		PhotoURL:              req.GetPhotoUrl(),
		Bio:                   req.GetBio(),
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create staff member")
	}

	return &hrstaffv1.CreateStaffResponse{Id: id}, nil
}

func (a *StaffAPI) GetStaff(ctx context.Context, req *hrstaffv1.GetStaffRequest) (*hrstaffv1.GetStaffResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	device := authPkg.DeviceCtx(ctx)

	staff, getErr := a.Service.GetStaffByID(ctx, businessID, req.GetId())
	if getErr != nil {
		if errors.Is(getErr, domain.ErrStaffNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Staff member not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get staff member")
	}

	hasSensitiveDetailsPermission := canViewSensitiveDetails(ctx, a.permissions, businessID, device.User.UserID)
	return &hrstaffv1.GetStaffResponse{
		Staff: newStaffPayload(*staff, hasSensitiveDetailsPermission, device.User.UserID),
	}, nil
}

func (a *StaffAPI) UpdateStaff(ctx context.Context, req *hrstaffv1.UpdateStaffRequest) (*hrstaffv1.UpdateStaffResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "body is required")
	}

	updateErr := a.Service.UpdateStaff(ctx, businessID, domain.StaffUpdateInput{
		ID:                    req.GetId(),
		DepartmentID:          body.GetDepartmentId(),
		ManagerStaffID:        body.GetManagerStaffId(),
		FirstName:             body.GetFirstName(),
		LastName:              body.GetLastName(),
		Email:                 body.GetEmail(),
		Phone:                 body.GetPhone(),
		Role:                  body.GetRole(),
		EmploymentType:        body.GetEmploymentType(),
		Status:                body.GetStatus(),
		Salary:                body.GetSalary(),
		HourlyRate:            body.GetHourlyRate(),
		OvertimeRate:          body.GetOvertimeRate(),
		EndDate:               body.GetEndDate(),
		Address:               body.GetAddress(),
		EmergencyContactName:  body.GetEmergencyContactName(),
		EmergencyContactPhone: body.GetEmergencyContactPhone(),
		Apps:                  body.GetApps(),
		Notes:                 body.GetNotes(),
		PhotoURL:              body.GetPhotoUrl(),
		Bio:                   body.GetBio(),
	})
	if updateErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update staff member")
	}

	return &hrstaffv1.UpdateStaffResponse{Success: true}, nil
}

func (a *StaffAPI) DeleteStaff(ctx context.Context, req *hrstaffv1.DeleteStaffRequest) (*hrstaffv1.DeleteStaffResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if deleteErr := a.Service.DeleteStaff(ctx, businessID, req.GetId()); deleteErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to terminate staff member")
	}

	return &hrstaffv1.DeleteStaffResponse{Success: true}, nil
}

func (a *StaffAPI) ClockIn(ctx context.Context, req *hrstaffv1.ClockInRequest) (*hrstaffv1.ClockInResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if clockErr := a.Service.ClockIn(ctx, businessID, req.GetStaffId()); clockErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to clock in")
	}

	return &hrstaffv1.ClockInResponse{Success: true}, nil
}

func (a *StaffAPI) ClockOut(ctx context.Context, req *hrstaffv1.ClockOutRequest) (*hrstaffv1.ClockOutResponse, error) {
	businessID, err := resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if clockErr := a.Service.ClockOut(ctx, businessID, req.GetStaffId()); clockErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to clock out")
	}

	return &hrstaffv1.ClockOutResponse{Success: true}, nil
}

func resolveBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}

	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func newStaffPayload(staff domain.Staff, hasSensitiveDetailsPermission bool, callerUserID int64) *hrstaffv1.Staff {
	// A caller always sees their own compensation and contact detail, even
	// without manage_employees/manage_payroll - that permission gates seeing
	// someone else's, not your own record back to you.
	isOwnRecord := callerUserID > 0 && staff.UserID == callerUserID
	if !hasSensitiveDetailsPermission && !isOwnRecord {
		staff.Salary = 0
		staff.HourlyRate = 0
		staff.OvertimeRate = 0
		staff.Address = ""
		staff.EmergencyContactName = ""
		staff.EmergencyContactPhone = ""
	}

	return &hrstaffv1.Staff{
		Id:                    staff.ID,
		OwnerId:               staff.OwnerID,
		UserId:                staff.UserID,
		DepartmentId:          staff.DepartmentID,
		ManagerStaffId:        staff.ManagerStaffID,
		FirstName:             staff.FirstName,
		LastName:              staff.LastName,
		Email:                 staff.Email,
		Phone:                 staff.Phone,
		Role:                  staff.Role,
		EmploymentType:        staff.EmploymentType,
		Status:                staff.Status,
		Salary:                staff.Salary,
		HourlyRate:            staff.HourlyRate,
		OvertimeRate:          staff.OvertimeRate,
		StartDate:             formatStaffTime(staff.StartDate),
		EndDate:               formatStaffTime(staff.EndDate),
		Address:               staff.Address,
		EmergencyContactName:  staff.EmergencyContactName,
		EmergencyContactPhone: staff.EmergencyContactPhone,
		Apps:                  staff.Apps,
		Notes:                 staff.Notes,
		PhotoUrl:              staff.PhotoURL,
		Bio:                   staff.Bio,
		CreatedAt:             formatStaffTime(staff.CreatedAt),
		UpdatedAt:             formatStaffTime(staff.UpdatedAt),
	}
}

// newDirectoryStaffEntry projects a Staff record onto the minimal directory
// shape. There is no permission fork here the way newStaffPayload has one for
// compensation/contact fields - DirectoryStaffEntry simply has no field to
// carry salary, address, or emergency-contact data, so there is nothing to
// zero out.
func newDirectoryStaffEntry(staff domain.Staff) *hrstaffv1.DirectoryStaffEntry {
	return &hrstaffv1.DirectoryStaffEntry{
		Id:             staff.ID,
		FirstName:      staff.FirstName,
		LastName:       staff.LastName,
		DepartmentId:   staff.DepartmentID,
		Role:           staff.Role,
		EmploymentType: staff.EmploymentType,
		Status:         staff.Status,
		Email:          staff.Email,
		Phone:          staff.Phone,
	}
}

func formatStaffTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02T15:04:05Z")
}
