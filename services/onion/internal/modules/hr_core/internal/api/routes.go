package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Setup
	serverHttp.RegisterGET[GetSetupStatusRequest, GetSetupStatusResponse](
		ctx, router, GetSetupStatusRoutes(), &GetSetupStatusHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CompleteSetupStepRequest, CompleteSetupStepResponse](
		ctx, router, CompleteSetupStepRoutes(), &CompleteSetupStepHandler{Service: service},
	)

	// Config
	serverHttp.RegisterGET[GetHRConfigRequest, GetHRConfigResponse](
		ctx, router, GetHRConfigRoutes(), &GetHRConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpsertHRConfigRequest, UpsertHRConfigResponse](
		ctx, router, UpsertHRConfigRoutes(), &UpsertHRConfigHandler{Service: service},
	)

	// Employees
	serverHttp.RegisterGET[ListEmployeesRequest, ListEmployeesResponse](
		ctx, router, ListEmployeesRoutes(), &ListEmployeesHandler{Service: service},
	)
	// Registered before GetEmployeeRoutes ({id}) so chi resolves the fixed
	// "me" segment first rather than parsing it as an employee id.
	serverHttp.RegisterGET[GetMyEmployeeRequest, GetMyEmployeeResponse](
		ctx, router, GetMyEmployeeRoutes(), &GetMyEmployeeHandler{Service: service},
	)
	serverHttp.RegisterGET[GetEmployeeRequest, GetEmployeeResponse](
		ctx, router, GetEmployeeRoutes(), &GetEmployeeHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateEmployeeRequest, CreateEmployeeResponse](
		ctx, router, CreateEmployeeRoutes(), &CreateEmployeeHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateEmployeeRequest, UpdateEmployeeResponse](
		ctx, router, UpdateEmployeeRoutes(), &UpdateEmployeeHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteEmployeeRequest, DeleteEmployeeResponse](
		ctx, router, DeleteEmployeeRoutes(), &DeleteEmployeeHandler{Service: service},
	)

	// Invites
	serverHttp.RegisterPOSTWithValidation[InviteEmployeeRequest, InviteEmployeeResponse](
		ctx, router, InviteEmployeeRoutes(), &InviteEmployeeHandler{Service: service},
	)
	serverHttp.RegisterGET[VerifyInviteRequest, VerifyInviteResponse](
		ctx, router, VerifyInviteRoutes(), &VerifyInviteHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[AcceptInviteRequest, AcceptInviteResponse](
		ctx, router, AcceptInviteRoutes(), &AcceptInviteHandler{Service: service},
	)

	// Locations
	serverHttp.RegisterGET[ListLocationsRequest, ListLocationsResponse](
		ctx, router, ListLocationsRoutes(), &ListLocationsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateLocationRequest, CreateLocationResponse](
		ctx, router, CreateLocationRoutes(), &CreateLocationHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateLocationRequest, UpdateLocationResponse](
		ctx, router, UpdateLocationRoutes(), &UpdateLocationHandler{Service: service},
	)

	// Departments
	serverHttp.RegisterGET[ListHRCoreDepartmentsRequest, ListHRCoreDepartmentsResponse](
		ctx, router, ListHRCoreDepartmentsRoutes(), &ListHRCoreDepartmentsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateHRCoreDepartmentRequest, CreateHRCoreDepartmentResponse](
		ctx, router, CreateHRCoreDepartmentRoutes(), &CreateHRCoreDepartmentHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateHRCoreDepartmentRequest, UpdateHRCoreDepartmentResponse](
		ctx, router, UpdateHRCoreDepartmentRoutes(), &UpdateHRCoreDepartmentHandler{Service: service},
	)

	// Work Schedules
	// Work schedules moved to the contracted oneKit path
	// (hr.work_schedule.v1, registered in module.go). The three hand-registered
	// routes that used to live here had no .onk definition and no caller in any
	// app.

	// Roles
	serverHttp.RegisterGET[ListRolesRequest, ListRolesResponse](
		ctx, router, ListRolesRoutes(), &ListRolesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateRoleRequest, CreateRoleResponse](
		ctx, router, CreateRoleRoutes(), &CreateRoleHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateRoleRequest, UpdateRoleResponse](
		ctx, router, UpdateRoleRoutes(), &UpdateRoleHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[AssignRoleRequest, AssignRoleResponse](
		ctx, router, AssignRoleRoutes(), &AssignRoleHandler{Service: service},
	)
	serverHttp.RegisterGET[ListPermissionCatalogRequest, ListPermissionCatalogResponse](
		ctx, router, ListPermissionCatalogRoutes(), &ListPermissionCatalogHandler{Service: service},
	)
}
