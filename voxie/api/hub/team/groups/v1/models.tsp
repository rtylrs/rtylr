import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.user.v1;

namespace hub.team.groups.v1;

model GroupSummary {
  @encode(string)
  id: int64;

  name: string;
  color: string;
}

model Group {
  @encode(string)
  id: int64;

  /** Identifies the business the group belongs to. */
  @encode(string)
  business_id: int64;

  name: string;
  description: string;
  color: string;
  licensed_app_ids: string[];
  permission_profile_assignments: PermissionProfileAssignment[];
  member_count: int32;

  /** When the record was created. */
  created_at: string;

  /** When the record was last updated. */
  updated_at: string;
}

model GroupMember {
  /** Identifies the member user. */
  @encode(string)
  user_id: int64;

  /** Email address used for lookup or contact. */
  email: string;

  full_name: string;

  /** When the member was added to the group. */
  added_at: string;
}

/**
 * GroupDetail mirrors Group with the member list flattened alongside it
 * (the legacy payload embeds the group fields at the top level).
 */
model GroupDetail {
  @encode(string)
  id: int64;

  /** Identifies the business the group belongs to. */
  @encode(string)
  business_id: int64;

  name: string;
  description: string;
  color: string;
  licensed_app_ids: string[];
  permission_profile_assignments: PermissionProfileAssignment[];
  member_count: int32;

  /** When the record was created. */
  created_at: string;

  /** When the record was last updated. */
  updated_at: string;

  members: GroupMember[];
}

model ListGroupsRequest {}

model ListGroupsResponse {
  groups: Group[];
}

model CreateGroupRequest {
  @Onkt.required
  name: string;

  description: string;
  color: string;
  licensed_app_ids: string[];
  permission_profile_assignments: PermissionProfileAssignment[];
}

model CreateGroupResponse {
  group: Group;
}

model GetGroupDetailRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetGroupDetailResponse {
  group: GroupDetail;
}

model UpdateGroupRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  name: string;

  description: string;
  color: string;
  licensed_app_ids: string[];
  permission_profile_assignments: PermissionProfileAssignment[];
}

model UpdateGroupResponse {
  group: Group;
}

model DeleteGroupRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteGroupResponse {
  message: string;
}

model AddGroupMembersRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @minItems(1)
  @Onkt.encode("string")
  user_ids: int64[];
}

model AddGroupMembersResponse {
  /** Number of members added. */
  added: int32;

  message: string;
}

model RemoveGroupMemberRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @minValueExclusive(0)
  @encode(string)
  user_id: int64;
}

model RemoveGroupMemberResponse {
  message: string;
}

model GetUserGroupsRequest {
  @minValueExclusive(0)
  @encode(string)
  user_id: int64;
}

model GetUserGroupsResponse {
  groups: GroupSummary[];
}
