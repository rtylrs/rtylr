import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.onboarding.v1;

/** ChecklistItemPayload is a single item within a checklist template. */
model ChecklistItemPayload {
  /** Unique identifier for the checklist item. */
  @encode(string)
  id: int64;

  /** Tenant owner this item belongs to. */
  @encode(string)
  owner_id: int64;

  /** Template this item belongs to. */
  @encode(string)
  template_id: int64;

  /** Short title describing the task. */
  title: string;

  /** Detailed description of what needs to be done. */
  description: string;

  /** Role responsible for completing this item (e.g. "hr_manager", "it_admin", "employee"). */
  assignee_role: string;

  /** Number of days after trigger date by which this item is due. */
  due_offset_days: int32;

  /** Whether the assignee must upload an attachment to complete this item. */
  requires_attachment: boolean;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** ChecklistTemplatePayload is a reusable onboarding/offboarding checklist definition. */
model ChecklistTemplatePayload {
  /** Unique identifier for the template. */
  @encode(string)
  id: int64;

  /** Tenant owner this template belongs to. */
  @encode(string)
  owner_id: int64;

  /** Human-readable name for the template (e.g. "Engineering Onboarding"). */
  name: string;

  /** Event that triggers this checklist (onboarding | offboarding). */
  trigger_type: string;

  /** Role this template applies to; absent if it applies to all roles. */
  @encode(string)
  role_id?: int64;

  /** Items included in this template. */
  items: ChecklistItemPayload[];

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** EmployeeChecklistItemPayload is one instantiated checklist item assigned to an employee. */
model EmployeeChecklistItemPayload {
  /** Unique identifier for the employee checklist item. */
  @encode(string)
  id: int64;

  /** Tenant owner this item belongs to. */
  @encode(string)
  owner_id: int64;

  /** Parent employee checklist this item belongs to. */
  @encode(string)
  employee_checklist_id: int64;

  /** Original template item this was created from. */
  @encode(string)
  checklist_item_id: int64;

  /** Employee assigned to complete this item. */
  @encode(string)
  assigned_to_employee_id: int64;

  /** Date by which this item should be completed (ISO-8601 date). */
  due_date: string;

  /** Completion status (incomplete | in_progress | complete). */
  status: string;

  /** Unix timestamp when the item was completed; absent if not yet complete. */
  @encode(string)
  completed_at?: int64;

  /** URL of the uploaded attachment; absent if none provided. */
  attachment_url?: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** EmployeeChecklistPayload is a checklist instantiated for a specific employee. */
model EmployeeChecklistPayload {
  /** Unique identifier for the employee checklist. */
  @encode(string)
  id: int64;

  /** Tenant owner this checklist belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee this checklist was triggered for. */
  @encode(string)
  employee_id: int64;

  /** Template this checklist was created from. */
  @encode(string)
  template_id: int64;

  /** Unix timestamp when this checklist was triggered. */
  @encode(string)
  triggered_at: int64;

  /** Individual items in this employee's checklist. */
  items: EmployeeChecklistItemPayload[];

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListChecklistTemplatesRequest {
  /** Filter by trigger type; absent to include all types. */
  @query("trigger_type")
  trigger_type?: string;
}

model ListChecklistTemplatesResponse {
  /** Page of checklist templates matching the filters. */
  templates: ChecklistTemplatePayload[];

  /** Total number of matching templates across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateChecklistTemplateRequest {
  /** Human-readable name for the template. */
  @Onkt.required
  name: string;

  /** trigger_type is the event that triggers this checklist (onboarding | offboarding). */
  trigger_type?: string;

  /** Role this template applies to; absent if it applies to all roles. */
  @encode(string)
  role_id?: int64;
}

model CreateChecklistTemplateResponse {
  /** The newly created checklist template. */
  template: ChecklistTemplatePayload;
}

/** UpdateChecklistTemplateBody carries updatable checklist template fields bound from the HTTP payload. */
model UpdateChecklistTemplateBody {
  /** name is the human-readable name for the template. */
  @Onkt.required
  name: string;

  /** trigger_type is the event that triggers this checklist (onboarding | offboarding). */
  trigger_type?: string;
}

model UpdateChecklistTemplateRequest {
  /** role_id is the checklist template role identifier from the path. */
  @minValueExclusive(0)
  @encode(string)
  role_id: int64;

  @body
  body: UpdateChecklistTemplateBody;
}

model UpdateChecklistTemplateResponse {
  /** success indicates the checklist template was updated. */
  success: boolean;
}

model AddItemToChecklistTemplateRequest {
  /** role_id is the checklist template role identifier from the path. */
  @minValueExclusive(0)
  @encode(string)
  role_id: int64;

  /** title is the checklist item title. */
  @Onkt.required
  title: string;

  /** description is the optional supporting text for the checklist item. */
  description?: string;

  /** assignee_role is the role responsible for completing this item. */
  assignee_role?: string;

  /** due_offset_days is the number of days after trigger date by which this item is due. */
  due_offset_days?: int32;

  /** requires_attachment indicates whether completion needs an uploaded attachment. */
  requires_attachment?: boolean;
}

model AddItemToChecklistTemplateResponse {
  /** item is the checklist template item that was created. */
  item: ChecklistItemPayload;
}

model ListEmployeeChecklistsRequest {
  /** employee_id is the employee whose checklists should be listed. */
  @Onkt.required
  @minValueExclusive(0)
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** limit is the maximum number of checklists to return. */
  @query("limit")
  limit?: int32;

  /** offset is the number of checklists to skip before returning results. */
  @query("offset")
  offset?: int32;
}

model ListEmployeeChecklistsResponse {
  /** Page of employee checklists matching the filters. */
  checklists: EmployeeChecklistPayload[];

  /** Total number of matching checklists across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model GetEmployeeChecklistRequest {
  /** employee_id is the employee whose checklist should be fetched. */
  @minValueExclusive(0)
  @encode(string)
  employee_id: int64;
}

model GetEmployeeChecklistResponse {
  /** checklist is the employee checklist returned for the path id. */
  checklist: EmployeeChecklistPayload;
}

model TriggerChecklistRequest {
  /** Employee to trigger the checklist for. */
  @minValueExclusive(0)
  @encode(string)
  employee_id: int64;

  /** template_id identifies the checklist template to trigger. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  template_id?: int64;

  /** trigger_date is the date the checklist starts, formatted as YYYY-MM-DD. */
  @Onkt.required
  @pattern("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
  trigger_date: string;
}

model TriggerChecklistResponse {
  /** The employee checklist that was created. */
  checklist: EmployeeChecklistPayload;
}

/**
 * CompleteChecklistItemRequest has no fields in the source proto - the item
 * being completed is not identified anywhere on this message (no path
 * param, no body field). Carried over unchanged and flagged for manual
 * follow-up; this looks like an incomplete contract upstream.
 */
model CompleteChecklistItemRequest {}

model CompleteChecklistItemResponse {
  /** success indicates the checklist item was marked complete. */
  success: boolean;
}
