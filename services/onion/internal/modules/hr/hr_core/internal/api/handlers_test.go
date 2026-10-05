package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
)

// --- CreateEmployeeHandler ---

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

// The onboarding/setup config owns /hr/config, so the full company config API
// has to live on its own path or the two handlers compete for the same method
// and path. This pins that separation so it cannot regress silently.
func TestCoreConfigRoutesUseDedicatedNamespace(t *testing.T) {
	const want = "/hr/company-config"
	getRoutes := GetHRConfigRoutes()
	if len(getRoutes) != 1 || getRoutes[0] != want {
		t.Fatalf("core config GET routes = %v, want [%s]", getRoutes, want)
	}
	putRoutes := UpsertHRConfigRoutes()
	if len(putRoutes) != 1 || putRoutes[0] != want {
		t.Fatalf("core config PUT routes = %v, want [%s]", putRoutes, want)
	}
}

func TestCreateEmployeeHandler_Validate_MissingFirstName(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		LastName: "Smith",
		Email:    "smith@example.com",
	})
	if se == nil {
		t.Fatal("expected validation error for missing first_name")
	}
	if oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", oneKitStatus(se))
	}
}

func TestCreateEmployeeHandler_Validate_MissingLastName(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		FirstName: "John",
		Email:     "john@example.com",
	})
	if se == nil {
		t.Fatal("expected validation error for missing last_name")
	}
}

func TestCreateEmployeeHandler_Validate_MissingEmail(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		FirstName: "John",
		LastName:  "Smith",
	})
	if se == nil {
		t.Fatal("expected validation error for missing email")
	}
}

func TestCreateEmployeeHandler_Validate_InvalidStartDate(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		FirstName: "John",
		LastName:  "Smith",
		Email:     "john@example.com",
		StartDate: "01/15/2026",
	})
	if se == nil {
		t.Fatal("expected validation error for invalid start_date format")
	}
}

func TestCreateEmployeeHandler_Validate_ValidRequest(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		FirstName: "John",
		LastName:  "Smith",
		Email:     "john@example.com",
		StartDate: "2026-01-15",
	})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

func TestCreateEmployeeHandler_Validate_EmptyStartDate_IsValid(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	})
	if se != nil {
		t.Errorf("empty start_date should pass validation, got %v", se)
	}
}

func TestCreateEmployeeHandler_Validate_WhitespaceFirstName(t *testing.T) {
	h := &CreateEmployeeHandler{}
	se := h.Validate(context.Background(), &CreateEmployeeRequest{
		FirstName: "   ",
		LastName:  "Doe",
		Email:     "jane@example.com",
	})
	if se == nil {
		t.Fatal("expected error for whitespace first_name")
	}
}

// --- UpdateEmployeeHandler ---

func TestUpdateEmployeeHandler_Validate_MissingFirstName(t *testing.T) {
	h := &UpdateEmployeeHandler{}
	se := h.Validate(context.Background(), &UpdateEmployeeRequest{LastName: "Smith"})
	if se == nil {
		t.Fatal("expected validation error for missing first_name")
	}
}

func TestUpdateEmployeeHandler_Validate_InvalidStartDate(t *testing.T) {
	h := &UpdateEmployeeHandler{}
	se := h.Validate(context.Background(), &UpdateEmployeeRequest{
		FirstName: "John",
		LastName:  "Smith",
		StartDate: "not-a-date",
	})
	if se == nil {
		t.Fatal("expected validation error for invalid start_date")
	}
}

func TestUpdateEmployeeHandler_Validate_InvalidEndDate(t *testing.T) {
	h := &UpdateEmployeeHandler{}
	se := h.Validate(context.Background(), &UpdateEmployeeRequest{
		FirstName: "John",
		LastName:  "Smith",
		EndDate:   "15/01/2026",
	})
	if se == nil {
		t.Fatal("expected validation error for invalid end_date")
	}
}

func TestUpdateEmployeeHandler_Validate_ValidRequest(t *testing.T) {
	h := &UpdateEmployeeHandler{}
	se := h.Validate(context.Background(), &UpdateEmployeeRequest{
		FirstName: "John",
		LastName:  "Smith",
		StartDate: "2020-01-01",
		EndDate:   "2026-12-31",
	})
	if se != nil {
		t.Errorf("expected nil for valid update, got %v", se)
	}
}

// --- CreateHRCoreDepartmentHandler ---

func TestCreateHRCoreDepartmentHandler_Validate_MissingName(t *testing.T) {
	h := &CreateHRCoreDepartmentHandler{}
	se := h.Validate(context.Background(), &CreateHRCoreDepartmentRequest{})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %v", se)
	}
}

func TestCreateHRCoreDepartmentHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateHRCoreDepartmentHandler{}
	se := h.Validate(context.Background(), &CreateHRCoreDepartmentRequest{Name: "  "})
	if se == nil {
		t.Fatal("expected validation error for whitespace name")
	}
}

