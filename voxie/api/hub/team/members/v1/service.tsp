import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.members.v1;

/** MembersService defines hub team member endpoints. */
@route("/hub/team/members/v1")
interface MembersService {
  /** Returns team members for the active business. */
  @get
  @route("/members")
  listTeam(...ListTeamRequest): ListTeamResponse;

  /** Returns a single team member by ID. */
  @get
  @route("/members/{user_id}")
  getMember(...GetMemberRequest): GetMemberResponse;

  /** Removes a member from the team. */
  @delete
  @route("/members/{user_id}")
  removeMember(...RemoveMemberRequest): RemoveMemberResponse;
}
