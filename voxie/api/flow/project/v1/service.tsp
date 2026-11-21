import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.project.v1;

@route("/flow/project/v1")
interface ProjectService {
  @get
  @route("/projects")
  listProjects(...ListProjectsRequest): ListProjectsResponse;

  @post
  @route("/projects")
  createProject(...CreateProjectRequest): CreateProjectResponse;

  @put
  @route("/projects/{id}")
  updateProject(...UpdateProjectRequest): UpdateProjectResponse;

  @delete
  @route("/projects/{id}")
  deleteProject(...DeleteProjectRequest): DeleteProjectResponse;

  @get
  @route("/projects/columns")
  listProjectColumns(...ListProjectColumnsRequest): ListProjectColumnsResponse;

  @post
  @route("/projects/columns")
  createProjectColumn(
    ...CreateProjectColumnRequest,
  ): CreateProjectColumnResponse;

  @put
  @route("/projects/columns/{id}")
  updateProjectColumn(
    ...UpdateProjectColumnRequest,
  ): UpdateProjectColumnResponse;

  @delete
  @route("/projects/columns/{id}")
  deleteProjectColumn(
    ...DeleteProjectColumnRequest,
  ): DeleteProjectColumnResponse;

  @put
  @route("/projects/columns/move")
  moveTaskToColumn(...MoveTaskToColumnRequest): MoveTaskToColumnResponse;

  @get
  @route("/projects/members")
  listProjectMembers(...ListProjectMembersRequest): ListProjectMembersResponse;

  @post
  @route("/projects/members")
  addProjectMember(...AddProjectMemberRequest): AddProjectMemberResponse;

  @delete
  @route("/projects/members")
  removeProjectMember(
    ...RemoveProjectMemberRequest,
  ): RemoveProjectMemberResponse;

  @put
  @route("/projects/members/role")
  updateProjectMemberRole(
    ...UpdateProjectMemberRoleRequest,
  ): UpdateProjectMemberRoleResponse;

  @post
  @route("/projects/members/invite")
  inviteProjectMember(
    ...InviteProjectMemberRequest,
  ): InviteProjectMemberResponse;

  @put
  @route("/projects/transfer")
  transferProjectOwnership(
    ...TransferProjectOwnershipRequest,
  ): TransferProjectOwnershipResponse;

  @put
  @route("/projects/move")
  moveProject(...MoveProjectRequest): MoveProjectResponse;

  /**
   * Returns the live labour-cost aggregate for one project. Money is
   * computed server-side by HR from pay rates that never cross this
   * boundary: only aggregate cents and coverage counts are returned, so a
   * caller who may see a project's cost does not thereby gain sight of any
   * individual's pay.
   */
  @get
  @route("/projects/{id}/labour-cost")
  getProjectLabourCost(
    ...GetProjectLabourCostRequest,
  ): GetProjectLabourCostResponse;
}
