import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.permission.v1;

/** PermissionService defines hub permission and role endpoints. */
@route("/hub/permission/v1")
interface PermissionService {
  /** Returns the permission keys that can be assigned to profiles. */
  @get
  @route("/permissions")
  listPermissionCatalog(
    ...ListPermissionCatalogRequest,
  ): ListPermissionCatalogResponse;

  /** Returns defined roles. */
  @get
  @route("/roles")
  listRoles(...ListRolesRequest): ListRolesResponse;

  /** Creates a new role. */
  @post
  @route("/roles")
  createRole(...CreateRoleRequest): CreateRoleResponse;

  /** Updates a custom role. */
  @put
  @route("/roles/{id}")
  updateRole(...UpdateRoleRequest): UpdateRoleResponse;

  /** Archives a custom role. Assigned roles are revoked as part of archiving. */
  @delete
  @route("/roles/{id}")
  deleteRole(...DeleteRoleRequest): DeleteRoleResponse;

  /** Returns one role with its normalized lifecycle metadata. */
  @get
  @route("/roles/{id}")
  getRole(...GetRoleRequest): GetRoleResponse;

  /** Clones a role into a new custom role. */
  @post
  @route("/roles/{id}/clone")
  cloneRole(...CloneRoleRequest): CloneRoleResponse;

  /** Archives a custom role with an audit reason. */
  @post
  @route("/roles/{id}/archive")
  archiveRole(...ArchiveRoleRequest): ArchiveRoleResponse;

  /** Lists normalized role assignments. */
  @get
  @route("/role-assignments")
  listRoleAssignments(
    ...ListRoleAssignmentsRequest,
  ): ListRoleAssignmentsResponse;

  /** Assigns a role to one member or group. */
  @post
  @route("/role-assignments")
  assignRole(...AssignRoleRequest): AssignRoleResponse;

  /** Revokes one normalized role assignment. */
  @delete
  @route("/role-assignments/{id}")
  removeRoleAssignment(
    ...RemoveRoleAssignmentRequest,
  ): RemoveRoleAssignmentResponse;

  /** Previews the permission delta for a role update. */
  @post
  @route("/roles/{id}/preview")
  previewPermissionImpact(
    ...PreviewPermissionImpactRequest,
  ): PreviewPermissionImpactResponse;

  /** Lists role lifecycle and assignment audit events. */
  @get
  @route("/roles/audit")
  listRoleAudit(...ListRoleAuditRequest): ListRoleAuditResponse;
}
