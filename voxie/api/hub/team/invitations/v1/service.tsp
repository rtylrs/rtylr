import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.invitations.v1;

/** InvitationsService defines hub team invitation endpoints. */
@route("/hub/team/invitations/v1")
interface InvitationsService {
  /** Sends a team invitation. */
  @post
  @route("/invitations")
  inviteUser(...InviteUserRequest): InviteUserResponse;

  /** Sends team invitations to multiple emails at once. */
  @post
  @route("/invitations/bulk")
  bulkInvite(...BulkInviteRequest): BulkInviteResponse;

  /** Validates an invitation token. */
  @post
  @route("/invitations/verify")
  verifyInvite(...VerifyInviteRequest): VerifyInviteResponse;

  /** Accepts a team invitation. */
  @post
  @route("/invitations/accept")
  acceptInvite(...AcceptInviteRequest): AcceptInviteResponse;
}
