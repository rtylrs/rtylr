import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.time_entry.v1;

/** TimeEntryPayload is a logged unit of employee work time. */
model TimeEntryPayload {
  /** Unique identifier for the time entry. */
  @encode(string)
  id: int64;

  /** Tenant owner this entry belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee who logged the time. */
  @encode(string)
  employee_id: int64;

  /** Associated project; absent if not project-based. */
  @encode(string)
  project_id?: int64;

  /** Associated task; absent if not task-level tracking. */
  @encode(string)
  task_id?: int64;

  /** Time category for reporting; absent if uncategorised. */
  @encode(string)
  category_id?: int64;

  /** Date the work was performed (ISO-8601 format, e.g. "2026-04-14"). */
  date: string;

  /** Total duration of work in minutes. */
  duration_minutes: int32;

  /** Clock-in time (ISO-8601 time or datetime); absent if duration-only entry. */
  start_time?: string;

  /** Clock-out time (ISO-8601 time or datetime); absent if duration-only entry. */
  end_time?: string;

  /** Free-text description of the work performed. */
  description: string;

  /** Whether this time is billable to a client. */
  is_billable: boolean;

  /** Workflow status (draft | submitted | approved | rejected). */
  status: string;

  /** Reason the entry was rejected; absent unless status is rejected. */
  rejection_reason?: string;

  /** Employee ID of the approver; absent if not yet approved. */
  @encode(string)
  approved_by?: int64;

  /** How the entry was created (manual | timer | flow_import). */
  source: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** TimesheetPayload is an employee's time entries grouped over a period. */
model TimesheetPayload {
  /** Unique identifier for the timesheet. */
  @encode(string)
  id: int64;

  /** Tenant owner this timesheet belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee this timesheet is for. */
  @encode(string)
  employee_id: int64;

  /** Start of the timesheet period (ISO-8601 date). */
  period_start: string;

  /** End of the timesheet period (ISO-8601 date). */
  period_end: string;

  /** Workflow status (draft | submitted | approved | rejected | locked). */
  status: string;

  /** Current step in a multi-level approval chain. */
  current_approval_level: int32;

  /** Unix timestamp when the timesheet was submitted; absent if still in draft. */
  @encode(string)
  submitted_at?: int64;

  /** Sum of all entry durations in minutes. */
  total_minutes: int32;

  /** Individual time entries within this timesheet. */
  entries: TimeEntryPayload[];

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/**
 * ListTimeEntriesRequest selects logged time. Every filter is optional and
 * they combine with AND. The response is always a page: callers that need
 * everything must follow offset until total is reached.
 */
model ListTimeEntriesRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /**
   * Filter by exact date; absent to include all dates. Ignored when from_date
   * or to_date is supplied.
   */
  @query("date")
  date?: string;

  /**
   * Include entries dated on or after this day (ISO-8601, e.g. "2026-04-01");
   * absent to leave the range open at the start.
   */
  @query("from_date")
  from_date?: string;

  /**
   * Include entries dated on or before this day; absent to leave the range
   * open at the end.
   */
  @query("to_date")
  to_date?: string;

  /**
   * Filter by the Flow project the time was logged against; absent to include
   * all projects and entries with no project.
   */
  @query("project_id")
  @encode(string)
  project_id?: int64;

  /**
   * Filter by billability; absent to include both billable and non-billable
   * time.
   */
  @query("is_billable")
  is_billable?: boolean;

  /**
   * Filter by workflow status (draft | submitted | approved | rejected);
   * absent to include all statuses.
   */
  @query("status")
  status?: string;

  /** Maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** Number of records to skip for pagination. */
  @query("offset")
  offset: int32;
}

model ListTimeEntriesResponse {
  /** Time entries matching the filters. */
  entries: TimeEntryPayload[];

  /** Total number of matching time entries across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateTimeEntryRequest {
  /** Employee who logged the time. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** Associated project; absent if not project-based. */
  @encode(string)
  project_id?: int64;

  /** Associated task; absent if not task-level tracking. */
  @encode(string)
  task_id?: int64;

  /** Time category for reporting; absent if uncategorised. */
  @encode(string)
  category_id?: int64;

  /** Date the work was performed (ISO-8601 format). */
  @Onkt.required
  date: string;

  /** Total duration of work in minutes. */
  @Onkt.required
  @minValueExclusive(0)
  duration_minutes?: int32;

  /** Clock-in time; absent if duration-only entry. */
  start_time?: string;

  /** Clock-out time; absent if duration-only entry. */
  end_time?: string;

  /** Free-text description of the work performed. */
  @Onkt.required
  description: string;

  /** Whether this time is billable to a client. */
  is_billable: boolean;

  /** How the entry was created (manual | timer | flow_import). */
  @Onkt.required
  source: string;
}

model CreateTimeEntryResponse {
  /** The newly created time entry. */
  entry: TimeEntryPayload;
}

model ListTimesheetsRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by status; absent to include all statuses. */
  @query("status")
  status?: string;
}

model ListTimesheetsResponse {
  /** Page of timesheets matching the filters. */
  timesheets: TimesheetPayload[];

  /** Total number of matching timesheets across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model SubmitTimesheetRequest {
  /** ID of the timesheet to submit for approval. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  timesheet_id?: int64;
}

model SubmitTimesheetResponse {
  /** Whether the submission was accepted. */
  success: boolean;
}

model ApproveTimesheetBody {
  /** True to approve, false to reject. */
  approved: boolean;

  /** Reason for rejection; absent if approving. */
  rejection_reason?: string;
}

model ApproveTimesheetRequest {
  /** ID of the timesheet being reviewed. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries approval decision fields bound from the HTTP payload. */
  @body
  body: ApproveTimesheetBody;
}

model ApproveTimesheetResponse {
  /** Whether the approval action was processed. */
  success: boolean;
}

/**
 * MarkTimeEntriesInvoicedRequest records that the listed approved billable
 * entries are settled by one finance invoice. Idempotent per entry: an entry
 * already carrying a billable_invoice_id is skipped, never overwritten.
 */
model MarkTimeEntriesInvoicedRequest {
  /** ID kept for route symmetry with approveTimesheet; unused by the handler. */
  @bodyIgnore
  @encode(string)
  id: int64;

  /** body carries the batch fields bound from the HTTP payload. */
  @body
  body: MarkTimeEntriesInvoicedBody;
}

/** MarkTimeEntriesInvoicedBody is the payload of markTimeEntriesInvoiced. */
model MarkTimeEntriesInvoicedBody {
  /**
   * entry_ids are the time entries this invoice bills. Every entry must
   * belong to the caller's business; foreign or missing ids are counted as
   * skipped rather than failing the batch.
   */
  @minItems(1)
  @Onkt.encode("string")
  entry_ids: int64[];

  /** invoice_id is the finance invoice that settles them. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  invoice_id?: int64;
}

/**
 * MarkTimeEntriesInvoicedResponse reports how the batch landed, so a caller
 * can tell "all marked" from "some were already invoiced" without a second
 * read.
 */
model MarkTimeEntriesInvoicedResponse {
  /** marked_count is how many entries now carry this invoice id. */
  marked_count: int32;

  /**
   * skipped_count is how many of entry_ids were already invoiced, not
   * approved billable work, or belong to another business.
   */
  skipped_count: int32;
}
