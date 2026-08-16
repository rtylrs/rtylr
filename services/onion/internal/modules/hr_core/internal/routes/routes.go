package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Setup
	serverHttp.RegisterGET[handlers.GetSetupStatusRequest, handlers.GetSetupStatusResponse](
		ctx, router, handlers.GetSetupStatusRoutes(), &handlers.GetSetupStatusHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CompleteSetupStepRequest, handlers.CompleteSetupStepResponse](
		ctx, router, handlers.CompleteSetupStepRoutes(), &handlers.CompleteSetupStepHandler{Service: service},
	)

	// Config
	serverHttp.RegisterGET[handlers.GetHRConfigRequest, handlers.GetHRConfigResponse](
		ctx, router, handlers.GetHRConfigRoutes(), &handlers.GetHRConfigHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpsertHRConfigRequest, handlers.UpsertHRConfigResponse](
		ctx, router, handlers.UpsertHRConfigRoutes(), &handlers.UpsertHRConfigHandler{Service: service},
	)

	// Employees
	serverHttp.RegisterGET[handlers.ListEmployeesRequest, handlers.ListEmployeesResponse](
		ctx, router, handlers.ListEmployeesRoutes(), &handlers.ListEmployeesHandler{Service: service},
	)
	// Registered before GetEmployeeRoutes ({id}) so chi resolves the fixed
	// "me" segment first rather than parsing it as an employee id.
	serverHttp.RegisterGET[handlers.GetMyEmployeeRequest, handlers.GetMyEmployeeResponse](
		ctx, router, handlers.GetMyEmployeeRoutes(), &handlers.GetMyEmployeeHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetEmployeeRequest, handlers.GetEmployeeResponse](
		ctx, router, handlers.GetEmployeeRoutes(), &handlers.GetEmployeeHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateEmployeeRequest, handlers.CreateEmployeeResponse](
		ctx, router, handlers.CreateEmployeeRoutes(), &handlers.CreateEmployeeHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateEmployeeRequest, handlers.UpdateEmployeeResponse](
		ctx, router, handlers.UpdateEmployeeRoutes(), &handlers.UpdateEmployeeHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteEmployeeRequest, handlers.DeleteEmployeeResponse](
		ctx, router, handlers.DeleteEmployeeRoutes(), &handlers.DeleteEmployeeHandler{Service: service},
	)

	// Invites
	serverHttp.RegisterPOSTWithValidation[handlers.InviteEmployeeRequest, handlers.InviteEmployeeResponse](
		ctx, router, handlers.InviteEmployeeRoutes(), &handlers.InviteEmployeeHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.VerifyInviteRequest, handlers.VerifyInviteResponse](
		ctx, router, handlers.VerifyInviteRoutes(), &handlers.VerifyInviteHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.AcceptInviteRequest, handlers.AcceptInviteResponse](
		ctx, router, handlers.AcceptInviteRoutes(), &handlers.AcceptInviteHandler{Service: service},
	)

	// Locations
	serverHttp.RegisterGET[handlers.ListLocationsRequest, handlers.ListLocationsResponse](
		ctx, router, handlers.ListLocationsRoutes(), &handlers.ListLocationsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateLocationRequest, handlers.CreateLocationResponse](
		ctx, router, handlers.CreateLocationRoutes(), &handlers.CreateLocationHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateLocationRequest, handlers.UpdateLocationResponse](
		ctx, router, handlers.UpdateLocationRoutes(), &handlers.UpdateLocationHandler{Service: service},
	)

	// Departments
	serverHttp.RegisterGET[handlers.ListHRCoreDepartmentsRequest, handlers.ListHRCoreDepartmentsResponse](
		ctx, router, handlers.ListHRCoreDepartmentsRoutes(), &handlers.ListHRCoreDepartmentsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateHRCoreDepartmentRequest, handlers.CreateHRCoreDepartmentResponse](
		ctx, router, handlers.CreateHRCoreDepartmentRoutes(), &handlers.CreateHRCoreDepartmentHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateHRCoreDepartmentRequest, handlers.UpdateHRCoreDepartmentResponse](
		ctx, router, handlers.UpdateHRCoreDepartmentRoutes(), &handlers.UpdateHRCoreDepartmentHandler{Service: service},
	)

	// Work Schedules
	serverHttp.RegisterGET[handlers.ListWorkSchedulesRequest, handlers.ListWorkSchedulesResponse](
		ctx, router, handlers.ListWorkSchedulesRoutes(), &handlers.ListWorkSchedulesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateWorkScheduleRequest, handlers.CreateWorkScheduleResponse](
		ctx, router, handlers.CreateWorkScheduleRoutes(), &handlers.CreateWorkScheduleHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateWorkScheduleRequest, handlers.UpdateWorkScheduleResponse](
		ctx, router, handlers.UpdateWorkScheduleRoutes(), &handlers.UpdateWorkScheduleHandler{Service: service},
	)

	// Roles
	serverHttp.RegisterGET[handlers.ListRolesRequest, handlers.ListRolesResponse](
		ctx, router, handlers.ListRolesRoutes(), &handlers.ListRolesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateRoleRequest, handlers.CreateRoleResponse](
		ctx, router, handlers.CreateRoleRoutes(), &handlers.CreateRoleHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateRoleRequest, handlers.UpdateRoleResponse](
		ctx, router, handlers.UpdateRoleRoutes(), &handlers.UpdateRoleHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.AssignRoleRequest, handlers.AssignRoleResponse](
		ctx, router, handlers.AssignRoleRoutes(), &handlers.AssignRoleHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListPermissionCatalogRequest, handlers.ListPermissionCatalogResponse](
		ctx, router, handlers.ListPermissionCatalogRoutes(), &handlers.ListPermissionCatalogHandler{Service: service},
	)
}
