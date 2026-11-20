import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.user.v1;

namespace hub.team.permissions.v1;

model UpdatePermissionsRequest {
  @minValueExclusive(0)
  @encode(string)
  user_id: int64;

  /** Use admin for a full-access business administrator, otherwise member. */
  member_type: string;

  /** App licenses assigned directly to this member. */
  licensed_app_ids: string[];

  /** Replace direct app licenses when true. */
  replace_licensed_app_ids: boolean;

  full_name?: string;

  /** Direct permission profiles assigned to the member. */
  permission_profile_assignments: PermissionProfileAssignment[];

  /** Replace direct permission profiles when true. */
  replace_permission_profile_assignments: boolean;

  /** Groups that provide inherited access. */
  @Onkt.encode("string")
  group_ids: int64[];

  /** Replace group memberships when true. */
  replace_group_ids: boolean;
}

model UpdatePermissionsResponse {
  /** Effective access after the update. */
  effective_access?: EffectiveAccess;
}
