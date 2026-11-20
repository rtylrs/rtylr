import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.equipment.v1;

/** EquipmentRequestPayload is a request for equipment made on behalf of an employee. */
model EquipmentRequestPayload {
  /** Unique identifier for the equipment request. */
  @encode(string)
  id: int64;

  /** Tenant owner this request belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee who needs the equipment. */
  @encode(string)
  employee_id: int64;

  /** Category of equipment being requested (e.g. "laptop", "monitor", "desk"). */
  equipment_type: string;

  /** Free-text description of the specific equipment needed. */
  description: string;

  /** Priority level (low | medium | high). */
  urgency: string;

  /** Date by which the equipment is needed (ISO-8601 date); absent if no deadline. */
  needed_by_date?: string;

  /** Workflow status (pending | approved | ordered | delivered | rejected). */
  status: string;

  /** Current step in a multi-level approval chain. */
  current_approval_level: int32;

  /** Employee who approved the request; absent if not yet approved. */
  @encode(string)
  approved_by?: int64;

  /** Unix timestamp when the equipment was delivered; absent if not yet delivered. */
  @encode(string)
  delivered_at?: int64;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** AssetPayload is a physical asset assigned to an employee. */
model AssetPayload {
  /** Unique identifier for the asset. */
  @encode(string)
  id: int64;

  /** Tenant owner this asset belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee the asset is currently assigned to. */
  @encode(string)
  employee_id: int64;

  /** Human-readable name of the asset (e.g. "MacBook Pro 16-inch"). */
  name: string;

  /** Manufacturer serial number; absent if not tracked. */
  serial_number?: string;

  /** Date the asset was assigned to the employee (ISO-8601 date). */
  assigned_date: string;

  /** Date the asset was returned; absent if still assigned (ISO-8601 date). */
  returned_date?: string;

  /** Physical condition of the asset (new | good | fair | damaged). */
  condition: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListAssetsRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by condition; absent to include all conditions. */
  @query("condition")
  condition?: string;
}

model ListAssetsResponse {
  /** Page of assets matching the filters. */
  assets: AssetPayload[];

  /** Total number of matching assets across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model ListEquipmentRequestsRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by status; absent to include all statuses. */
  @query("status")
  status?: string;
}

model ListEquipmentRequestsResponse {
  /** Page of equipment requests matching the filters. */
  requests: EquipmentRequestPayload[];

  /** Total number of matching equipment requests across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateEquipmentRequestRequest {
  /** Employee who needs the equipment. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** Category of equipment being requested. */
  @Onkt.required
  equipment_type: string;

  /** Free-text description of the specific equipment needed. */
  @Onkt.required
  description: string;

  /** Priority level (low | medium | high). */
  @Onkt.required
  urgency: string;

  /** Date by which the equipment is needed (ISO-8601 date); absent if no deadline. */
  needed_by_date?: string;
}

model CreateEquipmentRequestResponse {
  /** The newly created equipment request. */
  request: EquipmentRequestPayload;
}

/** ApproveEquipmentBody carries approval decision fields bound from the HTTP payload. */
model ApproveEquipmentBody {
  /** True to approve, false to reject. */
  approved: boolean;
}

model ApproveEquipmentRequest {
  /** ID of the equipment request being reviewed. */
  @minValueExclusive(0)
  @encode(string)
  equipment_request_id: int64;

  @body
  body: ApproveEquipmentBody;
}

model ApproveEquipmentResponse {
  /** Whether the approval action was processed. */
  success: boolean;
}
