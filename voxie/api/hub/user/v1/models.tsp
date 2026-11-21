import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.user.v1;

model PermissionScope {
  /** Scope resource type, such as location, department, or project. */
  scope_type: string;

  /** Tenant-owned resource identifier. */
  scope_id: string;
}

model PermissionProfileAssignment {
  /** Application receiving the role assignment. */
  app_id: string;

  /** Permission profiles applied to this app. */
  permission_profile_ids: string[];

  /** Server-owned assignment source. */
  assignment_source: string;

  /** Assignment lifecycle status. */
  status: string;

  /** Resource scopes attached to the assignment. */
  scopes: PermissionScope[];

  /** Group target when the assignment is group-derived. */
  group_id?: string;
}

model LicenseAssignment {
  /** Licensed application identifier. */
  app_id: string;

  /** License lifecycle status. */
  status: string;

  /** Server-owned license source. */
  assignment_source: string;

  /** Group target when the license is group-derived. */
  group_id?: string;

  /** Time the license was granted. */
  granted_at: string;

  /** Time the license was revoked, when revoked. */
  revoked_at?: string;
}

model EffectiveAppAccess {
  /** Application represented by this access entry. */
  app_id: string;

  /** Whether an active user or group license exists. */
  licensed: boolean;

  /** Whether the user has an active role assignment. */
  has_access: boolean;

  /** Effective permission keys for this application. */
  permissions: string[];

  /** Resource scopes for each effective permission. */
  permission_scopes: PermissionScopeGrant[];

  /** Sources contributing to effective access. */
  sources: string[];
}

model PermissionScopeGrant {
  permission_key: string;
  scopes: PermissionScope[];
}

model EffectiveAccess {
  /** Active business scope for this authorization snapshot. */
  business_id: string;

  /** Effective Dash permission keys. */
  dash_permissions: string[];

  /** Business applications subscribed for this snapshot. */
  subscribed_apps: string[];

  /** Active and historical license assignments. */
  licenses: LicenseAssignment[];

  /** Effective application access entries. */
  apps: EffectiveAppAccess[];

  /** Whether the user is the business owner. */
  is_owner: boolean;

  /** Whether the membership has the full access administrator profile. */
  is_admin: boolean;

  /** Monotonic authorization version for cache refresh. */
  @encode(string)
  authorization_version: int64;
}

model ProvisioningStatus {
  /** Provisioning mode used for the account. */
  mode?: string;

  /** Current provisioning lifecycle state. */
  state?: string;

  /** Whether the work email is verified. */
  email_verified?: boolean;

  /** Whether a first-login password change is required. */
  password_change_required?: boolean;

  /** Whether MFA is required before sensitive access. */
  mfa_required?: boolean;

  /** Whether MFA enrollment is complete. */
  mfa_enrolled?: boolean;

  /** Whether the current MFA verification window is valid. */
  mfa_verified?: boolean;

  /** Whether account recovery is available. */
  recovery_available?: boolean;
}

model GetMeRequest {}

model GetMeResponse {
  @encode(string)
  id: int64;

  /** Email address used for lookup or contact. */
  email: string;

  full_name: string;

  @encode(string)
  business_id: int64;

  business_name: string;

  /** is_console_admin indicates the user can access the internal console. */
  is_console_admin: boolean;

  onboarded: boolean;

  /** Server-resolved authorization snapshot. */
  effective_access?: EffectiveAccess;

  /** Account provisioning and security state. */
  provisioning_status?: ProvisioningStatus;

  /**
   * The avatar the user chose. Absent means the default: initials on a colour
   * derived from the user id.
   */
  avatar?: UserAvatar;
}

model UpdateProfileRequest {
  @Onkt.required
  name: string;

  /** Reserved for future profile email changes. */
  @format("email")
  email?: string;
}

model UpdateProfileResponse {
  user: GetMeResponse;
}

model ChangePasswordRequest {
  @Onkt.required
  current_password: string;

  @Onkt.required
  new_password: string;
}

model ChangePasswordResponse {}

model SwitchBusinessRequest {
  /** Business to activate in the current session. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  business_id?: int64;
}

model SwitchBusinessResponse {
  /** Fresh JWT scoped to the selected business. */
  token: string;
}

/**
 * A user's chosen avatar. One per user, shown in every app and business.
 *
 * kind is one of:
 * - "photo": url is an image the user uploaded to the RTYLR CDN.
 * - "illustrated": a generated avatar, style + seed (rendered by the client).
 * - "initials": initials on palette colour `color` (0 to 11).
 * - "emoji": `emoji` on palette colour `color`.
 */
model UserAvatar {
  @Onkt.required
  @minLength(1)
  @maxLength(16)
  kind: string;

  @minLength(0)
  @maxLength(1024)
  url?: string;

  @minLength(0)
  @maxLength(32)
  style?: string;

  @minLength(0)
  @maxLength(64)
  seed?: string;

  color?: int32;

  @minLength(0)
  @maxLength(32)
  emoji?: string;
}

model UpdateAvatarRequest {
  /** The new avatar. Absent resets the user to the default. */
  avatar?: UserAvatar;
}

model UpdateAvatarResponse {
  user: GetMeResponse;
}

model ListBusinessAvatarsRequest {}

/** One colleague's chosen avatar. */
model BusinessUserAvatar {
  @encode(string)
  user_id: int64;

  avatar: UserAvatar;
}

model ListBusinessAvatarsResponse {
  /**
   * Every active member of the caller's business who has chosen an avatar.
   * Members on the default are left out; clients derive theirs from the id.
   */
  avatars: BusinessUserAvatar[];
}
