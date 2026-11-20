import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.access.v1;

/** MyLocation is a location entry for the current staff member. */
model MyLocation {
  @encode(string)
  location_id: int64;

  name: string;
  address: string;
  phone: string;
}

/** RosterEntry is a staff member entry within a location roster. */
model RosterEntry {
  /** user_id identifies the staff member. */
  @encode(string)
  user_id: int64;

  name: string;

  /** has_pin tells the terminal whether the staff member can sign in with a PIN. */
  has_pin: boolean;

  /** is_manager tells whether this staff member can manage POS access. */
  is_manager: boolean;

  /** pos_role is the POS-specific role label. */
  pos_role: string;
}

/** MemberLocationEntry is a location assignment for a staff member. */
model MemberLocationEntry {
  /** id is the unique assignment identifier. */
  @encode(string)
  id: int64;

  /** location_id identifies the assigned location. */
  @encode(string)
  location_id: int64;

  /** can_manage_pos_users controls POS staff management. */
  can_manage_pos_users: boolean;

  /** can_manage_terminals controls terminal management. */
  can_manage_terminals: boolean;

  /** can_view_reports controls POS reporting access. */
  can_view_reports: boolean;
}

model AssignLocationRequest {
  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  location_id?: int64;

  can_manage_pos_users: boolean;
  can_manage_terminals: boolean;
  can_view_reports: boolean;
}

model AssignLocationResponse {
  @encode(string)
  id: int64;
}

model GetLocationRosterRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("location_id")
  @encode(string)
  location_id?: int64;
}

model GetLocationRosterResponse {
  roster: RosterEntry[];
}

model GetMemberLocationsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("user_id")
  @encode(string)
  user_id?: int64;
}

model GetMemberLocationsResponse {
  locations: MemberLocationEntry[];
}

model GetMyLocationsRequest {}

model GetMyLocationsResponse {
  locations: MyLocation[];
}

model RemoveLocationRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;
}

model RemoveLocationResponse {
  success: boolean;
}
