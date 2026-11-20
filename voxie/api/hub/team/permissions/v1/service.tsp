import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.permissions.v1;

/** PermissionsService defines hub team permission endpoints. */
@route("/hub/team/permissions/v1")
interface PermissionsService {
  /** Assigns apps and role for a team member. */
  @put
  @route("/permissions/{user_id}")
  updatePermissions(...UpdatePermissionsRequest): UpdatePermissionsResponse;
}
