import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.overtime.v1;

/** OvertimeRequestPayload is a single employee's request to work overtime. */
model OvertimeRequestPayload {
  /** Unique identifier for the overtime request. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;

  /** Tenant owner this request belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee requesting overtime. */
  @encode(string)
  employee_id: int64;

  /** Date the overtime is planned for (ISO-8601 date). */
  date: string;

  /** Number of extra minutes planned beyond the regular shift. */
  planned_extra_minutes: int32;

  /** Justification for the overtime work. */
  reason: string;

  /** Workflow status (pending | approved | rejected | cancelled). */
  status: string;

  /** Current step in a multi-level approval chain. */
  current_approval_level: int32;

  /** Employee ID of the approver; absent if not yet approved. */
  @encode(string)
  approved_by?: int64;

  /** Pay multiplier applied to the overtime hours (e.g. 1.5); absent if not yet calculated. */
  multiplier_applied?: float64;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListOvertimeRequestsRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by status; absent to include all statuses. */
  @query("status")
  status?: string;
}

model ListOvertimeRequestsResponse {
  /** Page of overtime requests matching the filters. */
  requests: OvertimeRequestPayload[];

  /** Total number of matching overtime requests across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateOvertimeRequestRequest {
  /** Employee requesting overtime. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** Date the overtime is planned for (ISO-8601 date). */
  @Onkt.required
  date: string;

  /** Number of extra minutes planned beyond the regular shift. */
  @Onkt.required
  @minValueExclusive(0)
  planned_extra_minutes?: int32;

  /** Justification for the overtime work. */
  @Onkt.required
  reason: string;
}

model CreateOvertimeRequestResponse {
  /** The newly created overtime request. */
  request: OvertimeRequestPayload;
}

/** ApproveOvertimeBody carries approval decision fields bound from the HTTP payload. */
model ApproveOvertimeBody {
  /** True to approve, false to reject. */
  approved: boolean;

  /** Pay multiplier to apply (e.g. 1.5); absent to use the default rate. */
  multiplier_applied?: float64;

  /** Reason for rejection; absent if approving. */
  rejection_reason?: string;
}

model ApproveOvertimeRequest {
  /** ID of the overtime request being reviewed. */
  @minValueExclusive(0)
  @encode(string)
  overtime_request_id: int64;

  @body
  body: ApproveOvertimeBody;
}

model ApproveOvertimeResponse {
  /** Whether the approval action was processed. */
  success: boolean;
}
