import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.time.v1;

/**
 * WorkType is one node of the two-level work classification a workspace admin
 * maintains, for example Development with a Code review child. Entries point at
 * either level; hr_category_id maps the node onto HR's payroll category so one
 * entry serves both project reporting and payroll.
 */
model WorkType {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /**
   * parent_id is the parent node, or null for a top-level type. The tree is
   * exactly two levels deep.
   */
  @encode(string)
  parent_id?: int64;

  /** name is the display name of this work type. */
  name: string;

  /**
   * description is a business-authored explanation of what this work type
   * covers, shown to employees as a read-only "Task Description" once they
   * pick the type in Log Hours - distinct from the free-text description an
   * employee types for what they actually did.
   */
  description: string;

  /** hr_category_id maps this node to an HR payroll category. */
  @encode(string)
  hr_category_id?: int64;

  /** sort_order controls the order the tree is presented in. */
  sort_order: int32;

  /** archived hides the node from pickers without breaking historic entries. */
  archived: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last changed. */
  updated_at: string;
}

/**
 * WorkDay is one person's working day: the mode, the clock, and the state of its
 * approval. The individual pieces of work are its lines.
 */
model WorkDay {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** employee_id is the person the day belongs to. */
  @encode(string)
  employee_id: int64;

  /**
   * user_id is the employee's auth user id, populated when the caller already
   * resolved it (the team view, which joins against workspace membership).
   * Callers reading their own day already know their own identity and get
   * this unset.
   */
  @encode(string)
  user_id?: int64;

  /** date is the calendar day in the person's own timezone, as YYYY-MM-DD. */
  date: string;

  /**
   * timezone is the IANA zone the day was recorded in. Stored per record because
   * a travelling worker's day boundary is their own, not the workspace's.
   */
  timezone: string;

  /**
   * work_mode is one of onsite, remote or hybrid. It is the day's default and a
   * line may override it.
   */
  work_mode: string;

  /**
   * start_time is the local start of the day as HH:MM, or null for a day that
   * records only its lines.
   */
  start_time?: string;

  /** end_time is the local end of the day as HH:MM. */
  end_time?: string;

  /** break_minutes is unpaid break time, subtracted from the clocked total. */
  break_minutes: int32;

  /** status is one of draft, submitted, approved, rejected or reopened. */
  status: string;

  /** clocked_minutes is end minus start minus breaks, or zero when unclocked. */
  clocked_minutes: int32;

  /** logged_minutes is the sum of this day's lines. */
  logged_minutes: int32;

  /**
   * gap_minutes is clocked minus logged. It is reported, never enforced: a
   * blocked submission produces invented lines rather than honest ones.
   */
  gap_minutes: int32;

  /**
   * logged_by is the user who recorded the day, which differs from employee_id
   * when a manager fills it in on someone's behalf.
   */
  @encode(string)
  logged_by?: int64;

  /** approved_by is the approver, once the day has been approved. */
  @encode(string)
  approved_by?: int64;

  /** approved_at is when it was approved. */
  approved_at?: string;

  /** reopened_by is the manager who reopened an approved day. */
  @encode(string)
  reopened_by?: int64;

  /** reopened_at is when it was reopened. */
  reopened_at?: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last changed. */
  updated_at: string;
}

/**
 * WorkDayLine is one piece of work inside a day. It is persisted as an HR time
 * entry, so payroll and Flow read the same rows.
 */
model WorkDayLine {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** work_day_id is the day this line belongs to. */
  @encode(string)
  work_day_id: int64;

  /**
   * project_id is the project the work was for. A line may have a project and no
   * task, for work that is not tracked as one.
   */
  @encode(string)
  project_id?: int64;

  /** task_id is the task the work was for. */
  @encode(string)
  task_id?: int64;

  /** work_type_id is the classification node, at either level of the tree. */
  @encode(string)
  work_type_id?: int64;

  /** work_mode overrides the day's mode for this line only. Null means inherit. */
  work_mode?: string;

  /** description is what was done. */
  description: string;

  /** duration_minutes is the time spent. */
  duration_minutes: int32;

  /** is_billable marks the line as chargeable, defaulted from the work type. */
  is_billable: boolean;

  /** status mirrors the day's approval state for this line. */
  status: string;

  /**
   * client_key is generated by the client before the line is sent. It makes an
   * offline retry idempotent, so a field worker syncing twice does not log twice.
   */
  client_key: string;

  /** logged_by is who recorded the line. */
  @encode(string)
  logged_by?: int64;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last changed. */
  updated_at: string;
}

model GetWorkDayRequest {
  /** date is the day to read, as YYYY-MM-DD. */
  @Onkt.required
  @query("date")
  date: string;

  /**
   * employee_id reads another person's day, which requires manage team time.
   * Omitted, it reads the caller's own.
   */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;
}

