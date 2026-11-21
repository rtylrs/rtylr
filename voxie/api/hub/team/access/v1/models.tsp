import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.user.v1;

namespace hub.team.access.v1;

model AccessRequest {
  @encode(string)
  id: int64;

  /** Identifies the user who requested access. */
  @encode(string)
  user_id: int64;

  app: string;

  /** Requested normalized role identifiers. */
  role_ids: string[];

  /** Requested resource scopes. */
  scopes: PermissionScope[];

  /** Optional business reason for the request. */
  reason?: string;

  /** Current request lifecycle status. */
  status: string;

  /** User who approved the request, when approved. */
  @encode(string)
  approver_id?: int64;

  /** Approval timestamp, when approved. */
  approved_at?: string;

  requester_name: string;
  requester_email: string;

  /** When the record was created. */
  created_at: string;
}

model RequestAccessRequest {
  @Onkt.required
  app: string;

  /** Requested normalized role identifiers. */
  role_ids: string[];

  /** Requested resource scopes. */
  scopes: PermissionScope[];

  /** Optional business reason for the request. */
  reason?: string;
}

model RequestAccessResponse {
  message: string;
}

model ListAccessRequestsRequest {}

model ListAccessRequestsResponse {
  requests: AccessRequest[];
}

model ApproveAccessRequestRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** Identifies the user being granted access. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  @Onkt.required
  app: string;

  /** Approved normalized role identifiers. */
  role_ids: string[];

  /** Approved resource scopes. */
  scopes: PermissionScope[];
}

model ApproveAccessRequestResponse {
  message: string;
}