func TestCreateHRCoreDepartmentHandler_Validate_ValidRequest(t *testing.T) {
	h := &CreateHRCoreDepartmentHandler{}
	se := h.Validate(context.Background(), &CreateHRCoreDepartmentRequest{Name: "Engineering"})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

// --- UpdateHRCoreDepartmentHandler ---

func TestUpdateHRCoreDepartmentHandler_Validate_MissingName(t *testing.T) {
	h := &UpdateHRCoreDepartmentHandler{}
	se := h.Validate(context.Background(), &UpdateHRCoreDepartmentRequest{})
	if se == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestUpdateHRCoreDepartmentHandler_Validate_ValidName(t *testing.T) {
	h := &UpdateHRCoreDepartmentHandler{}
	se := h.Validate(context.Background(), &UpdateHRCoreDepartmentRequest{Name: "Product"})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

// --- CreateLocationHandler ---

func TestCreateLocationHandler_Validate_MissingName(t *testing.T) {
	h := &CreateLocationHandler{}
	se := h.Validate(context.Background(), &CreateLocationRequest{})
	if se == nil {
		t.Fatal("expected validation error for missing location name")
	}
}

func TestCreateLocationHandler_Validate_ValidName(t *testing.T) {
	h := &CreateLocationHandler{}
	se := h.Validate(context.Background(), &CreateLocationRequest{Name: "HQ"})
	if se != nil {
		t.Errorf("expected nil for valid location, got %v", se)
	}
}

func TestCreateLocationHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateLocationHandler{}
	se := h.Validate(context.Background(), &CreateLocationRequest{Name: "  "})
	if se == nil {
		t.Fatal("expected error for whitespace location name")
	}
}

// --- UpdateLocationHandler ---

func TestUpdateLocationHandler_Validate_MissingName(t *testing.T) {
	h := &UpdateLocationHandler{}
	se := h.Validate(context.Background(), &UpdateLocationRequest{})
	if se == nil {
		t.Fatal("expected validation error for missing location name on update")
	}
}

func TestUpdateLocationHandler_Validate_ValidName(t *testing.T) {
	h := &UpdateLocationHandler{}
	se := h.Validate(context.Background(), &UpdateLocationRequest{Name: "Branch Office"})
	if se != nil {
		t.Errorf("expected nil for valid update, got %v", se)
	}
}

// --- CreateRoleHandler ---

func TestCreateRoleHandler_Validate_MissingName(t *testing.T) {
	h := &CreateRoleHandler{}
	se := h.Validate(context.Background(), &CreateRoleRequest{})
	if se == nil {
		t.Fatal("expected validation error for missing role name")
	}
}

func TestCreateRoleHandler_Validate_ValidName(t *testing.T) {
	h := &CreateRoleHandler{}
	se := h.Validate(context.Background(), &CreateRoleRequest{Name: "Manager"})
	if se != nil {
		t.Errorf("expected nil for valid role, got %v", se)
	}
}

// --- UpdateRoleHandler ---

func TestUpdateRoleHandler_Validate_MissingName(t *testing.T) {
	h := &UpdateRoleHandler{}
	se := h.Validate(context.Background(), &UpdateRoleRequest{})
	if se == nil {
		t.Fatal("expected validation error for missing role name on update")
	}
}

func TestUpdateRoleHandler_Validate_ValidName(t *testing.T) {
	h := &UpdateRoleHandler{}
	se := h.Validate(context.Background(), &UpdateRoleRequest{Name: "Senior Engineer"})
	if se != nil {
		t.Errorf("expected nil for valid update, got %v", se)
	}
}

// --- AssignRoleHandler ---

func TestAssignRoleHandler_Validate_MissingRoleID(t *testing.T) {
	h := &AssignRoleHandler{}
	se := h.Validate(context.Background(), &AssignRoleRequest{RoleID: 0})
	if se == nil {
		t.Fatal("expected validation error for missing role_id")
	}
}

func TestAssignRoleHandler_Validate_ValidRoleID(t *testing.T) {
	h := &AssignRoleHandler{}
	se := h.Validate(context.Background(), &AssignRoleRequest{RoleID: 3})
	if se != nil {
		t.Errorf("expected nil for valid role_id, got %v", se)
	}
}

// --- ListPermissionCatalogHandler ---

func TestListPermissionCatalogHandler_Serve_AnonymousDenied(t *testing.T) {
	h := &ListPermissionCatalogHandler{}
	_, err := h.Serve(context.Background(), &ListPermissionCatalogRequest{})
	if oneKitStatus(err) != http.StatusForbidden && oneKitStatus(err) != http.StatusUnauthorized {
		t.Errorf("expected a permission-denied status for anonymous device, got %d", oneKitStatus(err))
	}
}