model GetWorkDayResponse {
  /** work_day is null when nothing has been recorded for that date yet. */
  work_day?: WorkDay;

  /** lines are the day's pieces of work. */
  lines: WorkDayLine[];
}

model UpsertWorkDayRequest {
  /** date is the day being recorded, as YYYY-MM-DD. */
  @Onkt.required
  date: string;

  /** timezone is the IANA zone of the person recording it. */
  @Onkt.required
  timezone: string;

  /** work_mode is one of onsite, remote or hybrid. */
  @Onkt.required
  work_mode: string;

  /** start_time is the local start as HH:MM. */
  start_time?: string;

  /** end_time is the local end as HH:MM. */
  end_time?: string;

  /** break_minutes is unpaid break time. */
  break_minutes: int32;

  /** employee_id records the day on someone else's behalf. */
  @encode(string)
  employee_id?: int64;
}

model UpsertWorkDayResponse {
  work_day: WorkDay;
}

model ListWorkDaysRequest {
  /** from_date is the inclusive start of the range, as YYYY-MM-DD. */
  @Onkt.required
  @query("from")
  from_date: string;

  /** to_date is the inclusive end of the range, as YYYY-MM-DD. */
  @Onkt.required
  @query("to")
  to_date: string;

  /** employee_id lists another person's days. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;
}

model ListWorkDaysResponse {
  work_days: WorkDay[];
  pagination: PaginationResponse;
}

model ListWorkDaysForWorkspaceRequest {
  /**
   * workspace_id is whose members' days this lists. The caller must hold at
   * least admin on this workspace.
   */
  @Onkt.required
  @minValueExclusive(0)
  @query("workspace_id")
  @encode(string)
  workspace_id?: int64;

  /** from_date is the inclusive start of the range, as YYYY-MM-DD. */
  @Onkt.required
  @query("from")
  from_date: string;

  /** to_date is the inclusive end of the range, as YYYY-MM-DD. */
  @Onkt.required
  @query("to")
  to_date: string;

  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;
}

model ListWorkDaysForWorkspaceResponse {
  work_days: WorkDay[];
  pagination: PaginationResponse;
}

model CreateWorkDayLineRequest {
  /** work_day_id is the day to add the line to. */
  @minValueExclusive(0)
  @encode(string)
  work_day_id: int64;

  /** client_key makes the create idempotent across an offline retry. */
  @Onkt.required
  client_key: string;

  /** duration_minutes is the time spent. */
  @Onkt.required
  @minValueExclusive(0)
  duration_minutes?: int32;

  /** description is what was done. */
  @Onkt.required
  description: string;

  @encode(string)
  project_id?: int64;

  @encode(string)
  task_id?: int64;

  @encode(string)
  work_type_id?: int64;

  /** work_mode overrides the day's mode for this line only. */
  work_mode?: string;

  is_billable: boolean;
}

model CreateWorkDayLineResponse {
  line: WorkDayLine;
}

model UpdateWorkDayLineRequest {
  /** id is the line to change. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  @minValueExclusive(0)
  duration_minutes?: int32;

  @Onkt.required
  description: string;

  @encode(string)
  project_id?: int64;

  @encode(string)
  task_id?: int64;

  @encode(string)
  work_type_id?: int64;

  work_mode?: string;
  is_billable: boolean;
}

model UpdateWorkDayLineResponse {
  line: WorkDayLine;
}

model DeleteWorkDayLineRequest {
  /** id is the line to remove. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteWorkDayLineResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ListWorkTypesRequest {
  /** include_archived returns archived nodes too, for editing the tree. */
  @query("include_archived")
  include_archived: boolean;
}

model ListWorkTypesResponse {
  work_types: WorkType[];
}

model CreateWorkTypeRequest {
  /** name is the display name. */
  @Onkt.required
  name: string;

  /** description is shown to employees as the task description once picked. */
  description: string;

  /** parent_id nests this node under a top-level type. */
  @encode(string)
  parent_id?: int64;

  /** hr_category_id maps the node onto an HR payroll category. */
  @encode(string)
  hr_category_id?: int64;

  sort_order: int32;
}

model CreateWorkTypeResponse {
  work_type: WorkType;
}

model UpdateWorkTypeRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  name: string;

  /** description is shown to employees as the task description once picked. */
  description: string;

  @encode(string)
  parent_id?: int64;

  @encode(string)
  hr_category_id?: int64;

  sort_order: int32;
  archived: boolean;
}

model UpdateWorkTypeResponse {
  work_type: WorkType;
}

model DeleteWorkTypeRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteWorkTypeResponse {
  /** message is a human-readable result message. */
  message: string;
}
