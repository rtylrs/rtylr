import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.work_schedule.v1;

/** WorkSchedulePayload is a recurring staff work schedule. */
model WorkSchedulePayload {
  /** Unique identifier for the work schedule. */
  @encode(string)
  id: int64;

  /** Tenant owner this schedule belongs to. */
  @encode(string)
  owner_id: int64;

  /** Display name of the schedule (e.g. "Standard 9-5"). */
  name: string;

  /** Days of the week this schedule applies to (e.g. ["monday", "tuesday"]). */
  working_days: string[];

  /** Shift start time in HH:MM 24-hour format. */
  start_time: string;

  /** Shift end time in HH:MM 24-hour format. */
  end_time: string;

  /** Duration of the daily break in minutes. */
  break_minutes: int32;

  /** Total contracted hours per week (e.g. 40.0). */
  contracted_hours_per_week: float64;

  /** Daily overtime threshold in minutes; hours beyond this count as OT. */
  daily_ot_threshold_minutes: int32;

  /** Weekly overtime threshold in minutes; hours beyond this count as OT. */
  weekly_ot_threshold_minutes: int32;

  /** Unix timestamp when this schedule was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this schedule was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListWorkSchedulesRequest {
  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** offset is the number of records to skip, for paging through the list. */
  @query("offset")
  offset?: int32;
}

/**
 * ListWorkSchedulesResponse lists work schedules for the tenant.
 *
 * Read by Flow to size a member's workload bar against the hours they are
 * actually contracted for, resolved through Employee.work_schedule_id. Before
 * this had an RPC, Flow hardcoded 40 hours for everyone, which reads a
 * part-timer as under-loaded and a four-day week as a shortfall.
 */
model ListWorkSchedulesResponse {
  /** All work schedules for the tenant. */
  schedules: WorkSchedulePayload[];

  /** total is the total number of schedules for the tenant, for pagination math. */
  total: int32;
}

model CreateWorkScheduleRequest {
  /** Display name of the schedule. */
  @Onkt.required
  name: string;

  /** Days of the week this schedule applies to. */
  @minItems(1)
  working_days: string[];

  /** Shift start time in HH:MM 24-hour format. */
  @Onkt.required
  start_time: string;

  /** Shift end time in HH:MM 24-hour format. */
  @Onkt.required
  end_time: string;

  /** Duration of the daily break in minutes. */
  @Onkt.required
  @minValue(0)
  break_minutes?: int32;

  /** Total contracted hours per week. */
  @minValue(0)
  contracted_hours_per_week: float64;

  /** Daily overtime threshold in minutes. */
  @Onkt.required
  @minValue(0)
  daily_ot_threshold_minutes?: int32;

  /** Weekly overtime threshold in minutes. */
  @Onkt.required
  @minValue(0)
  weekly_ot_threshold_minutes?: int32;
}

model CreateWorkScheduleResponse {
  /** The newly created work schedule. */
  schedule: WorkSchedulePayload;
}

model UpdateWorkScheduleBody {
  /** New name; absent to keep current. */
  name?: string;

  /** New working days; if provided, replaces all existing days. */
  @minItems(1)
  working_days: string[];

  /** New start time; absent to keep current. */
  start_time?: string;

  /** New end time; absent to keep current. */
  end_time?: string;

  /** New break duration in minutes; absent to keep current. */
  break_minutes?: int32;

  /** New contracted hours per week; absent to keep current. */
  contracted_hours_per_week?: float64;

  /** New daily OT threshold; absent to keep current. */
  daily_ot_threshold_minutes?: int32;

  /** New weekly OT threshold; absent to keep current. */
  weekly_ot_threshold_minutes?: int32;
}

model UpdateWorkScheduleRequest {
  /** ID of the schedule to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable work schedule fields bound from the HTTP payload. */
  @body
  body: UpdateWorkScheduleBody;
}

model UpdateWorkScheduleResponse {
  /** The updated work schedule. */
  schedule: WorkSchedulePayload;
}
