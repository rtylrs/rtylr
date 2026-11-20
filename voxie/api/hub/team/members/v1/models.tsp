import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.team.groups.v1;
using hub.user.v1;

namespace hub.team.members.v1;

model StaffUserPayload {
  @encode(string)
  id: int64;

  /** Email address used for lookup or contact. */
  email: string;

  full_name: string;

  /** When the record was created. */
  created_at: string;

  is_active: boolean;

  /** Administrative members inherit all business permissions. */
  member_type: string;

  /** Direct app license assignments. */
  licensed_app_ids: string[];

  /** Direct application permission profile assignments. */
  permission_profile_assignments: PermissionProfileAssignment[];

  /** Server-resolved effective access. */
  effective_access?: EffectiveAccess;
}

model StaffUserWithGroups {
  @encode(string)
  id: int64;

  /** Email address used for lookup or contact. */
  email: string;

  full_name: string;

  /** When the record was created. */
  created_at: string;

  is_active: boolean;

  /** Administrative members inherit all business permissions. */
  member_type: string;

  /** Direct app license assignments. */
  licensed_app_ids: string[];

  /** Direct application permission profile assignments. */
  permission_profile_assignments: PermissionProfileAssignment[];

  /** Server-resolved effective access. */
  effective_access?: EffectiveAccess;

  /** Groups the staff user belongs to. */
  groups: GroupSummary[];
}

model ListTeamRequest {
  /** page is the 1-based result page to return; defaults to 1. */
  @query("page")
  page: int32;

  /** page_size is the maximum number of members per page; server caps at 500. */
  @query("page_size")
  page_size: int32;
}

model ListTeamResponse {
  users: StaffUserPayload[];

  /** total is the total number of members across all pages, for pagination math. */
  total: int32;
}

model GetMemberRequest {
  /** Identifies the member to fetch. */
  @minValueExclusive(0)
  @encode(string)
  user_id: int64;
}

model GetMemberResponse {
  user: StaffUserWithGroups;
}

model RemoveMemberRequest {
  @minValueExclusive(0)
  @encode(string)
  user_id: int64;
}

model RemoveMemberResponse {
  message: string;
}
