import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.project.v2;

/**
 * ProjectService is business-scoped: every operation resolves the caller's
 * business from auth. Delete is a recycle-bin archive; permanent deletion
 * happens by retention policy or through purgeProject.
 */
@route("/flow/project/v2")
interface ProjectService {
  @get
  @route("/projects/{id}")
  getProject(...GetProjectRequest): GetProjectResponse;

  @get
  @route("/projects")
  listProjects(...ListProjectsRequest): ListProjectsResponse;

  @post
  @route("/projects")
  createProject(...CreateProjectRequest): CreateProjectResponse;

  @put
  @route("/projects/{id}")
  updateProject(...UpdateProjectRequest): UpdateProjectResponse;

  /**
   * archiveProject moves the project to the recycle bin. It is reversible via
   * restoreProject until purge_at passes.
   */
  @delete
  @route("/projects/{id}")
  archiveProject(...ArchiveProjectRequest): ArchiveProjectResponse;

  @put
  @route("/projects/visibility")
  setProjectVisibility(
    ...SetProjectVisibilityRequest,
  ): SetProjectVisibilityResponse;

  /**
   * setProjectDeveloperMode turns developer features on or off for a project.
   * Requires project admin; the My Tasks bucket cannot be switched.
   */
  @put
  @route("/projects/developer-mode")
  setProjectDeveloperMode(
    ...SetProjectDeveloperModeRequest,
  ): SetProjectDeveloperModeResponse;

  /**
   * setProjectPointsScale picks the scale the project's story points use.
   * Requires project admin. Points already set are kept.
   */
  @put
  @route("/projects/points-scale")
  setProjectPointsScale(...SetProjectPointsScaleRequest): SetProjectPointsScaleResponse;

  @get
  @route("/projects/archived")
  listArchivedProjects(
    ...ListArchivedProjectsRequest,
  ): ListArchivedProjectsResponse;

  @put
  @route("/projects/archived/restore")
  restoreProject(...RestoreProjectRequest): RestoreProjectResponse;

  /**
   * purgeProject immediately and irreversibly deletes an archived project and
   * everything inside it, including attachment files. Server-gated to the
   * project creator and business admins.
   */
  @delete
  @route("/projects/purged/{id}")
  purgeProject(...PurgeProjectRequest): PurgeProjectResponse;

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

  /**
   * getProjectLabourCost returns the labour cost of billable logged time on a
   * project. Requires admin on the project: it exposes aggregate spend, and
   * the aggregate is derived from HR rates that must not leak through Flow.
   */
  @get
  @route("/projects/{id}/labour-cost")
  getProjectLabourCost(
    ...GetProjectLabourCostRequest,
  ): GetProjectLabourCostResponse;
}
