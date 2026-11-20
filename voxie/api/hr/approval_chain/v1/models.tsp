import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.approval_chain.v1;

/** ApprovalChainLevelPayload is a single level within an approval chain pipeline. */
model ApprovalChainLevelPayload {
  /** Unique identifier for this approval level. */
  @encode(string)
  id: int64;

  /** Tenant owner this level belongs to. */
  @encode(string)
  owner_id: int64;

  /** Parent approval chain this level is part of. */
  @encode(string)
  chain_id: int64;

  /** Position in the approval sequence (1 = first approver). */
  level_order: int32;

  /** Type of approver (direct_manager | dept_head | hr_admin | role | specific_person). */
  approver_type: string;

  /** Role that can approve at this level; absent unless approver_type is role. */
  @encode(string)
  approver_role_id?: int64;

  /** Specific employee who approves; absent unless approver_type is specific_person. */
  @encode(string)
  approver_employee_id?: int64;

  /** Minimum value that triggers this level; absent if no lower bound. */
  threshold_min_value?: float64;

  /** Maximum value that triggers this level; absent if no upper bound. */
  threshold_max_value?: float64;

  /** Days before the level times out and the timeout_action fires. */
  timeout_days: int32;

  /** What happens on timeout (escalate | auto_approve | do_nothing). */
  timeout_action: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** ApprovalChainPayload is the ordered approval pipeline for a request type. */
model ApprovalChainPayload {
  /** Unique identifier for the approval chain. */
  @encode(string)
  id: int64;

  /** Tenant owner this chain belongs to. */
  @encode(string)
  owner_id: int64;

  /** Type of request this chain governs (timesheet | leave | overtime | equipment | payroll). */
  request_type: string;

  /** Human-readable name for the chain. */
  name: string;

  /** Whether approvers can delegate their approval authority. */
  delegation_enabled: boolean;

  /** Ordered levels that make up the approval pipeline. */
  levels: ApprovalChainLevelPayload[];

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model UpsertApprovalChainRequest {
  /** Chain ID for update; absent (zero) to create a new chain. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;

  /** Type of request this chain governs (timesheet | leave | overtime | equipment | payroll). */
  @Onkt.required
  request_type: string;

  /** Human-readable name for the chain. */
  @Onkt.required
  name: string;

  /** Whether approvers can delegate their approval authority. */
  delegation_enabled: boolean;

  /** Ordered levels to set on the chain; replaces all existing levels. */
  @minItems(1)
  levels: ApprovalChainLevelPayload[];
}

model UpsertApprovalChainResponse {
  /** The created or updated approval chain. */
  chain: ApprovalChainPayload;
}
