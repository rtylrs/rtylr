import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.shift.v1;

/**
 * ShiftPayload is a row in the shift table, which holds two kinds of record.
 * A template row is a reusable shift for a location: it carries location_id,
 * name and color, and has no staff_id. An attendance row records one staff
 * member's worked shift: it carries staff_id and date plus the clock and hours
 * fields. The two are told apart by whether staff_id is present.
 */
model ShiftPayload {
  /** Unique identifier for the shift row. */
  @encode(string)
  id: int64;

  /** Tenant owner this shift belongs to. */
  @encode(string)
  owner_id: int64;

  /**
   * Location where this shift takes place. Zero on attendance rows, which have
   * no location; kept non-optional so existing readers see no type change.
   */
  @encode(string)
  location_id: int64;

  /**
   * Human-readable shift name (e.g. "Morning", "Night"). Empty on attendance
   * rows; kept non-optional so existing readers see no type change.
   */
  name: string;

  /** Shift start time (HH:MM 24-hour format). */
  start_time: string;

  /** Shift end time (HH:MM 24-hour format). */
  end_time: string;

  /** Hex colour code used to render this shift on the roster (e.g. "#4CAF50"). */
  color: string;

  /** Staff member whose worked shift this is; absent on template rows. */
  @encode(string)
  staff_id?: int64;

  /** Department the staff member worked under; absent when not tracked. */
  @encode(string)
  department_id?: int64;

  /** Calendar date this attendance row covers (ISO-8601); absent on template rows. */
  date?: string;

  /** Unpaid break length in minutes, deducted from worked hours. */
  break_duration_minutes?: int32;

  /** Lifecycle state (scheduled | in_progress | completed | absent | cancelled). */
  status?: string;

  /** Unix timestamp when the staff member clocked in; absent until clock-in. */
  @encode(string)
  clock_in_at?: int64;

  /** Unix timestamp when the staff member clocked out; absent until clock-out. */
  @encode(string)
  clock_out_at?: int64;

  /** Hours worked on this shift, computed at clock-out. */
  hours_worked?: float64;

  /** Hours worked beyond the scheduled window. */
  overtime_hours?: float64;

  /** Free-text note attached to this shift. */
  notes?: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** ShiftAssignmentPayload assigns an employee to a shift on a given date. */
model ShiftAssignmentPayload {
  /** Unique identifier for the assignment. */
  @encode(string)
  id: int64;

  /** Tenant owner this assignment belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee assigned to the shift. */
  @encode(string)
  employee_id: int64;

  /** Shift template being assigned. */
  @encode(string)
  shift_id: int64;

  /** Date the assignment is for (ISO-8601 date). */
  date: string;

  /** Assignment status (scheduled | worked | missed | swapped). */
  status: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListShiftsRequest {
  /** Filter by location; absent to include all locations. */
  @query("location_id")
  @encode(string)
  location_id?: int64;

  /** Filter to one staff member's attendance rows; absent to include all. */
  @query("staff_id")
  @encode(string)
  staff_id?: int64;

  /** Filter attendance rows to a single date (ISO-8601); absent to include all. */
  @query("date")
  date?: string;

  /** Restrict to one record kind (template | attendance); absent returns both. */
  @query("kind")
  kind?: string;

  /** Page size; absent applies the server default. */
  @query("limit")
  limit?: int32;

  /** Offset for pagination; absent starts at the first page. */
  @query("offset")
  offset?: int32;
}

model ListShiftsResponse {
  /** Shift templates matching the filter. */
  shifts: ShiftPayload[];

  /** Total number of matching shifts across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

/**
 * CreateShiftRequest creates either a template row or an attendance row.
 * Supply location_id and name for a template; supply staff_id and date for an
 * attendance row. location_id and name are no longer unconditionally required
 * because attendance rows have neither.
 */
model CreateShiftRequest {
  /** Location where this shift takes place; required for a template row. */
  @encode(string)
  location_id?: int64;

  /** Human-readable shift name; required for a template row. */
  name?: string;

  /** Shift start time (HH:MM 24-hour format). */
  @Onkt.required
  start_time: string;

  /** Shift end time (HH:MM 24-hour format). */
  @Onkt.required
  end_time: string;

  /** Hex colour code for roster display; defaults server-side when absent. */
  color?: string;

  /** Staff member this attendance row belongs to; required for an attendance row. */
  @encode(string)
  staff_id?: int64;

  /** Department the staff member worked under. */
  @encode(string)
  department_id?: int64;

  /** Calendar date of the attendance row (ISO-8601); required for an attendance row. */
  date?: string;

  /** Unpaid break length in minutes. */
  break_duration_minutes?: int32;

  /** Lifecycle state; defaults to "scheduled" when absent. */
  status?: string;

  /** Free-text note attached to this shift. */
  notes?: string;
}

model CreateShiftResponse {
  /** The newly created shift. */
  shift: ShiftPayload;
}

/**
 * UpdateShiftRequest edits an existing shift row of either kind. Only the
 * fields present are applied; absent fields keep their stored value.
 */
model UpdateShiftRequest {
  /** Identifier of the shift row to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** New location for a template row. */
  @encode(string)
  location_id?: int64;

  /** New shift name for a template row. */
  name?: string;

  /** New start time (HH:MM 24-hour format). */
  start_time?: string;

  /** New end time (HH:MM 24-hour format). */
  end_time?: string;

  /** New hex colour code for roster display. */
  color?: string;

  /** New department for an attendance row. */
  @encode(string)
  department_id?: int64;

  /** New calendar date for an attendance row (ISO-8601). */
  date?: string;

  /** New unpaid break length in minutes. */
  break_duration_minutes?: int32;

  /** New lifecycle state. */
  status?: string;

  /** New free-text note. */
  notes?: string;
}

model UpdateShiftResponse {
  /** The shift as stored after the update. */
  shift: ShiftPayload;
}

model DeleteShiftRequest {
  /** Identifier of the shift row to delete. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteShiftResponse {
  /** Whether the shift was deleted. */
  success: boolean;
}

model CreateShiftAssignmentRequest {
  /** Employee to assign. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** Shift template to assign. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  shift_id?: int64;

  /** Date of the assignment (ISO-8601 date). */
  @Onkt.required
  date: string;
}

model CreateShiftAssignmentResponse {
  /** The newly created assignment. */
  assignment: ShiftAssignmentPayload;

  /**
   * Human-readable availability warnings raised against the assignee for this
   * date (e.g. "Jane Doe is unavailable 09:00-12:00 on 2026-08-20"), read off
   * the AvailabilityPayload rows in this package.
   *
   * Advisory only. The assignment is still created, because a scheduler
   * legitimately needs to staff someone against a stated preference: cover for
   * a no-show, a swap the pair agreed offline, a rush the roster did not
   * anticipate. A hard block would make those ordinary cases impossible
   * without first editing the availability record, so the check informs rather
   * than refuses. Empty when no conflict was found.
   *
   * Mirrors GetRosterResponse.conflicts, which already reports this class of
   * problem the same non-blocking way.
   */
  warnings: string[];
}

model GetRosterRequest {
  /** Location to fetch the roster for. */
  @minValueExclusive(0)
  @encode(string)
  location_id: int64;

  /** Monday of the week to retrieve (ISO-8601 date). */
  @Onkt.required
  @query("week_start")
  week_start: string;
}

model GetRosterResponse {
  /** All shift assignments for the requested week and location. */
  assignments: ShiftAssignmentPayload[];

  /** Human-readable conflict descriptions (e.g. "John Smith: double-booked on 2026-04-15"). */
  conflicts: string[];
}

model PublishRosterRequest {
  /**
   * Location to publish the roster for. Optional, and currently not used to
   * scope the publish: a publish covers every location's assignments for the
   * week, so requiring it rejected callers over a field that had no effect.
   */
  @encode(string)
  location_id?: int64;

  /** Monday of the week being published (ISO-8601 date). */
  @Onkt.required
  week_start: string;
}

model PublishRosterResponse {
  /** Whether the roster was successfully published. */
  success: boolean;
}

/**
 * AvailabilityPayload is one window in which a staff member can, or explicitly
 * cannot, be scheduled for a shift. A row is either a recurring weekly pattern
 * (day_of_week set, specific_date absent) or a single-date override
 * (specific_date set, day_of_week absent); an override takes precedence over the
 * recurring pattern for the date it names.
 *
 * Exactly one of day_of_week and specific_date is set on every stored row. .onk
 * has no cross-field decorator for that constraint (the same gap documented on
 * flow.task.v1.ListTasksRequest), so it is enforced in the onion handler rather
 * than in this schema.
 */
model AvailabilityPayload {
  /** Unique identifier for the availability entry. */
  @encode(string)
  id: int64;

  /**
   * Staff member whose availability this is. Matches ShiftPayload.staff_id, not
   * the older ShiftAssignmentPayload.employee_id in this same package.
   */
  @encode(string)
  staff_id: int64;

  /** Tenant owner this entry belongs to. */
  @encode(string)
  owner_id: int64;

  /**
   * Weekday of a recurring pattern: 0 = Monday through 6 = Sunday. Absent on a
   * single-date override. Numbered Monday-first so the value indexes the same
   * week the roster already uses (GetRosterRequest.week_start is a Monday).
   * This is deliberately not JavaScript's Date.getDay(), where 0 = Sunday.
   */
  day_of_week?: int32;

  /**
   * Calendar date a single-date override applies to (ISO-8601, e.g.
   * "2026-08-17"). Absent on a recurring pattern.
   */
  specific_date?: string;

  /** Start of the window (HH:MM 24-hour format). */
  start_time: string;

  /**
   * End of the window (HH:MM 24-hour format). An end_time earlier than
   * start_time means the window crosses midnight and ends the following day,
   * which is how availability for a night shift is expressed.
   */
  end_time: string;

  /**
   * Whether the staff member can be scheduled in this window. False marks a
   * deliberate block, which is how a single-date override withdraws
   * availability that the recurring pattern would otherwise grant.
   * Wire note: the generated Go server tags every scalar `omitempty`, so a
   * false value is dropped from the response body. Read an absent key as false.
   */
  is_available: boolean;

  /** Unix timestamp when this entry was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this entry was last updated. */
  @encode(string)
  updated_at: int64;
}

/**
 * ListAvailabilityRequest filters availability entries. Recurring patterns carry
 * no date, so date_from and date_to constrain single-date overrides only and
 * leave recurring rows in the result; narrow those with day_of_week.
 */
model ListAvailabilityRequest {
  /**
   * Filter to one staff member; absent returns entries for every staff member
   * in the tenant, which is what the weekly roster grid loads in one call.
   */
  @query("staff_id")
  @encode(string)
  staff_id?: int64;

  /**
   * Filter recurring patterns to one weekday, 0 = Monday through 6 = Sunday;
   * absent includes every weekday.
   */
  @query("day_of_week")
  day_of_week?: int32;

  /**
   * Inclusive start of a date range matched against specific_date (ISO-8601);
   * absent leaves the range open at the start.
   */
  @query("date_from")
  date_from?: string;

  /**
   * Inclusive end of a date range matched against specific_date (ISO-8601);
   * absent leaves the range open at the end.
   */
  @query("date_to")
  date_to?: string;

  /** Page size; absent applies the server default. */
  @query("limit")
  limit?: int32;

  /** Offset for pagination; absent starts at the first page. */
  @query("offset")
  offset?: int32;
}

model ListAvailabilityResponse {
  /**
   * Availability entries matching the filters, recurring patterns and
   * single-date overrides alike.
   */
  entries: AvailabilityPayload[];

  /** Total number of matching entries across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

/**
 * SetAvailabilityRequest creates or updates one availability window. With id
 * present the identified entry is replaced. With id absent the write upserts on
 * owner_id, staff_id, start_time and whichever of day_of_week or specific_date
 * is set, so re-submitting an unchanged window is idempotent instead of
 * producing a duplicate. start_time belongs to that key so one staff member can
 * hold several windows on the same day, for example a split shift of 09:00 to
 * 12:00 and 14:00 to 18:00.
 *
 * Supply exactly one of day_of_week and specific_date. As on
 * AvailabilityPayload, no decorator can express that, so the handler rejects a
 * request carrying both or neither.
 */
model SetAvailabilityRequest {
  /** Entry to replace; absent creates the entry or upserts it by natural key. */
  @encode(string)
  id?: int64;

  /** Staff member this availability is for. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  staff_id?: int64;

  /** Weekday for a recurring pattern: 0 = Monday through 6 = Sunday. */
  @minValue(0)
  @maxValue(6)
  day_of_week?: int32;

  /** Calendar date for a single-date override (ISO-8601). */
  specific_date?: string;

  /** Start of the window (HH:MM 24-hour format). */
  @Onkt.required
  start_time: string;

  /**
   * End of the window (HH:MM 24-hour format). Earlier than start_time means the
   * window crosses midnight into the following day.
   */
  @Onkt.required
  end_time: string;

  /**
   * Whether the staff member can be scheduled in this window. Optional so that
   * an explicit false arrives as a set value rather than as a decoded zero:
   * absent means true, the ordinary case of declaring availability, and false
   * marks a deliberate block. Without that distinction, omitting the field
   * would silently block the window.
   */
  is_available?: boolean;
}

model SetAvailabilityResponse {
  /** The availability entry as stored after the write. */
  entry: AvailabilityPayload;
}

model DeleteAvailabilityRequest {
  /** Identifier of the availability entry to delete. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteAvailabilityResponse {
  /** Whether the availability entry was deleted. */
  success: boolean;
}
