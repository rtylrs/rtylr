import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.groups.v1;

/** GroupsService defines hub team group endpoints. */
@route("/hub/team/groups/v1")
interface GroupsService {
  /** Returns groups for the active business. */
  @get
  @route("/groups")
  listGroups(...ListGroupsRequest): ListGroupsResponse;

  /** Creates a new group. */
  @post
  @route("/groups")
  createGroup(...CreateGroupRequest): CreateGroupResponse;

  /** Returns a group with its members. */
  @get
  @route("/groups/{id}")
  getGroupDetail(...GetGroupDetailRequest): GetGroupDetailResponse;

  /** Updates an existing group. */
  @put
  @route("/groups/{id}")
  updateGroup(...UpdateGroupRequest): UpdateGroupResponse;

  /** Deletes a group. */
  @delete
  @route("/groups/{id}")
  deleteGroup(...DeleteGroupRequest): DeleteGroupResponse;

  /** Adds users to a group. */
  @post
  @route("/groups/{id}/members")
  addGroupMembers(...AddGroupMembersRequest): AddGroupMembersResponse;

  /** Removes a user from a group. */
  @delete
  @route("/groups/{id}/members/{user_id}")
  removeGroupMember(...RemoveGroupMemberRequest): RemoveGroupMemberResponse;

  /** Returns the groups a user belongs to. */
  @get
  @route("/users/{user_id}/groups")
  getUserGroups(...GetUserGroupsRequest): GetUserGroupsResponse;
}
