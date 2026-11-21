import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.role.v1;

/** RoleService defines HR role endpoints. */
@route("/hr/role/v1")
interface RoleService {
  /** Creates a new role. */
  @post
  @route("/roles")
  createRole(...CreateRoleRequest): CreateRoleResponse;

  /** Updates an existing role. */
  @put
  @route("/roles/{id}")
  updateRole(...UpdateRoleRequest): UpdateRoleResponse;

  /** Assigns a role to an employee. */
  @post
  @route("/roles/assign")
  assignRole(...AssignRoleRequest): AssignRoleResponse;
}
