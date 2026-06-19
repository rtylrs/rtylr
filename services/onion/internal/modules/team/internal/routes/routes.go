package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	serverHttp.RegisterGET[handlers.ListUsersRequest, handlers.ListUsersResponse](
		ctx, router, handlers.ListUsersRoutes(),
		&handlers.ListUsersHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.GetMemberByIDRequest, handlers.GetMemberByIDResponse](
		ctx, router, handlers.GetMemberByIDRoutes(),
		&handlers.GetMemberByIDHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.InviteUserRequest, handlers.InviteUserResponse](
		ctx, router, handlers.InviteUserRoutes(),
		&handlers.InviteUserHandler{Service: service},
	)

	serverHttp.RegisterPUTWithValidation[handlers.UpdatePermissionsRequest, handlers.UpdatePermissionsResponse](
		ctx, router, handlers.UpdatePermissionsRoutes(),
		&handlers.UpdatePermissionsHandler{Service: service},
	)

	serverHttp.RegisterDELETEWithValidation[handlers.DeactivateUserRequest, handlers.DeactivateUserResponse](
		ctx, router, handlers.ListUsersRoutes(),
		&handlers.DeactivateUserHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.MeRequest, handlers.MeResponse](
		ctx, router, handlers.MeRoutes(),
		&handlers.MeHandler{Service: service},
	)

	// Invite verification and acceptance (anonymous - no auth required)
	serverHttp.RegisterGET[handlers.VerifyInviteRequest, handlers.VerifyInviteResponse](
		ctx, router, handlers.VerifyInviteRoutes(),
		&handlers.VerifyInviteHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.AcceptInviteRequest, handlers.AcceptInviteResponse](
		ctx, router, handlers.AcceptInviteRoutes(),
		&handlers.AcceptInviteHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.MarkOnboardedRequest, handlers.MarkOnboardedResponse](
		ctx, router, handlers.MarkOnboardedRoutes(),
		&handlers.MarkOnboardedHandler{Service: service},
	)

	serverHttp.RegisterPOSTWithValidation[handlers.RequestAccessRequest, handlers.RequestAccessResponse](
		ctx, router, handlers.RequestAccessRoutes(),
		&handlers.RequestAccessHandler{Service: service},
	)

	// PATCH /team/members/{id}/deactivate — convenience path-param deactivation for per-app UIs
	serverHttp.RegisterPATCHWithValidation[handlers.DeactivateMemberRequest, handlers.DeactivateMemberResponse](
		ctx, router, handlers.DeactivateMemberRoutes(),
		&handlers.DeactivateMemberHandler{Service: service},
	)

	// Groups
	serverHttp.RegisterGET[handlers.ListGroupsRequest, handlers.ListGroupsResponse](
		ctx, router, handlers.ListGroupsRoutes(),
		&handlers.ListGroupsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateGroupRequest, handlers.CreateGroupResponse](
		ctx, router, handlers.ListGroupsRoutes(),
		&handlers.CreateGroupHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetGroupDetailRequest, handlers.GetGroupDetailResponse](
		ctx, router, handlers.GetGroupDetailRoutes(),
		&handlers.GetGroupDetailHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateGroupRequest, handlers.UpdateGroupResponse](
		ctx, router, handlers.GetGroupDetailRoutes(),
		&handlers.UpdateGroupHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteGroupRequest, handlers.DeleteGroupResponse](
		ctx, router, handlers.GetGroupDetailRoutes(),
		&handlers.DeleteGroupHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.AddGroupMembersRequest, handlers.AddGroupMembersResponse](
		ctx, router, handlers.AddGroupMembersRoutes(),
		&handlers.AddGroupMembersHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.RemoveGroupMemberRequest, handlers.RemoveGroupMemberResponse](
		ctx, router, handlers.RemoveGroupMemberRoutes(),
		&handlers.RemoveGroupMemberHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetUserGroupsRequest, handlers.GetUserGroupsResponse](
		ctx, router, handlers.GetUserGroupsRoutes(),
		&handlers.GetUserGroupsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.BulkInviteRequest, handlers.BulkInviteResponse](
		ctx, router, handlers.BulkInviteRoutes(),
		&handlers.BulkInviteHandler{Service: service},
	)

	serverHttp.RegisterGET[handlers.ListAccessRequestsRequest, handlers.ListAccessRequestsResponse](
		ctx, router, handlers.ListAccessRequestsRoutes(),
		&handlers.ListAccessRequestsHandler{Service: service},
	)
}
