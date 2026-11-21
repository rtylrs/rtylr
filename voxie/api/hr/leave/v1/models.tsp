import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.leave.v1;

/** LeaveTypePayload defines a category of leave employees can request. */
model LeaveTypePayload {
  /** Unique identifier for the leave type. */
  @encode(string)
  id: int64;

  /** Tenant owner this leave type belongs to. */
  @encode(string)
  owner_id: int64;

  /** Human-readable name (e.g. "Annual Leave", "Sick Leave"). */
  name: string;

  /** Whether days taken under this type are paid. */
  is_paid: boolean;

  /** Whether leave requests of this type require manager approval. */
  requires_approval: boolean;

  /** Rule describing how entitlement accrues (e.g. "monthly", "yearly"); absent if entitlement is fixed. */
  accrual_rule?: string;

  /** Days granted per accrual period (per year for "annual", per month for "monthly"); absent when none. */
  accrual_amount?: float64;

  /** Maximum unused days that may carry over to the next year. */
  max_carryover_days: int32;

  /** Minimum calendar days of notice required before the leave starts. */
  min_notice_days: int32;

  /** Number of consecutive days after which supporting documentation is required; absent if never required. */
  requires_documentation_after_days?: int32;

  /** Whether this leave type is currently available for new requests. */
  is_active: boolean;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** LeaveRequestPayload is a single employee's request for time off. */
model LeaveRequestPayload {
  /** Unique identifier for the leave request. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** Tenant owner this request belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee who is requesting leave. */
  @encode(string)
  employee_id: int64;

  /** Leave type being requested. */
  @encode(string)
  leave_type_id: int64;

  /** First day of leave (ISO-8601 date). */
  start_date: string;

  /** Last day of leave (ISO-8601 date). */
  end_date: string;

  /** Total number of working days requested (may be fractional for half-days). */
  total_days: float64;

  /** Additional context from the employee; absent if none provided. */
  notes?: string;

  /** Workflow status (pending | approved | rejected | cancelled). */
  status: string;

  /** Current step in a multi-level approval chain. */
  current_approval_level: int32;

  /** Reason the request was rejected; absent unless status is rejected. */
  rejection_reason?: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** LeaveBalancePayload tracks an employee's remaining entitlement for a leave type and year. */
model LeaveBalancePayload {
  /** Unique identifier for the balance record. */
  @encode(string)
  id: int64;

  /** Tenant owner this balance belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee this balance is for. */
  @encode(string)
  employee_id: int64;

  /** Leave type this balance tracks. */
  @encode(string)
  leave_type_id: int64;

  /** Calendar year this balance applies to. */
  year: int32;

  /** Total entitled days for the year (including carryover). */
  entitled_days: float64;

  /** Days already consumed by approved requests. */
  used_days: float64;

  /** Days reserved by pending (not yet approved) requests. */
  pending_days: float64;

  /** Days still available (entitled - used - pending). */
  remaining_days: float64;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListLeaveRequestsRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by status; absent to include all statuses. */
  @query("status")
  status?: string;

  /**
   * Include requests overlapping this day or later (ISO-8601 date); absent
   * to leave the range open at the start. A request overlaps when it starts
   * on or before to_date and ends on or after from_date.
   */
  @query("from_date")
  from_date?: string;

  /**
   * Include requests overlapping this day or earlier (ISO-8601 date); absent
   * to leave the range open at the end.
   */
  @query("to_date")
  to_date?: string;

  /** Maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** Number of records to skip for pagination. */
  @query("offset")
  offset: int32;
}

model ListLeaveRequestsResponse {
  /** Page of leave requests matching the filters. */
  requests: LeaveRequestPayload[];

  /** Total number of matching leave requests across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateLeaveRequestRequest {
  /** Employee requesting leave. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** Leave type being requested. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  leave_type_id?: int64;

  /** First day of leave (ISO-8601 date). */
  @Onkt.required
  start_date: string;

  /** Last day of leave (ISO-8601 date). */
  @Onkt.required
  end_date: string;

  /** Total number of working days requested. */
  @Onkt.required
  @minValueExclusive(0)
  total_days?: float64;

  /** Additional context from the employee; absent if none provided. */
  notes?: string;
}

model CreateLeaveRequestResponse {
  /** The newly created leave request. */
  request: LeaveRequestPayload;
}

/** ApproveLeaveBody carries approval decision fields bound from the HTTP payload. */
model ApproveLeaveBody {
  /** True to approve, false to reject. */
  approved: boolean;

  /** Reason for rejection; absent if approving. */
  rejection_reason?: string;
}

model ApproveLeaveRequest {
  /** request_id is the leave request being reviewed. */
  @minValueExclusive(0)
  @encode(string)
  request_id: int64;

  @body
  body: ApproveLeaveBody;
}

model ApproveLeaveResponse {
  /** Whether the approval action was processed. */
  success: boolean;
}

model CancelLeaveRequestRequest {
  /**
   * request_id is the leave request being withdrawn. Only the request's own
   * subject may cancel it, and only while it is still pending - the leave
   * approval permission grants no override here.
   */
  @minValueExclusive(0)
  @encode(string)
  request_id: int64;
}

model CancelLeaveResponse {
  /** Whether the request was still pending and is now canceled. */
  success: boolean;
}

model ListLeaveTypesRequest {}

model ListLeaveTypesResponse {
  /** All leave types configured for the tenant. */
  leave_types: LeaveTypePayload[];

  /** Total number of leave types. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateLeaveTypeRequest {
  /** Human-readable name (e.g. "Annual Leave", "Sick Leave"). */
  @Onkt.required
  name: string;

  /** Whether days taken under this type are paid. */
  is_paid: boolean;

  /** Whether leave requests of this type require manager approval. */
  requires_approval: boolean;

  /** Accrual rule; absent if entitlement is fixed. */
  accrual_rule?: string;

  /** Days granted per accrual period; absent or zero grants no entitlement. */
  @minValue(0)
  accrual_amount?: float64;

  /** Maximum unused days that may carry over to the next year. */
  @Onkt.required
  @minValueExclusive(0)
  max_carryover_days?: int32;

  /** Minimum calendar days of notice required. */
  @Onkt.required
  @minValueExclusive(0)
  min_notice_days?: int32;

  /** Consecutive days after which documentation is required; absent if never. */
  requires_documentation_after_days?: int32;
}

model CreateLeaveTypeResponse {
  /** The newly created leave type. */
  leave_type: LeaveTypePayload;
}

/** UpdateLeaveTypeBody carries updatable leave type fields bound from the HTTP payload. */
model UpdateLeaveTypeBody {
  /** Updated name; absent to keep current value. */
  name?: string;

  /** Updated paid status; absent to keep current value. */
  is_paid?: boolean;

  /** Updated approval requirement; absent to keep current value. */
  requires_approval?: boolean;

  /** Updated accrual rule; absent to keep current value. */
  accrual_rule?: string;

  /** Updated days granted per accrual period; absent to keep current value. */
  @minValue(0)
  accrual_amount?: float64;

  /** Updated max carryover days; absent to keep current value. */
  max_carryover_days?: int32;

  /** Updated min notice days; absent to keep current value. */
  min_notice_days?: int32;

  /** Updated documentation threshold; absent to keep current value. */
  requires_documentation_after_days?: int32;

  /** Updated active status; absent to keep current value. */
  is_active?: boolean;
}

model UpdateLeaveTypeRequest {
  /** ID of the leave type to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateLeaveTypeBody;
}

model UpdateLeaveTypeResponse {
  /** The updated leave type. */
  leave_type: LeaveTypePayload;
}

model ListLeaveBalancesRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by year; absent to return current year. */
  @query("year")
  year?: int32;
}

model ListLeaveBalancesResponse {
  /** Leave balances matching the filters. */
  balances: LeaveBalancePayload[];

  /** Total number of matching balances across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}
