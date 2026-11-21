import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.staff.v1;

/** Staff is an HR staff member record. */
model Staff {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  user_id: int64;

  @encode(string)
  department_id: int64;

  /**
   * manager_staff_id is this person's direct manager, as a STAFF id (not a
   * user id). Zero means no direct line, which is the common case: a
   * department manager is resolved separately from hr_departments.
   *
   * Flow's dept-manager visibility proxy unions this with the department
   * relationship, so setting it widens what that manager can see.
   */
  @encode(string)
  manager_staff_id: int64;

  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: string;
  employment_type: string;

  /** The current lifecycle state. */
  status: string;

  salary: float64;
  hourly_rate: float64;
  overtime_rate: float64;
  start_date: string;
  end_date: string;
  address: string;
  emergency_contact_name: string;
  emergency_contact_phone: string;
  apps: string;
  notes: string;

  /**
   * A URL to the staff member's profile photo. Optional and unset by default:
   * consumers fall back to the hub account avatar, which is what Flow's profile
   * screen already does.
   */
  photo_url: string;

  /**
   * A short self-description shown on a profile screen.
   *
   * Deliberately separate from `notes`. `notes` is HR-admin commentary about a
   * person and must never be rendered as an "About" section; a dead write path
   * in the HR app mapped bio onto notes, and that mapping was wrong. See
   * tryrtylr/rtylr#875.
   */
  bio: string;

  /** When the record was created. */
  created_at: string;

  /** When the record was last updated. */
  updated_at: string;
}

model ListStaffRequest {
  @query("search")
  search: string;

  @query("department_id")
  department_id: string;

  /** The current lifecycle state. */
  @query("status")
  status: string;
}

model ListStaffResponse {
  staff: Staff[];
}

model CreateStaffRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  department_id?: int64;

  @doc("Optional, unlike department_id: most staff have no explicit manager, and\n@required @gt(0) would force callers to invent one.")
  @encode(string)
  manager_staff_id: int64;

  @Onkt.required
  first_name: string;

  @Onkt.required
  last_name: string;

  @Onkt.required
  @format("email")
  email: string;

  phone: string;

  @Onkt.required
  role: string;

  @Onkt.required
  employment_type: string;

  /** The current lifecycle state. */
  @Onkt.required
  status: string;

  @minValue(0)
  salary: float64;

  @minValue(0)
  hourly_rate: float64;

  @minValue(0)
  overtime_rate: float64;

  @Onkt.required
  start_date: string;

  /**
   * Empty for anyone still employed, which is why this is not \@required.
   * Requiring it made every update to an active employee invalid, and the
   * rejection happened client side, so the write vanished with no error.
   */
  end_date: string;

  address: string;

  /**
   * Optional HR detail. Not every record has next of kin on file, and
   * blocking an unrelated field update over it is not the way to collect it.
   */
  emergency_contact_name: string;

  emergency_contact_phone: string;

  /**
   * App access, empty meaning none. Optional for the same reason: a caller
   * changing one field should not have to assert app access to do it.
   */
  apps: string;

  notes: string;

  /**
   * A URL to the staff member's profile photo. Optional and unset by default:
   * consumers fall back to the hub account avatar, which is what Flow's profile
   * screen already does.
   */
  photo_url: string;

  /**
   * A short self-description shown on a profile screen.
   *
   * Deliberately separate from `notes`. `notes` is HR-admin commentary about a
   * person and must never be rendered as an "About" section; a dead write path
   * in the HR app mapped bio onto notes, and that mapping was wrong. See
   * tryrtylr/rtylr#875.
   */
  bio: string;
}

model CreateStaffResponse {
  @encode(string)
  id: int64;
}

model GetStaffRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetStaffResponse {
  staff: Staff;
}

model UpdateStaffBody {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  department_id?: int64;

  /** Optional for the same reason as on create. Zero clears the reporting line. */
  @encode(string)
  manager_staff_id: int64;

  @Onkt.required
  first_name: string;

  @Onkt.required
  last_name: string;

  @Onkt.required
  @format("email")
  email: string;

  phone: string;

  @Onkt.required
  role: string;

  @Onkt.required
  employment_type: string;

  /** The current lifecycle state. */
  @Onkt.required
  status: string;

  @minValue(0)
  salary: float64;

  @minValue(0)
  hourly_rate: float64;

  @minValue(0)
  overtime_rate: float64;

  /**
   * Empty for anyone still employed, which is why this is not \@required.
   * Requiring it made every update to an active employee invalid, and the
   * rejection happened client side, so the write vanished with no error.
   */
  end_date: string;

  address: string;

  /**
   * Optional HR detail. Not every record has next of kin on file, and
   * blocking an unrelated field update over it is not the way to collect it.
   */
  emergency_contact_name: string;

  emergency_contact_phone: string;

  /**
   * App access, empty meaning none. Optional for the same reason: a caller
   * changing one field should not have to assert app access to do it.
   */
  apps: string;

  notes: string;

  /**
   * A URL to the staff member's profile photo. Optional and unset by default:
   * consumers fall back to the hub account avatar, which is what Flow's profile
   * screen already does.
   */
  photo_url: string;

  /**
   * A short self-description shown on a profile screen.
   *
   * Deliberately separate from `notes`. `notes` is HR-admin commentary about a
   * person and must never be rendered as an "About" section; a dead write path
   * in the HR app mapped bio onto notes, and that mapping was wrong. See
   * tryrtylr/rtylr#875.
   */
  bio: string;
}

model UpdateStaffRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateStaffBody;
}

model UpdateStaffResponse {
  success: boolean;
}

model DeleteStaffRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteStaffResponse {
  success: boolean;
}

model ClockInRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  staff_id?: int64;
}

model ClockInResponse {
  success: boolean;
}

model ClockOutRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  staff_id?: int64;
}

model ClockOutResponse {
  success: boolean;
}

/**
 * DirectoryStaffEntry is the deliberately minimal projection of Staff shown
 * in the company-wide directory. It never carries salary, hourly_rate,
 * overtime_rate, address, or emergency_contact_* - those stay HR-admin-only
 * regardless of caller, enforced here at the contract level rather than
 * relying solely on runtime field-zeroing.
 */
model DirectoryStaffEntry {
  @encode(string)
  id: int64;

  first_name: string;
  last_name: string;

  @encode(string)
  department_id: int64;

  role: string;
  employment_type: string;

  /** The current lifecycle state. */
  status: string;

  email: string;
  phone: string;
}

model ListStaffDirectoryRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("page")
  @encode(string)
  page?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @query("page_size")
  @encode(string)
  page_size?: int64;
}

model ListStaffDirectoryResponse {
  staff: DirectoryStaffEntry[];

  @encode(string)
  total: int64;

  @encode(string)
  page: int64;

  @encode(string)
  page_size: int64;
}
