import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.user.v1;

namespace hub.team.invitations.v1;

model InviteUserRequest {
  @Onkt.required
  @format("email")
  email: string;

  @Onkt.required
  full_name: string;

  /** Use admin for a full-access business administrator, otherwise member. */
  member_type: string;

  /** App licenses assigned directly to the invited member. */
  licensed_app_ids: string[];

  /** Direct permission profiles assigned to the invited member. */
  permission_profile_assignments: PermissionProfileAssignment[];

  /** Groups that provide inherited access. */
  @Onkt.encode("string")
  group_ids: int64[];

  /** Optional user-provided secret; omit it to send an invitation link. */
  password?: string;

  /** Direct credential provisioning uses a write-only temporary secret. */
  username?: string;

  /** Write-only temporary password for direct provisioning. */
  temporary_password?: string;

  /** Invitation or direct-credential provisioning mode. */
  provisioning_mode?: string;
}

model InviteUserResponse {
  @encode(string)
  user_id: int64;

  message: string;
}

model BulkInviteRequest {
  @minItems(1)
  emails: string[];

  /** Use admin for full-access business administrators, otherwise member. */
  member_type: string;

  /** App licenses shared by all invitees. */
  licensed_app_ids: string[];

  /** Permission profiles shared by all invitees. */
  permission_profile_assignments: PermissionProfileAssignment[];

  /** Groups that provide inherited access to all invitees. */
  @Onkt.encode("string")
  group_ids: int64[];
}

model BulkInviteResult {
  email: string;

  @encode(string)
  user_id: int64;

  status: string;
  message: string;
}

model BulkInviteResponse {
  /** Per-email invite outcome list. */
  results: BulkInviteResult[];

  message: string;
}

model VerifyInviteRequest {
  /** Opaque token used to complete the operation. */
  @Onkt.required
  token: string;
}

model VerifyInviteResponse {
  email: string;
  full_name: string;
  valid: boolean;

  /** App IDs for which the invitation reserves a license. */
  licensed_app_ids: string[];

  /** Permission profiles that will become active after acceptance. */
  permission_profile_assignments: PermissionProfileAssignment[];
}

model AcceptInviteRequest {
  /** Opaque token used to complete the operation. */
  @Onkt.required
  token: string;

  /** User-provided secret for authentication. */
  @Onkt.required
  password: string;
}

model AcceptInviteResponse {
  message: string;
}
