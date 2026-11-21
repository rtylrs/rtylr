import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.audit.v1;

/** AuditLog is a recorded POS audit event. */
model AuditLog {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  branch_id: int64;

  @encode(string)
  terminal_id?: int64;

  @encode(string)
  employee_id: int64;

  action: string;
  entity_type: string;

  @encode(string)
  entity_id?: int64;

  details?: string;
  ip_address?: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model CreateAuditRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  terminal_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  @Onkt.required
  action: string;

  @Onkt.required
  entity_type: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  entity_id?: int64;

  @Onkt.required
  details: string;

  @Onkt.required
  ip_address: string;
}

model CreateAuditResponse {
  @encode(string)
  audit_log_id: int64;

  message: string;
}

model ListAuditRequest {
  @query("action")
  action: string;

  @query("entity_type")
  entity_type: string;

  @query("employee_id")
  @encode(string)
  employee_id: int64;

  @Onkt.required
  @query("start_date")
  start_date: string;

  @Onkt.required
  @query("end_date")
  end_date: string;

  /** Maximum records to return. Defaults to 50 and is capped by the server. */
  @query("limit")
  limit: int32;

  /** Opaque continuation token from a previous response. */
  @query("cursor")
  cursor: string;
}

model ListAuditResponse {
  logs: AuditLog[];
  next_cursor: string;
}
