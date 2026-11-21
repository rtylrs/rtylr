import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.access.v1;

/** AccessService defines hub team access request endpoints. */
@route("/hub/team/access/v1")
interface AccessService {
  /** Submits a request for app access. */
  @post
  @route("/requests")
  requestAccess(...RequestAccessRequest): RequestAccessResponse;

  /** Returns pending access requests for the business. */
  @get
  @route("/requests")
  listAccessRequests(...ListAccessRequestsRequest): ListAccessRequestsResponse;

  /** Grants the requested app access. */
  @post
  @route("/requests/{id}/approve")
  approveAccessRequest(
    ...ApproveAccessRequestRequest,
  ): ApproveAccessRequestResponse;
}
