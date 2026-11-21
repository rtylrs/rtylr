import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.workspace.v1;

@route("/flow/workspace/v1")
interface WorkspaceService {
  @get
  @route("/workspaces")
  listWorkspaces(...ListWorkspacesRequest): ListWorkspacesResponse;

  @post
  @route("/workspaces")
  createWorkspace(...CreateWorkspaceRequest): CreateWorkspaceResponse;

  @put
  @route("/workspaces/{id}")
  updateWorkspace(...UpdateWorkspaceRequest): UpdateWorkspaceResponse;

  @delete
  @route("/workspaces/{id}")
  deleteWorkspace(...DeleteWorkspaceRequest): DeleteWorkspaceResponse;

  @get
  @route("/workspaces/members")
  listWorkspaceMembers(
    ...ListWorkspaceMembersRequest,
  ): ListWorkspaceMembersResponse;

  @post
  @route("/workspaces/members")
  addWorkspaceMember(...AddWorkspaceMemberRequest): AddWorkspaceMemberResponse;

  @delete
  @route("/workspaces/members")
  removeWorkspaceMember(
    ...RemoveWorkspaceMemberRequest,
  ): RemoveWorkspaceMemberResponse;

  @put
  @route("/workspaces/members/role")
  updateWorkspaceMemberRole(
    ...UpdateWorkspaceMemberRoleRequest,
  ): UpdateWorkspaceMemberRoleResponse;
}
