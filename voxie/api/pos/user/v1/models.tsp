import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.user.v1;

/** POSUser is a POS roster operator. */
model POSUser {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  location_id: int64;

  name: string;

  /** role is the role value for this contract. */
  role: string;

  is_active: boolean;
  max_discount_pct: float64;
  can_void: boolean;
  can_refund: boolean;
  can_open_drawer: boolean;
  can_apply_discount: boolean;

  /** role_management identifies the authoritative Dash control-plane surface. */
  role_management?: RoleManagementManagedInDash;
}

/**
 * RoleManagementManagedInDash explains that POS role and capability changes
 * must be completed in Dash.
 */
model RoleManagementManagedInDash {
  @Onkt.required
  code: string;

  @Onkt.required
  dash_url: string;
}

model ListPOSUsersRequest {
  @query("location_id")
  @encode(string)
  location_id: int64;
}

model ListPOSUsersResponse {
  users: POSUser[];
}

model CreatePOSUserRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  location_id?: int64;

  @Onkt.required
  name: string;

  /** pin is the pin value for this contract. */
  @Onkt.required
  pin: string;

  @minValue(0)
  hourly_rate: float64;
}

model CreatePOSUserResponse {
  @encode(string)
  id: int64;

  @Onkt.required
  role_management: RoleManagementManagedInDash;
}

model UpdatePOSUserRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;

  @Onkt.required
  name: string;

  /** Optional replacement PIN. Omit it to preserve the current PIN. */
  pin: string;

  is_active: boolean;

  @minValue(0)
  hourly_rate: float64;
}

model UpdatePOSUserResponse {
  success: boolean;

  @Onkt.required
  role_management: RoleManagementManagedInDash;
}

model DeactivatePOSUserRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;
}

model DeactivatePOSUserResponse {
  success: boolean;
}

model VerifyPOSUserPINRequest {
  /** location_id is the POS location where the operator is signing in. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  location_id?: int64;

  /** user_id is the hub user id of the roster operator. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  /** pin is the operator PIN. Sensitive: do not log this field. */
  @Onkt.required
  pin: string;
}

model VerifyPOSUserPINResponse {
  /** user_id is the hub user id for the verified operator. */
  @encode(string)
  user_id: int64;

  /** pos_user_id is the POS roster row id. */
  @encode(string)
  pos_user_id: int64;

  /** location_id is the location where the operator signed in. */
  @encode(string)
  location_id: int64;

  /** role is the operator role at this location. */
  role: string;

  /** display_name is the operator name shown on receipts and session UI. */
  display_name: string;

  /** can_void grants void permission for this session. */
  can_void: boolean;

  /** can_refund grants refund permission for this session. */
  can_refund: boolean;

  /** can_open_drawer grants cash drawer access for this session. */
  can_open_drawer: boolean;

  /** can_apply_discount grants discount permission for this session. */
  can_apply_discount: boolean;

  max_discount_pct: float64;

  /** authorization_version changes whenever Dash authorization changes. */
  @encode(string)
  authorization_version: int64;

  @Onkt.required
  role_management: RoleManagementManagedInDash;
}

model SetPINRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  location_id?: int64;

  @Onkt.required
  name: string;

  /** pin is the pin value for this contract. Sensitive: do not log this field. */
  @Onkt.required
  pin: string;
}

model SetPINResponse {
  success: boolean;
}
